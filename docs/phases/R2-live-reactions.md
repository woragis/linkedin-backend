# R2 — Vivo: reações LinkedIn + threads

## Objetivo

Reações complexas em **posts e comentários**; comentários com `parent_comment_id` (2 níveis).

## Schema (migration `000006`)

- [x] `content_reactions(target_type, target_id, user_id, kind)`
- [x] Kinds: `like`, `celebrate`, `support`, `insightful`, `love`, `funny`
- [x] `comments.parent_comment_id` nullable FK
- [x] Migração de `reactions` → `content_reactions`

## API

- [x] `POST /v1/comments/{id}/reactions`
- [x] Feed/post/comment: `reaction_summary`, `my_reaction`
- [x] `POST /v1/posts/{id}/comments` aceita `parent_comment_id`
- [x] `GET /v1/posts/{id}/comments` retorna árvore com `replies`

## Frontend

- [x] `ReactionBar` em posts e comentários
- [x] Breakdown por tipo de reação (`ReactionSummaryStrip`)
- [x] Threads: responder (1 nível)

## Critério de pronto

Usuário humano reage e vê resumo; comentário pode ser resposta (1 nível de thread).
