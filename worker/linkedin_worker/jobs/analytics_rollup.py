"""DAU, MAU, cohorts and post engagement rollups (phase 4)."""

from __future__ import annotations

import logging

import psycopg

log = logging.getLogger("linkedin-worker.analytics")


def run_batch(conn: psycopg.Connection) -> None:
    log.info("analytics_rollup batch started")
    # TODO phase 4: aggregate events → analytics.daily_active_users, etc.
    log.info("analytics_rollup batch finished (stub)")
