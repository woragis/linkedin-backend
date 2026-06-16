"""Worker entrypoint — microservice roles via WORKER_ROLE."""

from __future__ import annotations

import logging
import os
import threading

from linkedin_worker import settings
from linkedin_worker.connections import connect_db, connect_redis
from linkedin_worker.health import start_health_server
from linkedin_worker.queue.consumer import consume_loop
from linkedin_worker.queue.relay import relay_loop
from linkedin_worker.scheduler import batch as batch_scheduler

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
log = logging.getLogger("linkedin-worker")

VALID_ROLES = frozenset({
    "realtime",
    "indexer",
    "graph",
    "ml",
    "batch",
    "simulator",
    "all",  # dev only
})


def run_realtime() -> None:
    relay_conn = connect_db()
    consumer_conn = connect_db()
    r = connect_redis()

    relay_thread = threading.Thread(target=relay_loop, args=(relay_conn, r), daemon=True)
    relay_thread.start()
    consume_loop(r, consumer_conn, role="realtime")


def run_indexer() -> None:
    conn = connect_db()
    r = connect_redis()
    consume_loop(r, conn, role="indexer")


def run_graph() -> None:
    r = connect_redis()
    consumer_conn = connect_db()

    if settings.BATCH_ENABLED:
        sched_conn = connect_db()
        threading.Thread(
            target=consume_loop,
            args=(r, consumer_conn),
            kwargs={"role": "graph"},
            daemon=True,
        ).start()
        batch_scheduler.start_graph(sched_conn)
    else:
        consume_loop(r, consumer_conn, role="graph")


def run_ml() -> None:
    conn = connect_db()
    if not settings.BATCH_ENABLED:
        log.info("ml disabled; sleeping")
        threading.Event().wait()
        return
    batch_scheduler.start_ml(conn)


def run_batch() -> None:
    conn = connect_db()
    if not settings.BATCH_ENABLED:
        log.info("batch disabled; sleeping")
        threading.Event().wait()
        return
    batch_scheduler.start_batch(conn)


def run_all() -> None:
    """Dev: relay + all queues + all crons in one container."""
    relay_conn = connect_db()
    consumer_conn = connect_db()
    r = connect_redis()

    batch_thread = threading.Thread(target=run_batch_legacy, daemon=True)
    batch_thread.start()

    relay_thread = threading.Thread(target=relay_loop, args=(relay_conn, r), daemon=True)
    relay_thread.start()
    consume_loop(r, consumer_conn, role="all")


def run_batch_legacy() -> None:
    conn = connect_db()
    if settings.BATCH_ENABLED:
        batch_scheduler.start_all(conn)


def main() -> None:
    role = settings.WORKER_ROLE
    if role not in VALID_ROLES:
        raise SystemExit(f"unknown WORKER_ROLE: {role!r}; valid: {sorted(VALID_ROLES)}")

    log.info(
        "worker starting role=%s health_port=%s railway=%s",
        role,
        settings.WORKER_HEALTH_PORT,
        bool(os.getenv("RAILWAY_ENVIRONMENT") or os.getenv("RAILWAY_SERVICE_NAME")),
    )
    start_health_server(role)

    runners = {
        "realtime": run_realtime,
        "indexer": run_indexer,
        "graph": run_graph,
        "ml": run_ml,
        "batch": run_batch,
        "all": run_all,
    }
    if role == "simulator":
        from linkedin_worker.simulator import run_simulator

        run_simulator()
        return
    runners[role]()


if __name__ == "__main__":
    main()
