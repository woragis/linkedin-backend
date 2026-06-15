from __future__ import annotations

import json
import logging
from typing import Any
from uuid import UUID

import psycopg

log = logging.getLogger("linkedin-worker.simulator.db")

PASSWORD_PLAIN = "simulator-internal"


def count_simulator_agents(conn: psycopg.Connection) -> int:
    row = conn.execute("SELECT COUNT(*)::int FROM simulator_agents").fetchone()
    return int(row[0]) if row else 0


def load_existing_slugs(conn: psycopg.Connection) -> set[str]:
    rows = conn.execute("SELECT slug FROM profiles").fetchall()
    return {row[0] for row in rows}


def slug_exists(conn: psycopg.Connection, slug: str) -> bool:
    row = conn.execute("SELECT 1 FROM profiles WHERE slug = %s LIMIT 1", (slug,)).fetchone()
    return row is not None


def ensure_unique_slug_conn(conn: psycopg.Connection, base: str) -> str:
    candidate = base
    for i in range(100):
        if not slug_exists(conn, candidate):
            return candidate
        candidate = f"{base}-{i + 2}"
    raise RuntimeError(f"could not allocate unique slug for {base!r}")


def ensure_catalog_entity(
    conn: psycopg.Connection,
    table: str,
    name: str,
    slug: str,
) -> UUID:
    row = conn.execute(
        f"SELECT id FROM {table} WHERE slug = %s",
        (slug,),
    ).fetchone()
    if row:
        return row[0]
    row = conn.execute(
        f"""
        INSERT INTO {table} (name, slug)
        VALUES (%s, %s)
        ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name
        RETURNING id
        """,
        (name, slug),
    ).fetchone()
    return row[0]


def ensure_skill(conn: psycopg.Connection, name: str, slug: str) -> UUID:
    return ensure_catalog_entity(conn, "skills", name, slug)


def insert_event(
    conn: psycopg.Connection,
    user_id: UUID,
    event_type: str,
    payload: dict[str, Any] | None = None,
) -> None:
    conn.execute(
        """
        INSERT INTO events (user_id, event_type, payload)
        VALUES (%s, %s, %s::jsonb)
        """,
        (user_id, event_type, json.dumps(payload or {})),
    )


def enqueue_outbox(
    conn: psycopg.Connection,
    job_type: str,
    payload: dict[str, Any],
) -> None:
    conn.execute(
        """
        INSERT INTO outbox_jobs (job_type, payload)
        VALUES (%s, %s::jsonb)
        """,
        (job_type, json.dumps(payload)),
    )
