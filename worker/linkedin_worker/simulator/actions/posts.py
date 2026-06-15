from __future__ import annotations

import random
from uuid import UUID, uuid4

import psycopg

from linkedin_worker import settings
from linkedin_worker.simulator.content.templates import pick_post_body
from linkedin_worker.simulator.db import enqueue_outbox, insert_event


def create_post(
    conn: psycopg.Connection,
    user_id: UUID,
    body: str,
    *,
    enqueue_search: bool | None = None,
) -> UUID:
    post_id = uuid4()
    conn.execute(
        """
        INSERT INTO posts (id, author_id, body)
        VALUES (%s, %s, %s)
        """,
        (post_id, user_id, body),
    )
    insert_event(conn, user_id, "post_created", {"post_id": str(post_id)})
    if settings.SIMULATOR_ENQUEUE_SEARCH if enqueue_search is None else enqueue_search:
        enqueue_outbox(conn, "search.index_post", {"post_id": str(post_id)})
    return post_id


def session_start(conn: psycopg.Connection, user_id: UUID) -> None:
    insert_event(conn, user_id, "session_start", {"source": "simulator"})


def create_bootstrap_post(
    conn: psycopg.Connection,
    user_id: UUID,
    rng: random.Random,
    topic: str,
) -> UUID:
    body = pick_post_body(rng, topic)
    return create_post(conn, user_id, body, enqueue_search=False)
