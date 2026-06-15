# API planejada

Base URL: `http://127.0.0.1:8080`

## Saúde

| Método | Rota | Descrição |
|---|---|---|
| GET | `/health` | Liveness |
| GET | `/ready` | Postgres ping |

## Auth (fase 1)

| Método | Rota | Descrição |
|---|---|---|
| POST | `/v1/auth/register` | Criar conta |
| POST | `/v1/auth/login` | Login |
| POST | `/v1/auth/logout` | Logout |
| GET | `/v1/me` | Usuário autenticado + perfil |

## Perfis (fase 1)

| Método | Rota | Descrição |
|---|---|---|
| PATCH | `/v1/me/profile` | Atualizar perfil |
| GET | `/v1/users/{slug}` | Perfil público |
| PUT | `/v1/me/experiences` | CRUD experiências |
| PUT | `/v1/me/educations` | CRUD formação |
| PUT | `/v1/me/skills` | Atualizar skills |

## Conexões (fase 2)

| Método | Rota | Descrição |
|---|---|---|
| POST | `/v1/connections/request` | Pedir conexão |
| PATCH | `/v1/connections/{id}/accept` | Aceitar |
| PATCH | `/v1/connections/{id}/reject` | Recusar |
| GET | `/v1/connections` | Listar conexões |
| GET | `/v1/connections/pending` | Pedidos pendentes |

## Posts e feed (fase 2)

| Método | Rota | Descrição |
|---|---|---|
| POST | `/v1/posts` | Criar post |
| GET | `/v1/posts/{id}` | Detalhe |
| POST | `/v1/posts/{id}/reactions` | Reagir |
| POST | `/v1/posts/{id}/comments` | Comentar |
| GET | `/v1/feed` | Feed do usuário |

## Eventos (fase 2)

| Método | Rota | Descrição |
|---|---|---|
| POST | `/v1/events` | Batch de eventos do frontend |

```json
{
  "events": [
    { "type": "post_viewed", "payload": { "post_id": "..." }, "at": "2026-06-15T12:00:00Z" }
  ]
}
```

## Busca (fase 3)

| Método | Rota | Descrição |
|---|---|---|
| GET | `/v1/search/people?q=` | Busca perfis |
| GET | `/v1/search/posts?q=` | Busca posts |

## Recomendações (fase 3)

| Método | Rota | Descrição |
|---|---|---|
| GET | `/v1/recommendations/people` | Pessoas sugeridas |
| GET | `/v1/recommendations/jobs` | Vagas (fase futura) |

## Rede e analytics (fase 4+)

| Método | Rota | Descrição |
|---|---|---|
| GET | `/v1/network/graph` | Subgrafo para visualização |
| GET | `/v1/analytics/overview` | Métricas (admin) |
| GET | `/v1/analytics/churn` | Distribuição de churn |

## Interno (dev)

| Método | Rota | Header | Descrição |
|---|---|---|---|
| POST | `/v1/internal/seed-demo` | `X-Internal-Token` | Popular dados demo |

## Erros

JSON padrão:

```json
{ "code": "PROFILE_NOT_FOUND", "message": "Profile not found." }
```
