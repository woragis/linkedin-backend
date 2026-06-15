"""Graph metrics: PageRank, centrality, communities (phase 4)."""

from __future__ import annotations

import logging

import psycopg

log = logging.getLogger("linkedin-worker.graph")


def run_batch(conn: psycopg.Connection) -> None:
    log.info("graph batch started")
    # TODO phase 4: load connections, compute metrics, upsert user_graph_metrics
    log.info("graph batch finished (stub)")
