# Plano de implementação por fases

Cada fase entrega valor testável. Commits atômicos por feature dentro da fase.

---

## Fase 0 — Fundação

**Objetivo:** estrutura, docs, API mínima, schema base, workers esqueleto, Docker.

| Entrega | Status |
|---|---|
| Documentação (`docs/`, README) | feito |
| Go server: health, ready, middleware | feito |
| Migrations: core + analytics tables | feito |
| Worker scaffold (realtime + batch) | feito |
| docker-compose (Postgres, Redis, API, workers) | feito |

**Critério de pronto:** `docker compose up` sobe tudo; `/health` e `/ready` OK.

---

## Fase 1 — Identidade e perfil

**Objetivo:** usuário cria conta e monta perfil profissional.

### Backend

- [x] `POST /v1/auth/register`, `POST /v1/auth/login`
- [x] JWT Bearer
- [x] `GET/PATCH /v1/me`, `GET /v1/users/{slug}`
- [x] CRUD de `experiences`, `educations`, `skills`
- [x] Tabelas `institutions`, `companies` com normalização básica
- [x] `POST /v1/internal/seed-demo` (dev)

### Frontend (repo separado)

- [x] Layout LinkedIn-like
- [x] Login / registro
- [x] Edição e visualização de perfil

**Critério de pronto:** criar conta, editar perfil, ver perfil público de outro usuário.

---

## Fase 2 — Rede social mínima

**Objetivo:** conexões, posts, feed cronológico, instrumentação de eventos.

### Backend

- [x] Conexões: pedir, aceitar, recusar, listar
- [x] Posts, reações, comentários
- [x] `GET /v1/feed` (cronológico — posts de conexões)
- [x] `POST /v1/events` (batch de eventos do frontend)
- [x] Outbox: jobs `index_profile`, `index_post`, `recompute_suggestions`
- [x] Worker realtime: relay outbox → Redis

**Critério de pronto:** feed funcional com seed; eventos gravados em `events`.

---

## Fase 3 — Busca e recomendações básicas

**Objetivo:** encontrar pessoas/posts; sugestões de conexão por regras.

### Backend

- [x] Elasticsearch no compose (índices `people`, `posts`)
- [x] Worker `indexer`: profile/post → ES
- [x] `GET /v1/search/people`, `GET /v1/search/posts`
- [x] Re-rank na busca: ES + affinity on-the-fly (top 50)
- [x] Worker batch: `recommendations` (mutual friends + affinity por regras)
- [x] `GET /v1/recommendations/people` (lê `user_connection_suggestions`)

**Critério de pronto:** buscar "react" retorna perfis; sidebar mostra sugestões com `reasons`.

---

## Fase 4 — Graph analytics

**Objetivo:** métricas de rede, visualização, link prediction.

### Backend

- [x] Worker batch: PageRank, centralidade, comunidades
- [x] `user_graph_metrics`, `user_pair_affinity`
- [x] Link prediction (Adamic-Adar + affinity híbrido)
- [x] `GET /v1/network/graph` (dados para visualização)
- [x] `GET /v1/network/influencers`

### Frontend

- [x] Página `/network` (grafo + influenciadores)

**Critério de pronto:** dashboard mostra influência; grafo renderiza subgrafo do usuário.

---

## Fase 5 — Analytics e churn

**Objetivo:** métricas de produto, coortes, predição de abandono.

### Backend

- [x] Worker `analytics_rollup`: DAU, MAU, coortes, engajamento por post
- [x] Worker `churn`: scoring diário → `user_churn_scores`
- [x] Worker `feed_ranking`: scores pré-computados
- [x] Dashboard endpoints: `/v1/analytics/overview`, `top-posts`, `cohorts`, `churn`, `dau`

**Critério de pronto:** métricas batem com eventos do seed; churn scores gerados.

---

## Fase 6 — ML e experimentos A/B

**Objetivo:** modelo aprendido, testes de hipótese no feed.

### Backend

- [x] Worker `ml_training`: regressão logística em aceites de conexão
- [x] Versionamento em `model_versions`
- [x] Framework A/B: variantes de feed (`chronological` vs `ranked`)
- [x] Documentação estatística em `docs/analytics/ab-testing.md`

**Critério de pronto:** modelo treinado; experimento documentado com IC.

---

## Fase 7 — Escala e pipeline de dados (opcional)

**Objetivo:** engenharia de dados, replay, benchmark.

- [x] Kafka/Redpanda entre outbox e consumers (profile `kafka`)
- [x] Export CSV de eventos (`scripts/export_events.py`)
- [x] Cache de feed no Redis
- [x] Benchmark k6 (`scripts/benchmark/k6_feed.js`)
- [x] Observabilidade (métricas Prometheus em `/metrics`)

---

## Ordem de commits sugerida (fases 1+)

Cada item = 1 commit quando possível.

1. `auth: register and login`
2. `profile: me and public slug`
3. `profile: experiences educations skills`
4. `connections: request accept list`
5. `posts: create list reactions comments`
6. `feed: chronological`
7. `events: ingest endpoint and outbox`
8. `worker: outbox relay`
9. `search: elasticsearch setup`
10. `search: people and posts endpoints`
11. `recommendations: affinity scorer`
12. `recommendations: batch suggestions`
13. `graph: metrics batch job`
14. `analytics: rollup and dashboards`
15. `churn: daily scoring`
16. `ml: connection acceptance model`
17. `ab: feed experiment framework`
18. `scale: prometheus redis cache kafka benchmark`

---

## Próximas etapas

Fases 0–7 concluídas. Evolução planejada em [**ROADMAP.md**](ROADMAP.md) (consolidação, simulador, ciência de dados, escala).

Simulador sintético: [**SIMULATOR.md**](SIMULATOR.md).

Dual-realm (volume + vivo): [**REALM_PLAN.md**](REALM_PLAN.md) e fases em [**phases/**](phases/).
