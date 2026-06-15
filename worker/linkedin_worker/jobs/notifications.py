"""In-app notifications (phase 4+)."""

from __future__ import annotations

import logging
from typing import Any

import psycopg

log = logging.getLogger("linkedin-worker.notifications")


def send(conn: psycopg.Connection, payload: dict[str, Any]) -> None:
    user_id = payload.get("user_id")
    kind = payload.get("kind")
    log.info("notification user_id=%s kind=%s", user_id, kind)
    # TODO phase 4: insert into notifications table
