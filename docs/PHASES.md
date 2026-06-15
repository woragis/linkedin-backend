# Plano de implementação por fases

Cada fase entrega valor testável. Commits atômicos por feature dentro da fase.

---

## Fase 0 — Fundação (atual)

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

- [ ] `POST /v1/auth/register`, `POST /v1/auth/login`
- [ ] JWT ou sessão em cookie httpOnly
- [ ] `GET/PATCH /v1/me`, `GET /v1/users/{slug}`
- [ ] CRUD de `experiences`, `educations`, `skills`
- [ ] Tabelas `institutions`, `companies` com normalização básica
- [ ] `POST /v1/internal/seed-demo` (dev)

### Frontend (repo separado)

- [ ] Layout LinkedIn-like
- [ ] Login / registro
- [ ] Edição e visualização de perfil

**Critério de pronto:** criar conta, editar perfil, ver perfil público de outro usuário.

---

## Fase 2 — Rede social mínima

**Objetivo:** conexões, posts, feed cronológico, instrumentação de eventos.

### Backend

- [ ] Conexões: pedir, aceitar, recusar, listar
- [ ] Posts, reações, comentários
- [ ] `GET /v1/feed` (cronológico — posts de conexões)
- [ ] `POST /v1/events` (batch de eventos do frontend)
- [ ] Outbox: jobs `index_profile`, `index_post`, `recompute_suggestions` (stub)
- [ ] Worker realtime: relay outbox → Redis

**Critério de pronto:** feed funcional com seed; eventos gravados em `events`.

---

## Fase 3 — Busca e recomendações básicas

**Objetivo:** encontrar pessoas/posts; sugestões de conexão por regras.

### Backend

- [ ] Elasticsearch no compose (índices `people`, `posts`)
- [ ] Worker `indexer`: profile/post → ES
- [ ] `GET /v1/search/people`, `GET /v1/search/posts`
- [ ] Re-rank na busca: ES + affinity on-the-fly (top 50)
- [ ] Worker batch: `recommendations` (mutual friends + affinity por regras)
- [ ] `GET /v1/recommendations/people` (lê `user_connection_suggestions`)

**Critério de pronto:** buscar "react" retorna perfis; sidebar mostra sugestões com `reasons`.

---

## Fase 4 — Graph analytics

**Objetivo:** métricas de rede, visualização, link prediction.

### Backend

- [ ] Worker batch: PageRank, centralidade, comunidades
- [ ] `user_graph_metrics`, `user_pair_affinity`
- [ ] Link prediction (Adamic-Adar + affinity híbrido)
- [ ] `GET /v1/network/graph` (dados para visualização)
- [ ] `GET /v1/analytics/overview` (DAU, engajamento — admin)

### Frontend

- [ ] Página `/network/graph`

**Critério de pronto:** dashboard mostra influência; grafo renderiza subgrafo do usuário.

---

## Fase 5 — Analytics e churn

**Objetivo:** métricas de produto, coortes, predição de abandono.

### Backend

- [ ] Worker `analytics_rollup`: DAU, MAU, coortes, engajamento por post
- [ ] Worker `churn`: scoring diário → `user_churn_scores`
- [ ] Worker `feed_ranking`: scores pré-computados
- [ ] Dashboard endpoints: retenção, churn, top posts

**Critério de pronto:** métricas batem com eventos do seed; churn scores gerados.

---

## Fase 6 — ML e experimentos A/B

**Objetivo:** modelo aprendido, testes de hipótese no feed.

### Backend

- [ ] Worker `ml_training`: regressão logística em aceites de conexão
- [ ] Versionamento em `model_versions`
- [ ] Framework A/B: variantes de feed, métricas, análise
- [ ] Documentação estatística em `docs/analytics/`

**Critério de pronto:** modelo melhora sugestões vs baseline; experimento documentado com IC.

---

## Fase 7 — Escala e pipeline de dados (opcional)

**Objetivo:** engenharia de dados, replay, benchmark.

- [ ] Kafka/Redpanda entre outbox e consumers
- [ ] Export Parquet / data lake
- [ ] Cache de feed no Redis
- [ ] Benchmark k6 (100k+ usuários simulados)
- [ ] Observabilidade (métricas Prometheus)

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
