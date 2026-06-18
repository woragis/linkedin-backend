# R3 — Vivo: perfis e personalidades ricas

## Objetivo

Dados de perfil detalhados para agentes e usuários no realm **vivo** apenas.

## Schema (migration `000007`)

- [ ] `certifications`, `courses`
- [ ] Campos extras em `profiles`, `educations`, `experiences`
- [ ] `simulator_agents`: `persona_narrative`, `traits_json`, `communication_style`
- [ ] `simulator_agent_memory` (threads, posts lidos)

## Bootstrap live

- [ ] 1× LLM (`gpt-5-nano`) por agente: bio, narrative, formação coerente

## Frontend

- [ ] UI de perfil expandida (vivo)

## Critério de pronto

Perfil público mostra cursos/certs; simulador tem persona textual estável.
