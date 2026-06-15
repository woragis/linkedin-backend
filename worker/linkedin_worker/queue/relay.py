"""Transactional outbox → Redis (+ optional Kafka) relay."""

from __future__ import annotations

import json
import logging
import time
from typing import Any

import psycopg
import redis

from linkedin_worker import settings

log = logging.getLogger("linkedin-worker.relay")

FETCH_SQL = """
SELECT id, job_type, payload
FROM outbox_jobs
WHERE processed_at IS NULL
ORDER BY created_at
LIMIT 50
FOR UPDATE SKIP LOCKED
"""

MARK_SQL = """
UPDATE outbox_jobs SET processed_at = now() WHERE id = %s
"""

ERROR_SQL = """
UPDATE outbox_jobs SET last_error = %s WHERE id = %s
"""

_kafka_producer = None


def _kafka_producer():
    global _kafka_producer
    if _kafka_producer is not None:
        return _kafka_producer
    if not settings.KAFKA_BROKERS:
        return None
    try:
        from kafka import KafkaProducer

        _kafka_producer = KafkaProducer(
            bootstrap_servers=settings.KAFKA_BROKERS.split(","),
            value_serializer=lambda v: json.dumps(v).encode(),
        )
        log.info("kafka producer connected brokers=%s", settings.KAFKA_BROKERS)
    except Exception:
        log.exception("kafka producer init failed")
        return None
    return _kafka_producer


def _publish(r: redis.Redis, job_type: str, payload: dict[str, Any]) -> None:
    body = {"type": job_type, "payload": payload}
    r.lpush(settings.REDIS_QUEUE_KEY, json.dumps(body))
    producer = _kafka_producer()
    if producer is not None:
        producer.send(settings.KAFKA_TOPIC, body)
        producer.flush(1)


def relay_once(conn: psycopg.Connection, r: redis.Redis) -> int:
    published = 0
    with conn.transaction():
        rows = conn.execute(FETCH_SQL).fetchall()
        for row in rows:
            job_id, job_type, payload = row[0], row[1], row[2]
            try:
                if isinstance(payload, str):
                    payload = json.loads(payload)
                _publish(r, job_type, payload or {})
                conn.execute(MARK_SQL, (job_id,))
                published += 1
            except Exception as exc:
                log.exception("relay failed job_id=%s", job_id)
                conn.execute(ERROR_SQL, (str(exc)[:500], job_id))
    return published


def relay_loop(conn: psycopg.Connection, r: redis.Redis) -> None:
    log.info(
        "outbox relay started interval=%ss kafka=%s",
        settings.OUTBOX_POLL_INTERVAL_SEC,
        bool(settings.KAFKA_BROKERS),
    )
    while True:
        try:
            n = relay_once(conn, r)
            if n:
                log.info("relayed %d outbox jobs", n)
        except Exception:
            log.exception("relay loop error")
        time.sleep(settings.OUTBOX_POLL_INTERVAL_SEC)
