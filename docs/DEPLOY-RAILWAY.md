# Deploy no Railway — LinkedIn acadêmico (API + micro-workers + simulador)

Este projeto é um **laboratório de redes sociais sintéticas**. O simulador roda em **produção** de propósito.

## Serviços Railway

| Serviço | Tipo | Root Directory | `WORKER_ROLE` | Deps image |
|---------|------|----------------|---------------|------------|
| `api` | Web | `server` ou `.` | — | Go Dockerfile |
| `worker-realtime` | Worker | `worker` | `realtime` | `WORKER_DEPS=base` |
| `worker-indexer` | Worker | `worker` | `indexer` | `WORKER_DEPS=base` |
| `worker-graph` | Worker | `worker` | `graph` | `WORKER_DEPS=base` |
| `worker-ml` | Worker | `worker` | `ml` | `WORKER_DEPS=ml` |
| `worker-batch` | Worker | `worker` | `batch` | `WORKER_DEPS=base` |
| `worker-simulator` | Worker | `worker` | `simulator` | `WORKER_DEPS=simulator` |
| `frontend` | Web | `frontend` | — | Next.js |

Plugins: **Postgres**, **Redis**. Elasticsearch opcional (busca degradada sem ES).

## Variáveis obrigatórias (workers)

```env
DATABASE_URL=postgresql://...
REDIS_URL=redis://...
WORKER_ROLE=realtime   # um por serviço
WORKER_HEALTH_ENABLED=1
PORT=8081              # Railway injeta PORT — usado como health port
```

Filas Redis (relay roteia automaticamente):

```env
REDIS_QUEUE_REALTIME=linkedin:jobs:realtime
REDIS_QUEUE_INDEXER=linkedin:jobs:indexer
REDIS_QUEUE_GRAPH=linkedin:jobs:graph
```

## Simulador em produção

```env
WORKER_ROLE=simulator
SIMULATOR_ENABLED=1
SIMULATOR_AGENT_COUNT=2000
SIMULATOR_PHASE=auto
SIMULATOR_ENQUEUE_SEARCH=0
SIMULATOR_METRICS_ENABLED=1
SIMULATOR_METRICS_PORT=9100
```

O tráfego é sintético (`sim-*@sim.local`). Documente na UI que é simulação acadêmica.

## Build Docker (worker)

```bash
# realtime / indexer / graph / batch
docker build --build-arg WORKER_DEPS=base -t linkedin-worker ./worker

# ML (scikit-learn)
docker build --build-arg WORKER_DEPS=ml -t linkedin-worker-ml ./worker

# Simulador
docker build --build-arg WORKER_DEPS=simulator -t linkedin-worker-sim ./worker
```

## Local (stack completa)

```bash
docker compose up -d --build
docker compose --profile prod up -d worker-simulator
```

## O que cada worker faz

| Worker | Fila / Cron | Isola |
|--------|-------------|-------|
| `realtime` | outbox relay → Redis; consome `linkedin:jobs:realtime` | eventos, notificações |
| `indexer` | consome `linkedin:jobs:indexer` | ES indexing |
| `graph` | cron PageRank; consome `linkedin:jobs:graph` | jobs pesados de grafo |
| `ml` | cron `ml_training` | RAM sklearn |
| `batch` | recommendations, feed, churn, rollup | analytics batch |
| `simulator` | loop Markov + psycopg | carga sintética contínua |

## Troubleshooting Railway

1. **Worker morre imediatamente** — criou como Web Service? Use **Worker**.
2. **Build timeout** — use `WORKER_DEPS=base` para serviços sem ML.
3. **Indexer sem ES** — `ELASTICSEARCH_URL` vazio ok; jobs são no-op.
4. **Fila atrasada** — escale só o worker afetado (ex. `worker-indexer`).
