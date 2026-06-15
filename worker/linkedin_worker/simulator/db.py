from __future__ import annotations

import json
import logging
from typing import Any
from uuid import UUID

import psycopg

from linkedin_worker.simulator.agent import Agent

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


def load_agents(conn: psycopg.Connection) -> list[Agent]:
    rows = conn.execute(
        """
        SELECT
            sa.user_id, sa.archetype, sa.age, sa.gender, sa.city,
            sa.latitude, sa.longitude, sa.extraversion, sa.activity_level,
            sa.interests, sa.markov_state, sa.rng_offset,
            p.full_name, p.slug, p.headline, p.location, COALESCE(p.birth_year, 1990)
        FROM simulator_agents sa
        JOIN profiles p ON p.user_id = sa.user_id
        ORDER BY sa.rng_offset
        """
    ).fetchall()
    agents: list[Agent] = []
    for row in rows:
        interests_raw = row[9]
        if isinstance(interests_raw, str):
            interests = json.loads(interests_raw)
        elif isinstance(interests_raw, list):
            interests = interests_raw
        else:
            interests = list(interests_raw or [])
        agents.append(
            Agent(
                user_id=row[0],
                archetype=row[1],
                age=row[2],
                gender=row[3],
                city=row[4],
                latitude=row[5],
                longitude=row[6],
                extraversion=float(row[7]),
                activity_level=float(row[8]),
                interests=interests,
                markov_state=row[10],
                rng_offset=row[11],
                full_name=row[12],
                slug=row[13],
                headline=row[14],
                location=row[15],
                birth_year=row[16],
            )
        )
    return agents


def sample_recent_posts(
    conn: psycopg.Connection,
    viewer_id: UUID,
    *,
    limit: int = 40,
) -> list[tuple[UUID, UUID]]:
    rows = conn.execute(
        """
        SELECT id, author_id
        FROM posts
        WHERE deleted_at IS NULL
          AND author_id <> %s
        ORDER BY created_at DESC
        LIMIT %s
        """,
        (viewer_id, limit),
    ).fetchall()
    return [(row[0], row[1]) for row in rows]


def pending_requester_for_addressee(conn: psycopg.Connection, addressee_id: UUID) -> UUID | None:
    row = conn.execute(
        """
        SELECT requester_id
        FROM connections
        WHERE addressee_id = %s AND status = 'pending'
        ORDER BY created_at ASC
        LIMIT 1
        """,
        (addressee_id,),
    ).fetchone()
    return row[0] if row else None


def update_markov_state(conn: psycopg.Connection, user_id: UUID, state: str) -> None:
    conn.execute(
        """
        UPDATE simulator_agents
        SET markov_state = %s
        WHERE user_id = %s
        """,
        (state, user_id),
    )

