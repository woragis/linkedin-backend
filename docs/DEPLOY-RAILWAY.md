# Deploy no Railway — LinkedIn acadêmico (API + micro-workers + simulador)

Este projeto é um **laboratório de redes sociais sintéticas**. O simulador roda em **produção** de propósito.

## Por que só 2 Dockerfiles aparecem?

Isso é **normal** no desenho original: `server/Dockerfile` (API Go) e `worker/Dockerfile` (imagem única com `WORKER_ROLE` por variável).

O Railway **não cria** um serviço por worker automaticamente. Um Dockerfile ≠ um serviço. Você precisa de **vários serviços** apontando para Dockerfiles diferentes (ou o mesmo, com env distinta).

### Solução no repositório

Adicionamos `deploy/railway/` com **um Dockerfile + `railway.toml` por serviço**:

| Caminho | Serviço |
|---------|---------|
| `server/Dockerfile` | `api` |
| `deploy/railway/worker-realtime/Dockerfile` | `worker-realtime` |
| `deploy/railway/worker-indexer/Dockerfile` | `worker-indexer` |
| `deploy/railway/worker-graph/Dockerfile` | `worker-graph` |
| `deploy/railway/worker-ml/Dockerfile` | `worker-ml` |
| `deploy/railway/worker-batch/Dockerfile` | `worker-batch` |
| `deploy/railway/worker-simulator/Dockerfile` | `worker-simulator` |

**Root Directory** de todos: vazio (raiz do repo `linkedin-backend`).  
O build context precisa incluir `worker/` e `server/`.

## Opção A — Importar compose (mais rápido)

1. Crie projeto vazio no Railway.
2. Adicione **Postgres** e **Redis** (plugins).
3. Arraste `docker-compose.railway.yml` para o canvas do projeto.
4. Para cada `worker-*`: Settings → mude **Web Service** para **Worker**.
5. Em cada serviço, Settings → **Config file path**:
   - `deploy/railway/api/railway.toml`
   - `deploy/railway/worker-realtime/railway.toml`
   - … (um por serviço)
6. Variables: `DATABASE_URL=${{Postgres.DATABASE_URL}}`, `REDIS_URL=${{Redis.REDIS_URL}}`, etc.

## Opção B — Serviço a serviço (manual)

Para cada linha da tabela abaixo:

1. **+ New** → **GitHub Repo** → mesmo repositório.
2. Settings → **Root Directory**: *(vazio)*.
3. Settings → **Dockerfile Path** ou **Config file path** (recomendado).
4. Settings → tipo **Worker** (exceto `api`, que é Web).
5. Variables conforme a seção abaixo.

| Serviço | Tipo | Config file | `WORKER_ROLE` |
|---------|------|-------------|---------------|
| `api` | Web | `deploy/railway/api/railway.toml` | — |
| `worker-realtime` | Worker | `deploy/railway/worker-realtime/railway.toml` | `realtime` |
| `worker-indexer` | Worker | `deploy/railway/worker-indexer/railway.toml` | `indexer` |
| `worker-graph` | Worker | `deploy/railway/worker-graph/railway.toml` | `graph` |
| `worker-ml` | Worker | `deploy/railway/worker-ml/railway.toml` | `ml` |
| `worker-batch` | Worker | `deploy/railway/worker-batch/railway.toml` | `batch` |
| `worker-simulator` | Worker | `deploy/railway/worker-simulator/railway.toml` | `simulator` |

Plugins: **Postgres**, **Redis**. Elasticsearch opcional (busca degradada sem ES).

## Variáveis obrigatórias

### API (`api`)

```env
DATABASE_URL=${{Postgres.DATABASE_URL}}
REDIS_URL=${{Redis.REDIS_URL}}
JWT_SECRET=...
INTERNAL_JOB_TOKEN=...
CORS_ALLOWED_ORIGINS=https://seu-frontend.up.railway.app
```

### Workers (todos)

```env
DATABASE_URL=${{Postgres.DATABASE_URL}}
REDIS_URL=${{Redis.REDIS_URL}}
WORKER_HEALTH_ENABLED=1
# Railway injeta PORT — usado como health port
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
```

O tráfego é sintético (`sim-*@sim.local`). Documente na UI que é simulação acadêmica.

## Local (docker compose)

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

## Troubleshooting

1. **Só vejo 2 Dockerfiles** — use `deploy/railway/*/Dockerfile` ou importe `docker-compose.railway.yml`.
2. **Worker morre imediatamente** — criou como Web Service? Use **Worker**.
3. **Healthcheck failed** — workers têm `healthcheckPath = ""` no `railway.toml`. Se a UI ainda mostra `/health`, apague em Settings → Healthcheck Path.
4. **Build falha com COPY worker/** — Root Directory deve ser a **raiz do repo**, não `worker/`.
5. **Postgres SSL** — use `DATABASE_SSLMODE=prefer` se `require` falhar.
6. **API precisa subir antes** — migrations rodam no serviço `api`; simulador/indexer falham sem tabelas.
7. **Indexer sem ES** — `ELASTICSEARCH_URL` vazio ok; jobs são no-op.
8. **Fila atrasada** — escale só o worker afetado (ex. `worker-indexer`).

## Dev local (um processo só)

`worker/Dockerfile` + `WORKER_ROLE=all` continua válido via `docker compose --profile dev up worker-all`.
