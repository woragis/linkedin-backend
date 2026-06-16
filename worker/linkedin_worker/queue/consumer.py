"""Redis queue job dispatch — handlers filtered by WORKER_ROLE."""

from __future__ import annotations

import json
import logging
from typing import Any, Callable

import psycopg
import redis

from linkedin_worker import settings
from linkedin_worker.jobs import (
    events_processor,
    indexer,
    notifications,
    recommendations,
)
from linkedin_worker.queue.routing import ROLE_JOB_TYPES, queues_for_role

log = logging.getLogger("linkedin-worker.queue")

Handler = Callable[[psycopg.Connection, dict[str, Any]], None]

ALL_HANDLERS: dict[str, Handler] = {
    "search.index_profile": indexer.index_profile,
    "search.index_post": indexer.index_post,
    "analytics.process_event": events_processor.process_event,
    "notifications.send": notifications.send,
    "graph.recompute_user": recommendations.recompute_user,
}


def handlers_for_role(role: str) -> dict[str, Handler]:
    if role == "all":
        return dict(ALL_HANDLERS)
    allowed = ROLE_JOB_TYPES.get(role, frozenset())
    return {k: v for k, v in ALL_HANDLERS.items() if k in allowed}


def dispatch(
    conn: psycopg.Connection,
    job_type: str,
    payload: dict[str, Any],
    *,
    role: str,
) -> None:
    handlers = handlers_for_role(role)
    handler = handlers.get(job_type)
    if handler is None:
        log.warning("unknown or disabled job type=%s role=%s", job_type, role)
        return
    handler(conn, payload)


def consume_loop(r: redis.Redis, conn: psycopg.Connection, *, role: str) -> None:
    queues = queues_for_role(role)
    if not queues:
        log.error("no queues for role=%s", role)
        raise SystemExit(f"role {role} does not consume Redis jobs")

    log.info("consumer started role=%s queues=%s", role, queues)
    while True:
        item = r.brpop(queues, timeout=5)
        if not item:
            continue
        queue_name, raw = item
        try:
            env = json.loads(raw)
        except json.JSONDecodeError:
            log.warning("invalid job json on %s: %s", queue_name, raw[:120])
            continue
        job_type = env.get("type", "")
        payload = env.get("payload") or {}
        log.info("job queue=%s type=%s", queue_name, job_type)
        try:
            dispatch(conn, job_type, payload, role=role)
        except Exception:
            log.exception("job failed type=%s", job_type)
