from __future__ import annotations

import random
from uuid import UUID, uuid4

import psycopg

from linkedin_worker.simulator.content.templates import pick_comment_body
from linkedin_worker.simulator.db import insert_event


def comment_on_post(
    conn: psycopg.Connection,
    user_id: UUID,
    post_id: UUID,
    rng: random.Random,
    topic: str,
    *,
    agent=None,
    post_body_text: str = "",
    parent_comment_id: UUID | None = None,
    parent_body: str | None = None,
) -> UUID | None:
    if agent is not None:
        from linkedin_worker.simulator.content.generator import comment_body as gen_comment

        body = gen_comment(
            rng,
            agent,
            topic,
            post_body_text=post_body_text,
            parent_comment=parent_body,
        )
    else:
        body = pick_comment_body(rng, topic)
    comment_id = uuid4()
    row = conn.execute(
        """
        INSERT INTO comments (id, post_id, author_id, body, parent_comment_id)
        VALUES (%s, %s, %s, %s, %s)
        RETURNING id
        """,
        (comment_id, post_id, user_id, body, parent_comment_id),
    ).fetchone()
    if not row:
        return None
    insert_event(
        conn,
        user_id,
        "comment_created",
        {"comment_id": str(comment_id), "post_id": str(post_id)},
    )
    return comment_id
