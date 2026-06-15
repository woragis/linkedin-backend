from __future__ import annotations

from uuid import UUID

import psycopg

from linkedin_worker.simulator.db import insert_event


def post_viewed(conn: psycopg.Connection, user_id: UUID, post_id: UUID) -> None:
    insert_event(conn, user_id, "post_viewed", {"post_id": str(post_id)})
