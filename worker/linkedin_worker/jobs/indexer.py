"""Index profiles and posts into Elasticsearch (phase 3)."""

from __future__ import annotations

import logging
from typing import Any

import psycopg

from linkedin_worker import settings

log = logging.getLogger("linkedin-worker.indexer")


def index_profile(conn: psycopg.Connection, payload: dict[str, Any]) -> None:
    user_id = payload.get("user_id")
    log.info("index_profile user_id=%s es=%s", user_id, bool(settings.ELASTICSEARCH_URL))
    # TODO phase 3: fetch profile, upsert into ES people index


def index_post(conn: psycopg.Connection, payload: dict[str, Any]) -> None:
    post_id = payload.get("post_id")
    log.info("index_post post_id=%s es=%s", post_id, bool(settings.ELASTICSEARCH_URL))
    # TODO phase 3: fetch post, upsert into ES posts index
