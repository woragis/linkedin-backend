"""Cron schedulers split by worker role."""

from __future__ import annotations

import logging

import psycopg
from apscheduler.schedulers.blocking import BlockingScheduler

from linkedin_worker import settings
from linkedin_worker.jobs import (
    analytics_rollup,
    churn,
    feed_ranking,
    graph,
    ml_training,
    recommendations,
)

log = logging.getLogger("linkedin-worker.scheduler")


def start_graph(conn: psycopg.Connection) -> None:
    sched = BlockingScheduler()
    sched.add_job(
        graph.run_batch,
        "cron",
        **_cron_kwargs(settings.BATCH_CRON_GRAPH),
        args=[conn],
        id="graph",
        replace_existing=True,
    )
    log.info("graph scheduler started")
    sched.start()


def start_ml(conn: psycopg.Connection) -> None:
    sched = BlockingScheduler()
    sched.add_job(
        ml_training.run_batch,
        "cron",
        **_cron_kwargs(settings.BATCH_CRON_ML_TRAINING),
        args=[conn],
        id="ml_training",
        replace_existing=True,
    )
    log.info("ml scheduler started")
    sched.start()


def start_batch(conn: psycopg.Connection) -> None:
    """Core batch: recommendations, feed, churn, analytics (no graph / ml)."""
    sched = BlockingScheduler()
    sched.add_job(
        recommendations.run_batch,
        "cron",
        **_cron_kwargs(settings.BATCH_CRON_RECOMMENDATIONS),
        args=[conn],
        id="recommendations",
        replace_existing=True,
    )
    sched.add_job(
        feed_ranking.run_batch,
        "cron",
        **_cron_kwargs(settings.BATCH_CRON_FEED_RANKING),
        args=[conn],
        id="feed_ranking",
        replace_existing=True,
    )
    sched.add_job(
        churn.run_batch,
        "cron",
        **_cron_kwargs(settings.BATCH_CRON_CHURN),
        args=[conn],
        id="churn",
        replace_existing=True,
    )
    sched.add_job(
        analytics_rollup.run_batch,
        "cron",
        **_cron_kwargs(settings.BATCH_CRON_ANALYTICS),
        args=[conn],
        id="analytics_rollup",
        replace_existing=True,
    )
    log.info("batch scheduler started")
    sched.start()


def start_all(conn: psycopg.Connection) -> None:
    """Legacy WORKER_ROLE=batch — all cron jobs in one process."""
    sched = BlockingScheduler()
    sched.add_job(
        graph.run_batch,
        "cron",
        **_cron_kwargs(settings.BATCH_CRON_GRAPH),
        args=[conn],
        id="graph",
        replace_existing=True,
    )
    sched.add_job(
        recommendations.run_batch,
        "cron",
        **_cron_kwargs(settings.BATCH_CRON_RECOMMENDATIONS),
        args=[conn],
        id="recommendations",
        replace_existing=True,
    )
    sched.add_job(
        feed_ranking.run_batch,
        "cron",
        **_cron_kwargs(settings.BATCH_CRON_FEED_RANKING),
        args=[conn],
        id="feed_ranking",
        replace_existing=True,
    )
    sched.add_job(
        churn.run_batch,
        "cron",
        **_cron_kwargs(settings.BATCH_CRON_CHURN),
        args=[conn],
        id="churn",
        replace_existing=True,
    )
    sched.add_job(
        analytics_rollup.run_batch,
        "cron",
        **_cron_kwargs(settings.BATCH_CRON_ANALYTICS),
        args=[conn],
        id="analytics_rollup",
        replace_existing=True,
    )
    sched.add_job(
        ml_training.run_batch,
        "cron",
        **_cron_kwargs(settings.BATCH_CRON_ML_TRAINING),
        args=[conn],
        id="ml_training",
        replace_existing=True,
    )
    log.info("batch scheduler (all crons) started")
    sched.start()


def _cron_kwargs(expr: str) -> dict[str, str]:
    minute, hour, day, month, day_of_week = expr.split()
    return {
        "minute": minute,
        "hour": hour,
        "day": day,
        "month": month,
        "day_of_week": day_of_week,
    }
