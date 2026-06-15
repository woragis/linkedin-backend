"""Churn prediction batch (phase 5)."""

from __future__ import annotations

import logging

import psycopg

log = logging.getLogger("linkedin-worker.churn")


def run_batch(conn: psycopg.Connection) -> None:
    log.info("churn batch started")
    # TODO phase 5: compute features, score users → user_churn_scores
    log.info("churn batch finished (stub)")
