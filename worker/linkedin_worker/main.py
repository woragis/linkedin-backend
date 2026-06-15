"""Worker entrypoint — roles: realtime, batch, all."""

from __future__ import annotations

import logging
import threading

import psycopg
import redis

from linkedin_worker import settings
from linkedin_worker.queue.consumer import consume_loop
from linkedin_worker.queue.relay import relay_loop
from linkedin_worker.scheduler import batch as batch_scheduler

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
log = logging.getLogger("linkedin-worker")


def _connect_db() -> psycopg.Connection:
    return psycopg.connect(settings.DATABASE_URL)


def _connect_redis() -> redis.Redis:
    return redis.from_url(settings.REDIS_URL, decode_responses=True)


def run_realtime() -> None:
    relay_conn = _connect_db()
    consumer_conn = _connect_db()
    r = _connect_redis()

    relay_thread = threading.Thread(target=relay_loop, args=(relay_conn, r), daemon=True)
    relay_thread.start()
    consume_loop(r, consumer_conn)


def run_batch() -> None:
    conn = _connect_db()
    if not settings.BATCH_ENABLED:
        log.info("batch disabled; sleeping")
        threading.Event().wait()
        return
    batch_scheduler.start(conn)


def main() -> None:
    role = settings.WORKER_ROLE
    log.info("worker starting role=%s", role)

    if role == "realtime":
        run_realtime()
    elif role == "batch":
        run_batch()
    elif role == "all":
        batch_thread = threading.Thread(target=run_batch, daemon=True)
        batch_thread.start()
        run_realtime()
    else:
        raise SystemExit(f"unknown WORKER_ROLE: {role}")


if __name__ == "__main__":
    main()
