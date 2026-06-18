# R2 — Vivo: reações LinkedIn + threads

## Objetivo

Reações complexas em **posts e comentários**; comentários com `parent_comment_id` (2 níveis).

## Schema (migration `000006`)

- [ ] `content_reactions(target_type, target_id, user_id, kind)`
- [ ] Kinds: `like`, `celebrate`, `support`, `insightful`, `love`, `funny`
- [ ] `comments.parent_comment_id` nullable FK

## API

- [ ] `POST /v1/comments/{id}/reactions`
- [ ] Feed/post/comment: `reaction_summary`, `my_reaction`

## Frontend

- [ ] `ReactionBar` em posts e comentários
- [ ] Breakdown por tipo de reação

## Critério de pronto

Usuário humano reage e vê resumo; comentário pode ser resposta (1 nível de thread).
