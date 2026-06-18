# R1 — Volume: grafo estático em escala

## Objetivo

~20.000 usuários sintéticos, grau médio ~300 (50–3000), **sem** atividade social no steady state.

## Backend / worker

- [ ] `SIMULATOR_MODE=graph_only` no realm volume
- [ ] `graph_bootstrap.py`: usuários mínimos + configuration model (Chung–Lu)
- [ ] Bulk `COPY` / batches para `connections` (~3M arestas)
- [ ] `SIMULATOR_VOLUME_AGENT_COUNT=20000` (monotônico, sem shrink)
- [ ] Steady desligado ou manutenção mínima
- [ ] Workers graph/batch/ml só no volume

## Frontend

- [ ] Network: ego-graph + aviso "modo laboratório"
- [ ] Badge realm volume no toggle

## Critério de pronto

PageRank roda no grafo; UI não tenta renderizar 20k nós.
