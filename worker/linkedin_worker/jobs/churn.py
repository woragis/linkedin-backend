"""Churn prediction batch."""

from __future__ import annotations

import json
import logging
import math
from datetime import datetime, timezone

import psycopg

log = logging.getLogger("linkedin-worker.churn")


def _sigmoid(x: float) -> float:
    return 1.0 / (1.0 + math.exp(-x))


def run_batch(conn: psycopg.Connection) -> None:
    log.info("churn batch started")
    now = datetime.now(timezone.utc)
    users = conn.execute("SELECT id::text FROM users").fetchall()

    for (uid,) in users:
        last_event = conn.execute(
            "SELECT MAX(created_at) FROM events WHERE user_id = %s::uuid",
            (uid,),
        ).fetchone()
        days_idle = 30.0
        if last_event and last_event[0]:
            days_idle = max(0, (now - last_event[0]).total_seconds() / 86400)

        posts_7d = conn.execute(
            """
            SELECT COUNT(*)::int FROM events
            WHERE user_id = %s::uuid AND event_type IN ('post_created', 'post_viewed')
              AND created_at >= now() - interval '7 days'
            """,
            (uid,),
        ).fetchone()
        activity = int(posts_7d[0]) if posts_7d else 0

        connections = conn.execute(
            """
            SELECT COUNT(*)::int FROM connections
            WHERE status = 'accepted' AND (%s::uuid IN (requester_id, addressee_id))
            """,
            (uid,),
        ).fetchone()
        conn_count = int(connections[0]) if connections else 0

        profile = conn.execute(
            "SELECT headline, bio FROM profiles WHERE user_id = %s::uuid",
            (uid,),
        ).fetchone()
        completion = 0.3
        if profile:
            if profile[0]:
                completion += 0.2
            if profile[1]:
                completion += 0.2

        # Logistic-style score (hand-tuned coefficients)
        logit = -2.0 + 0.08 * days_idle - 0.15 * activity - 0.1 * conn_count - 0.5 * completion
        prob = _sigmoid(logit)
        if prob >= 0.7:
            tier = "high"
        elif prob >= 0.4:
            tier = "medium"
        else:
            tier = "low"

        features = {
            "days_idle": round(days_idle, 2),
            "activity_7d": activity,
            "connections": conn_count,
            "profile_completion": completion,
        }
        conn.execute(
            """
            INSERT INTO user_churn_scores (user_id, churn_probability, risk_tier, features, computed_at)
            VALUES (%s::uuid, %s, %s, %s::jsonb, now())
            ON CONFLICT (user_id) DO UPDATE SET
              churn_probability = EXCLUDED.churn_probability,
              risk_tier = EXCLUDED.risk_tier,
              features = EXCLUDED.features,
              computed_at = now()
            """,
            (uid, prob, tier, json.dumps(features)),
        )

    conn.commit()
    log.info("churn batch finished users=%d", len(users))
