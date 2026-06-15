"""Process raw events into analytics tables (phase 2)."""

from __future__ import annotations

import logging
from typing import Any

import psycopg

log = logging.getLogger("linkedin-worker.events")


def process_event(conn: psycopg.Connection, payload: dict[str, Any]) -> None:
    event_id = payload.get("event_id")
    log.info("process_event event_id=%s", event_id)
    # TODO phase 2: normalize event, update lightweight counters
