# R1 — Volume: grafo estático em escala

## Objetivo

~1.000 usuários sintéticos (configurável 500–1000), grau médio ~80, **sem** atividade social no steady state.

## Backend / worker

- [x] `SIMULATOR_MODE=graph_only` no realm volume
- [x] `graph_bootstrap.py`: usuários mínimos + configuration model (Chung–Lu)
- [x] Bulk `COPY` para `connections`
- [x] `SIMULATOR_VOLUME_AGENT_COUNT=1000` (monotônico, sem shrink)
- [x] Steady desligado em `graph_only`
- [x] `GET /v1/network/lab-sample` (amostra BFS, max 300 nós)
- [ ] Workers graph/batch/ml só no volume (deploy Railway — manual)

## Frontend

- [x] Network: ego-graph + aviso "modo laboratório"
- [x] `/lab` — grafo interativo com física (cose), zoom, arrastar
- [x] Nav "Lab" no realm volume

## Critério de pronto

PageRank roda no grafo; UI não tenta renderizar 1000 nós (amostra ~150–180).

## Env recomendado (worker-simulator no volume DB)

```env
WORKER_ROLE=simulator
DATABASE_URL=postgres://...@postgres:5432/linkedin
SIMULATOR_MODE=graph_only
SIMULATOR_VOLUME_AGENT_COUNT=1000
SIMULATOR_PHASE=auto
SIMULATOR_GRAPH_MEAN_DEGREE=80
SIMULATOR_ENQUEUE_SEARCH=0
```
