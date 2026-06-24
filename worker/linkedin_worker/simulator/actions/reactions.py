from __future__ import annotations

from uuid import UUID

import psycopg

from linkedin_worker.simulator.db import insert_event


def like_post(conn: psycopg.Connection, user_id: UUID, post_id: UUID) -> bool:
    row = conn.execute(
        """
        INSERT INTO content_reactions (target_type, target_id, user_id, kind)
        VALUES ('post', %s, %s, 'like')
        ON CONFLICT DO NOTHING
        RETURNING target_id
        """,
        (post_id, user_id),
    ).fetchone()
    if not row:
        return False
    conn.execute(
        """
        INSERT INTO reactions (post_id, user_id, kind)
        VALUES (%s, %s, 'like')
        ON CONFLICT DO NOTHING
        """,
        (post_id, user_id),
    )
    insert_event(conn, user_id, "post_liked", {"post_id": str(post_id)})
    return True
