"""Weekly ML model training (phase 6)."""

from __future__ import annotations

import logging

import psycopg

log = logging.getLogger("linkedin-worker.ml")


def run_batch(conn: psycopg.Connection) -> None:
    log.info("ml_training batch started")
    # TODO phase 6: train connection acceptance model → model_versions
    log.info("ml_training batch finished (stub)")
