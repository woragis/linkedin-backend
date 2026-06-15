"""Connection suggestions and affinity scoring (phase 3)."""

from __future__ import annotations

import logging
from typing import Any

import psycopg

log = logging.getLogger("linkedin-worker.recommendations")

# Initial feature weights — see docs/WORKERS.md
WEIGHTS = {
    "mutual_connections": 0.35,
    "same_school": 0.15,
    "shared_skills": 0.12,
    "same_company": 0.10,
    "graduation_cohort": 0.08,
    "same_field": 0.05,
    "same_location": 0.05,
    "age_proximity": 0.03,
    "pagerank": 0.03,
}


def recompute_user(conn: psycopg.Connection, payload: dict[str, Any]) -> None:
    user_id = payload.get("user_id")
    log.info("recompute_user user_id=%s", user_id)
    # TODO phase 3: incremental suggestions for one user


def run_batch(conn: psycopg.Connection) -> None:
    log.info("recommendations batch started")
    # TODO phase 3: affinity scorer + link prediction → user_connection_suggestions
    log.info("recommendations batch finished (stub)")
