# R0 — Dual realm (fundação)

## Objetivo

Separar **volume** (laboratório de grafo) e **vivo** (experiência social) sem duplicar microserviços.

## Backend

- [x] `X-App-Realm: live|volume` (default `live`)
- [x] `DATABASE_URL` → volume, `DATABASE_URL_LIVE` → vivo
- [x] `MultiApp` roteia handlers por realm
- [x] Feed cache Redis: prefixo `live:` / `volume:`
- [x] CORS: header `X-App-Realm`
- [x] `postgres-live` no docker-compose
- [ ] Seed demo nos dois bancos (`POST /v1/internal/seed-demo` em cada realm)

## Frontend

- [x] `localStorage` realm, default `live`
- [x] Header em todas as chamadas API
- [x] Toggle no `AppShell` + badge
- [x] Troca → logout + redirect `/login`

## Critério de pronto

1. `docker compose up` sobe `postgres` + `postgres-live`
2. Criar conta no vivo; alternar para volume → login separado
3. Dados não vazam entre realms

## Commits esperados

1. `docs: dual-realm plan and phase R0-R4`
2. `feat(api): dual-realm routing with two Postgres pools`
3. `feat(frontend): realm toggle and X-App-Realm header`
