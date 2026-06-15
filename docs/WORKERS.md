# Workers

Um pacote Python (`linkedin_worker/`), múltiplos **papéis lógicos**, **dois containers** em produção.

## Catálogo

| # | Job | Tipo | Container | Gatilho |
|---|---|---|---|---|
| 1 | `indexer` | fila | realtime | `profile_updated`, `post_created` |
| 2 | `events_processor` | fila | realtime | eventos brutos |
| 3 | `notifications` | fila | realtime | `connection_accepted`, etc. |
| 4 | `outbox_relay` | loop | realtime | poll `outbox_jobs` |
| 5 | `graph` | batch | batch | cron 6h |
| 6 | `recommendations` | batch + fila | batch | cron 6h + gatilho |
| 7 | `feed_ranking` | batch | batch | cron 1h |
| 8 | `churn` | batch | batch | cron 1x/dia |
| 9 | `analytics_rollup` | batch | batch | cron 1h / 1x/dia |
| 10 | `ml_training` | batch | batch | cron 1x/semana |

Nenhum job roda em loop infinito "detectando churn". Churn e recomendações são **batch** que gravam scores; a API só lê.

## Filas Redis

| Chave | Payload | Consumidor |
|---|---|---|
| `linkedin:jobs` | `{ "type": "...", "payload": {...} }` | worker-realtime |

### Tipos de job (fila)

| `type` | Handler | Fase |
|---|---|---|
| `search.index_profile` | `jobs/indexer.py` | 3 |
| `search.index_post` | `jobs/indexer.py` | 3 |
| `analytics.process_event` | `jobs/events_processor.py` | 2 |
| `notifications.send` | `jobs/notifications.py` | 4+ |
| `graph.recompute_user` | `jobs/recommendations.py` | 3 |
| `outbox.relay` | interno | 2 |

## Cron batch (`worker-batch`)

| Job | Schedule default | Módulo |
|---|---|---|
| graph | `0 */6 * * *` | `jobs/graph.py` |
| recommendations | `30 */6 * * *` | `jobs/recommendations.py` |
| feed_ranking | `0 * * * *` | `jobs/feed_ranking.py` |
| churn | `0 3 * * *` | `jobs/churn.py` |
| analytics_rollup | `15 * * * *` | `jobs/analytics_rollup.py` |
| ml_training | `0 4 * * 0` | `jobs/ml_training.py` |

## Affinity scorer (recommendations)

Features e pesos iniciais (fase 3):

| Feature | Peso |
|---|---|
| Conexões em comum | 0.35 |
| Mesma instituição | 0.15 |
| Skills em comum (Jaccard) | 0.12 |
| Mesma empresa | 0.10 |
| Mesmo período de formação | 0.08 |
| Mesmo campo de estudo | 0.05 |
| Mesma localização | 0.05 |
| Idade próxima | 0.03 |
| PageRank do sugerido | 0.03 |

Output: `user_connection_suggestions(viewer, suggested, score, reasons JSONB)`.

## Variáveis de ambiente

| Variável | Default | Descrição |
|---|---|---|
| `WORKER_ROLE` | `all` | `realtime`, `batch`, `all` |
| `DATABASE_URL` | — | Postgres |
| `REDIS_URL` | `redis://redis:6379/0` | Fila |
| `REDIS_QUEUE_KEY` | `linkedin:jobs` | Lista de jobs |
| `OUTBOX_POLL_INTERVAL_SEC` | `2` | Relay poll |
| `BATCH_ENABLED` | `1` | Liga scheduler no batch |
| `ELASTICSEARCH_URL` | — | Fase 3+ |
