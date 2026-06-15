"""Precomputed feed ranking per user (phase 5)."""

from __future__ import annotations

import logging

import psycopg

log = logging.getLogger("linkedin-worker.feed")


def run_batch(conn: psycopg.Connection) -> None:
    log.info("feed_ranking batch started")
    # TODO phase 5: score posts for each active user → user_feed_scores
    log.info("feed_ranking batch finished (stub)")
