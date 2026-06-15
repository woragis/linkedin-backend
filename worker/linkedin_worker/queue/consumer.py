"""Redis queue job dispatch."""

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

log = logging.getLogger("linkedin-worker.queue")

Handler = Callable[[psycopg.Connection, dict[str, Any]], None]

HANDLERS: dict[str, Handler] = {
    "search.index_profile": indexer.index_profile,
    "search.index_post": indexer.index_post,
    "analytics.process_event": events_processor.process_event,
    "notifications.send": notifications.send,
    "graph.recompute_user": recommendations.recompute_user,
}


def dispatch(conn: psycopg.Connection, job_type: str, payload: dict[str, Any]) -> None:
    handler = HANDLERS.get(job_type)
    if handler is None:
        log.warning("unknown job type: %s", job_type)
        return
    handler(conn, payload)


def consume_loop(r: redis.Redis, conn: psycopg.Connection) -> None:
    while True:
        item = r.brpop(settings.REDIS_QUEUE_KEY, timeout=5)
        if not item:
            continue
        _, raw = item
        try:
            env = json.loads(raw)
        except json.JSONDecodeError:
            log.warning("invalid job json: %s", raw[:120])
            continue
        job_type = env.get("type", "")
        payload = env.get("payload") or {}
        log.info("job type=%s", job_type)
        try:
            dispatch(conn, job_type, payload)
        except Exception:
            log.exception("job failed type=%s", job_type)
