"""DB/Redis connections with startup retry for slow orchestrators."""

from __future__ import annotations

import logging
import time
from typing import TYPE_CHECKING
from urllib.parse import urlparse

import psycopg
import redis

from linkedin_worker import settings

if TYPE_CHECKING:
    from collections.abc import Callable

log = logging.getLogger("linkedin-worker.connections")

DEFAULT_MAX_ATTEMPTS = 60
DEFAULT_BASE_DELAY_SEC = 2.0
DEFAULT_MAX_DELAY_SEC = 15.0
CONNECT_TIMEOUT_SEC = 15


def _retry(
    label: str,
    connect: Callable[[], object],
    *,
    max_attempts: int = DEFAULT_MAX_ATTEMPTS,
    base_delay_sec: float = DEFAULT_BASE_DELAY_SEC,
    max_delay_sec: float = DEFAULT_MAX_DELAY_SEC,
) -> object:
    delay = base_delay_sec
    last_err: BaseException | None = None
    for attempt in range(1, max_attempts + 1):
        try:
            return connect()
        except BaseException as exc:  # noqa: BLE001 — retry transient driver/network errors
            last_err = exc
            if attempt >= max_attempts:
                break
            log.warning(
                "%s connect failed attempt=%s/%s: %s; retry in %.1fs",
                label,
                attempt,
                max_attempts,
                exc,
                delay,
            )
            time.sleep(delay)
            delay = min(delay * 1.25, max_delay_sec)
    assert last_err is not None
    raise last_err


def connect_db() -> psycopg.Connection:
    url = settings.DATABASE_URL
    host = urlparse(url).hostname or "?"

    def _connect() -> psycopg.Connection:
        conn = psycopg.connect(url, connect_timeout=CONNECT_TIMEOUT_SEC)
        conn.execute("SELECT 1")
        return conn

    log.info("connecting postgres host=%s", host)
    return _retry("postgres", _connect)  # type: ignore[return-value]


def connect_redis() -> redis.Redis:
    url = settings.REDIS_URL
    host = urlparse(url).hostname or "?"

    def _connect() -> redis.Redis:
        client = redis.from_url(
            url,
            decode_responses=True,
            socket_connect_timeout=CONNECT_TIMEOUT_SEC,
            socket_timeout=CONNECT_TIMEOUT_SEC,
        )
        client.ping()
        return client

    log.info("connecting redis host=%s", host)
    return _retry("redis", _connect)  # type: ignore[return-value]
