# Plano dual-realm: Volume + Vivo

Dois mundos no mesmo deploy (Opção 1): uma API, dois Postgres, toggle no frontend.

| Realm | Header `X-App-Realm` | Postgres | Propósito |
|-------|----------------------|----------|-----------|
| **Vivo** (default) | `live` | `DATABASE_URL_LIVE` | Rede densa, LLM, social rico |
| **Volume** | `volume` | `DATABASE_URL` | Grafo estático ~20k nós, ML em escala |

## Decisões fechadas

- Grau médio ~300 conexões (50–3000 outliers), grafo **estático**
- Default UI: **vivo**; troca de realm → **re-login**
- Threads em comentários: **2 níveis** (só no vivo)
- Reações em posts e comentários (só no vivo)
- LLM: **OpenAI `gpt-5-nano`**, idioma **PT-BR**
- `SIMULATOR_VOLUME_AGENT_COUNT=20000` fixo (só aumenta via config)

## Fases

| Fase | Doc | Entrega |
|------|-----|---------|
| **R0** | [phases/R0-dual-realm.md](phases/R0-dual-realm.md) | API realm routing, 2 Postgres, toggle UI |
| **R1** | [phases/R1-volume-graph.md](phases/R1-volume-graph.md) | Bootstrap 20k + ~3M arestas, sim só grafo |
| **R2** | [phases/R2-live-reactions.md](phases/R2-live-reactions.md) | `content_reactions`, comentários, API + UI |
| **R3** | [phases/R3-live-profiles.md](phases/R3-live-profiles.md) | Perfil rico, certs, cursos, persona |
| **R4** | [phases/R4-live-llm-simulator.md](phases/R4-live-llm-simulator.md) | Simulador social + `gpt-5-nano` |

## Infra Railway (sem serviços extras)

- `Postgres` → volume (`DATABASE_URL`)
- `PostgresLive` → vivo (`DATABASE_URL_LIVE`)
- Mesmos workers; simulator com duas URLs na Fase R4
- Redis com prefixo `volume:` / `live:` no feed cache

## LLM (Fase R4)

```env
LLM_PROVIDER=openai
LLM_MODEL=gpt-5-nano
OPENAI_API_KEY=sk-...
SIMULATOR_LLM_MAX_PER_MIN=12
```

Reações: heurística por persona (sem LLM). Posts/comentários/respostas: LLM.
