# Arquitetura

## Visão geral

```mermaid
flowchart TB
    subgraph client [Frontend Next.js]
        UI[UI]
        Tracker[Event tracker]
    end

    subgraph api [server Go]
        HTTP[HTTP API]
        Outbox[Outbox writer]
    end

    subgraph durable [Durável]
        PG[(PostgreSQL)]
    end

    subgraph fast [Rápido]
        RD[(Redis)]
        ES[(Elasticsearch fase 3+)]
    end

    subgraph workers [Workers Python]
        WRT[worker-realtime]
        WBT[worker-batch]
    end

    UI --> HTTP
    Tracker --> HTTP
    HTTP --> PG
    HTTP --> Outbox
    Outbox --> PG
    Outbox --> RD
    RD --> WRT
    WRT --> ES
    WRT --> PG
    WBT --> PG
    WBT --> ES
    HTTP --> RD
```

## Camadas

| Camada | Responsabilidade |
|---|---|
| **API (Go)** | Auth, CRUD, feed, busca, servir scores pré-computados |
| **Postgres OLTP** | Usuários, grafo, posts, eventos, outbox |
| **Postgres analytics** | Agregados, scores, sugestões |
| **Redis** | Fila de jobs, cache de feed (fase 7) |
| **Elasticsearch** | Full-text em perfis e posts (fase 3+) |
| **Workers** | Indexação, graph, ML, churn, rollups |

## Durabilidade de eventos

Fluxo transacional na API:

```sql
BEGIN;
  INSERT INTO events (...);
  INSERT INTO outbox_jobs (job_type, payload, ...);
COMMIT;
```

O `worker-realtime` executa **outbox relay**:

1. Lê `outbox_jobs WHERE processed_at IS NULL`
2. Publica no Redis (`linkedin:jobs`)
3. Marca `processed_at`

Se Redis cair, jobs acumulam no outbox e são reprocessados. **Nenhum evento se perde.**

## Busca + recomendação

Sistemas separados que se encontram na UI:

| Fluxo | Recuperação | Ranking |
|---|---|---|
| Busca | Elasticsearch | Go re-ranqueia com affinity (top 50) |
| Sidebar "Pessoas sugeridas" | Postgres `user_connection_suggestions` | Pré-computado pelo worker batch |

## Deploy de workers

| Container | `WORKER_ROLE` | Jobs |
|---|---|---|
| `worker-realtime` | `realtime` | outbox relay, events, notifications |
| `worker-indexer` | `indexer` | `search.index_profile`, `search.index_post` |
| `worker-graph` | `graph` | cron PageRank; `graph.recompute_user` |
| `worker-ml` | `ml` | cron `ml_training` (scikit-learn) |
| `worker-batch` | `batch` | recommendations, feed-ranking, churn, analytics-rollup |
| `worker-simulator` | `simulator` | população sintética (produção acadêmica) |
| `worker-all` | `all` | tudo (dev local) |

Mesma imagem Docker (`WORKER_DEPS` build-arg), papéis e filas diferentes via env.

### Filas Redis

| Fila | Jobs |
|---|---|
| `linkedin:jobs:realtime` | `analytics.process_event`, `notifications.send` |
| `linkedin:jobs:indexer` | `search.index_*` |
| `linkedin:jobs:graph` | `graph.recompute_user` |

O outbox relay publica na fila correta por `job_type`. Ver [DEPLOY-RAILWAY.md](DEPLOY-RAILWAY.md).

## Kafka (fase 7+)

Postgres continua fonte da verdade. Kafka entra **entre** outbox relay e consumers para replay e pipeline de dados — não substitui o outbox.

```
outbox → relay → Kafka topic → N consumers → warehouse / ML / ES
```
