# LinkedIn Profissional — Backend

Rede social profissional com plataforma analítica integrada. Backend em **Go** (API), **Python** (workers de ML/analytics) e **PostgreSQL** como fonte da verdade.

## Estrutura

```
backend/
├── server/           # API HTTP (Go)
├── worker/           # Jobs assíncronos e batch (Python)
├── migrations/       # SQL versionado
├── scripts/          # seed, utilitários
├── docs/             # arquitetura, fases, workers, API
├── docker-compose.yml
└── .env.example
```

## Stack

| Componente | Tecnologia | Papel |
|---|---|---|
| API | Go 1.24 | requests síncronos, auth, CRUD |
| Workers | Python 3.12 | indexação, graph, recomendação, churn |
| Banco OLTP | PostgreSQL 16 | dados operacionais + eventos |
| Fila | Redis 7 | jobs rápidos (outbox relay) |
| Busca | Elasticsearch 8 | fase 2+ (perfis e posts) |

## Quick start

```bash
cp .env.example .env
docker compose up -d --build
curl http://127.0.0.1:8080/health
curl http://127.0.0.1:8080/ready
```

Frontend (repositório separado): `../frontend` — `npm run dev` em `http://127.0.0.1:3000`.

## Documentação

- [Plano de fases](docs/PHASES.md)
- [Arquitetura](docs/ARCHITECTURE.md)
- [Workers](docs/WORKERS.md)
- [Modelo de dados](docs/DATA_MODEL.md)
- [API planejada](docs/API.md)

## Princípios

1. **Eventos nunca vivem só no Redis** — Postgres `events` + `outbox_jobs` é durável.
2. **Workers pesados são offline** — API serve scores pré-computados.
3. **Código modular, deploy flexível** — 1 pacote Python, 2 containers em produção.
