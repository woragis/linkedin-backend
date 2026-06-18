# R4 — Vivo: simulador social com LLM

## Objetivo

~50 agentes interagem entre si e com humanos, em horários de rotina, só no realm **live**.

## LLM

```env
LLM_PROVIDER=openai
LLM_MODEL=gpt-5-nano
OPENAI_API_KEY=...
SIMULATOR_LLM=1
SIMULATOR_LLM_MAX_PER_MIN=12
SIMULATOR_LLM_MAX_PER_DAY=800
```

## Worker

- [ ] `content/llm.py` — OpenAI client, PT-BR, fallback template
- [ ] Dual-loop simulator: volume URL + live URL
- [ ] Ações: post, comentário, resposta (thread), reação (heurística)
- [ ] Prioridade para threads com usuário humano
- [ ] `active_hours` + timezone por agente

## Economia de tokens

| Tarefa | LLM |
|--------|-----|
| Post / comentário / resposta | Sim |
| Escolher tipo de reação | Não (matriz persona) |
| Decidir se age | Não (Markov + horário) |
| Persona bootstrap | Sim (1×) |

## Custo estimado

~100–200 chamadas/dia → **&lt; US$ 1/mês** com `gpt-5-nano`.

## Critério de pronto

Feed vivo ganha posts contextualizados; agentes respondem em threads em ondas horárias.
