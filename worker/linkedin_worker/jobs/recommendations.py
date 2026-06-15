"""Connection suggestions and affinity scoring."""

from __future__ import annotations

import json
import logging
from typing import Any

import psycopg

from linkedin_worker.affinity import compute_affinity_score

log = logging.getLogger("linkedin-worker.recommendations")


def recompute_user(conn: psycopg.Connection, payload: dict[str, Any]) -> None:
    user_id = payload.get("user_id")
    if not user_id:
        return
    _score_viewer(conn, user_id)


def run_batch(conn: psycopg.Connection) -> None:
    log.info("recommendations batch started")
    viewers = conn.execute("SELECT user_id FROM profiles").fetchall()
    for (viewer_id,) in viewers:
        _score_viewer(conn, str(viewer_id))
    conn.commit()
    log.info("recommendations batch finished viewers=%d", len(viewers))


def _score_viewer(conn: psycopg.Connection, viewer_id: str) -> None:
    candidates = conn.execute(
        """
        SELECT p.user_id
        FROM profiles p
        WHERE p.user_id <> %s::uuid
          AND NOT EXISTS (
            SELECT 1 FROM connections c
            WHERE c.status = 'accepted'
              AND ((c.requester_id = %s::uuid AND c.addressee_id = p.user_id)
                OR (c.addressee_id = %s::uuid AND c.requester_id = p.user_id))
          )
          AND NOT EXISTS (
            SELECT 1 FROM connections c
            WHERE c.status = 'pending' AND c.requester_id = %s::uuid AND c.addressee_id = p.user_id
          )
        LIMIT 200
        """,
        (viewer_id, viewer_id, viewer_id, viewer_id),
    ).fetchall()

    scored: list[tuple[str, float, list[str]]] = []
    for (target_id,) in candidates:
        score, reasons = _affinity(conn, viewer_id, str(target_id))
        if score > 0.05:
            scored.append((str(target_id), score, reasons))

    scored.sort(key=lambda x: x[1], reverse=True)
    top = scored[:50]

    for target_id, score, reasons in top:
        conn.execute(
            """
            INSERT INTO user_pair_affinity (viewer_id, target_id, score, reasons, computed_at)
            VALUES (%s::uuid, %s::uuid, %s, %s::jsonb, now())
            ON CONFLICT (viewer_id, target_id) DO UPDATE
            SET score = EXCLUDED.score, reasons = EXCLUDED.reasons, computed_at = now()
            """,
            (viewer_id, target_id, score, json.dumps(reasons)),
        )

    conn.execute("DELETE FROM user_connection_suggestions WHERE viewer_id = %s::uuid", (viewer_id,))
    for rank, (target_id, score, reasons) in enumerate(top[:10], start=1):
        conn.execute(
            """
            INSERT INTO user_connection_suggestions
              (viewer_id, suggested_user_id, score, reasons, rank, computed_at)
            VALUES (%s::uuid, %s::uuid, %s, %s::jsonb, %s, now())
            """,
            (viewer_id, target_id, score, json.dumps(reasons), rank),
        )
    conn.commit()


def _affinity(conn: psycopg.Connection, viewer_id: str, target_id: str) -> tuple[float, list[str]]:
    mutual = conn.execute(
        """
        WITH viewer_peers AS (
          SELECT CASE WHEN requester_id = %s::uuid THEN addressee_id ELSE requester_id END AS peer
          FROM connections WHERE status = 'accepted' AND (%s::uuid IN (requester_id, addressee_id))
        ),
        target_peers AS (
          SELECT CASE WHEN requester_id = %s::uuid THEN addressee_id ELSE requester_id END AS peer
          FROM connections WHERE status = 'accepted' AND (%s::uuid IN (requester_id, addressee_id))
        )
        SELECT COUNT(*)::int FROM viewer_peers v JOIN target_peers t ON v.peer = t.peer
        """,
        (viewer_id, viewer_id, target_id, target_id),
    ).fetchone()
    m = int(mutual[0]) if mutual else 0

    school = conn.execute(
        """
        SELECT 1 FROM educations ev
        JOIN educations et ON et.institution_id = ev.institution_id
        WHERE ev.user_id = %s::uuid AND et.user_id = %s::uuid
        LIMIT 1
        """,
        (viewer_id, target_id),
    ).fetchone()

    skills = conn.execute(
        """
        SELECT COUNT(*)::int FROM user_skills vs
        JOIN user_skills ts ON ts.skill_id = vs.skill_id
        WHERE vs.user_id = %s::uuid AND ts.user_id = %s::uuid
        """,
        (viewer_id, target_id),
    ).fetchone()
    sk = int(skills[0]) if skills else 0

    company = conn.execute(
        """
        SELECT 1 FROM experiences ev
        JOIN experiences et ON et.company_id = ev.company_id
        WHERE ev.user_id = %s::uuid AND et.user_id = %s::uuid
        LIMIT 1
        """,
        (viewer_id, target_id),
    ).fetchone()

    cohort = conn.execute(
        """
        SELECT 1 FROM educations ev
        JOIN educations et ON et.user_id = %s::uuid
        WHERE ev.user_id = %s::uuid
          AND ev.end_year IS NOT NULL AND et.end_year IS NOT NULL
          AND ABS(ev.end_year - et.end_year) <= 2
        LIMIT 1
        """,
        (target_id, viewer_id),
    ).fetchone()

    return compute_affinity_score(
        mutual_connections=m,
        same_school=school is not None,
        shared_skills=sk,
        same_company=company is not None,
        graduation_cohort=cohort is not None,
    )
