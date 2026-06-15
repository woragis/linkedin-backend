# Roadmap — LinkedIn Profissional + Laboratório de Dados

Documento de referência para implementação incremental. As fases 0–7 do [`PHASES.md`](PHASES.md) estão **concluídas**; este arquivo define o que vem **depois**.

**Última atualização:** 2026-06-15

---

## Estado atual (baseline)

| Área | Status |
|------|--------|
| Backend API (Go) | Completo — auth, social, feed A/B, busca, grafo, analytics |
| Workers (Python) | Completo — graph, recommendations, churn, rollup, ML training |
| Frontend (Next.js) | MVP funcional — feed, conexões, rede, analytics, perfil |
| Testes | Unit + integration + e2e API + pytest |
| Dados | Seed demo com **8 usuários** — insuficiente para ML/analytics robustos |
| ML em produção | Modelo **treina**; sugestões ainda usam **regras** (`rule_based_affinity`) |
| Simulador | **Não implementado** — ver [`SIMULATOR.md`](SIMULATOR.md) |

---

## Curto prazo — consolidar o que tem

**Objetivo:** sistema estável, validado e com loop ML fechado antes de escalar dados.

| # | Entrega | Critério de pronto | Prioridade |
|---|---------|-------------------|------------|
| C1 | **Validação end-to-end manual** | `docker compose up` + seed + workers batch → `/analytics` com DAU, churn, coortes, A/B preenchidos | P0 |
| C2 | **GitHub Actions** | Workflow roda `make test-unit` em todo push; integration/e2e opcionais com Postgres service | P0 |
| C3 | **Inferência ML nas sugestões** | Worker `recommendations` carrega pickle ativo de `model_versions` e combina score regras + ML; UI mostra método `rules+ml` | P0 |
| C4 | **README operacional** | Doc “primeiro run”: tempo esperado dos crons, como forçar batch manualmente | P1 |
| C5 | **Frontend no compose** (opcional) | Serviço `frontend` no docker-compose raiz para demo em um comando | P2 |

### Commits sugeridos

1. `ci: add GitHub Actions test workflow`
2. `feat(recommendations): blend ML model scores in batch job`
3. `docs: operational runbook for first docker compose`

---

## Médio prazo — ciência de dados e simulador

**Objetivo:** gerar volume de dados sintéticos reproduzíveis e pipelines de análise.

| # | Entrega | Critério de pronto | Prioridade |
|---|---------|-------------------|------------|
| M1 | **`worker-simulator` Nível 1** | 500–2k agentes via **psycopg**; arquétipos; ações: post, like, comment, connection; eventos em `events` | P0 |
| M2 | **`worker-simulator` Nível 2** | Cadeia de Markov de estados (offline → browse → engage → post); perfis com atributos demográficos | P0 |
| M3 | **Migration `simulator_agents`** | Tabela de metadados dos agentes (atributos, arquétipo, lat/lon) separada de `users` | P1 |
| M4 | **Export Parquet** | `scripts/export_events.py` → Parquet particionado por dia; script de export de grafo | P1 |
| M5 | **Notebook de validação** | `notebooks/validation.ipynb`: power law (grau), homofilia, coortes simuladas vs esperadas | P1 |
| M6 | **Churn supervisionado** | Substituir heurística em `churn.py` por logistic regression / survival simples treinado em eventos | P2 |
| M7 | **Trigger manual de batch** | `POST /v1/internal/run-batch?job=graph` para dev (token interno) | P2 |

### Dependências

```
M1 (simulator v1) → M2 (Markov) → M5 (validação)
M4 (Parquet) → notebooks / BI externo
C3 (ML inferência) → M6 (churn ML) — padrão reutilizável
```

### Commits sugeridos

1. `feat(simulator): scaffold worker-simulator with psycopg actions`
2. `feat(simulator): archetypes and connection probability model`
3. `feat(simulator): Markov state machine per agent`
4. `feat(data): parquet export and validation notebook`
5. `feat(churn): supervised model replacing heuristic`

---

## Longo prazo — pesquisa, escala e benchmark

**Objetivo:** laboratório de redes complexas + ferramenta de benchmark.

| # | Entrega | Critério de pronto | Prioridade |
|---|---------|-------------------|------------|
| L1 | **Simulador 5k–50k** | Demografia geográfica (cidades BR); distribuições de idade/gênero; homofilia emergente | P0 |
| L2 | **Calibração estatística** | Comparar distribuição de grau, clustering coefficient e diâmetro com literatura (Barabási, etc.) | P1 |
| L3 | **LLM para conteúdo** (opcional) | Posts/comentários gerados por template + LLM com rate limit; flag `SIMULATOR_LLM=0` default | P2 |
| L4 | **Benchmark automatizado** | Pipeline k6 + Prometheus: latência feed, throughput eventos, uso CPU/RAM por N agentes | P1 |
| L5 | **Data lake** | Kafka → consumer → Parquet no MinIO/S3; replay de eventos | P2 |
| L6 | **Paper / TCC** | Documento formal: modelo agent-based, validação, experimentos A/B em escala | — |

---

## O que fica explicitamente fora do escopo (por enquanto)

- Mensagens diretas / chat
- Upload real de avatar (S3)
- Moderação de conteúdo / ML de toxicidade
- Vagas / jobs API
- Multi-tenant / painel admin
- Playwright E2E de browser

---

## Métricas de sucesso do projeto completo

| Métrica | Alvo |
|---------|------|
| Usuários simulados | ≥ 2.000 (médio) · ≥ 10.000 (longo) |
| Eventos/dia | ≥ 50.000 com simulador ativo |
| Workers batch | Todos os crons produzem dados não-vazios |
| A/B feed | Diferença mensurável ou IC documentado em amostra grande |
| ML connection | AUC > 0.65 em holdout com dados simulados |
| p95 feed API | < 200ms com 2k usuários (benchmark) |

---

## Links

- [Plano de fases (0–7)](PHASES.md)
- [Arquitetura](ARCHITECTURE.md)
- [Workers](WORKERS.md)
- [Especificação do simulador](SIMULATOR.md)
- [A/B testing](analytics/ab-testing.md)
