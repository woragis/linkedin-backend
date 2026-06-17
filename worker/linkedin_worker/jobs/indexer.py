"""Index profiles and posts into Elasticsearch."""

from __future__ import annotations

import json
import logging
from typing import Any

import psycopg
import urllib.request

from linkedin_worker import settings

log = logging.getLogger("linkedin-worker.indexer")


def _es_request(method: str, path: str, body: dict[str, Any] | None = None) -> bool:
    if not settings.ELASTICSEARCH_URL:
        log.debug("elasticsearch disabled, skip %s", path)
        return True
    url = settings.ELASTICSEARCH_URL.rstrip("/") + path
    data = json.dumps(body).encode() if body is not None else None
    req = urllib.request.Request(url, data=data, method=method)
    req.add_header("Content-Type", "application/json")
    try:
        with urllib.request.urlopen(req, timeout=10) as res:
            res.read()
        return True
    except Exception as exc:
        log.warning("elasticsearch %s %s failed: %s", method, path, exc)
        return False


def _ensure_indices() -> None:
    for index, mappings in (
        ("people", {"mappings": {"properties": {"full_name": {"type": "text"}, "headline": {"type": "text"}, "bio": {"type": "text"}, "skills": {"type": "text"}, "schools": {"type": "text"}, "location": {"type": "text"}}}}),
        ("posts", {"mappings": {"properties": {"body": {"type": "text"}, "author_name": {"type": "text"}}}}),
    ):
        try:
            _es_request("PUT", f"/{index}", mappings)
        except Exception:
            pass


def index_profile(conn: psycopg.Connection, payload: dict[str, Any]) -> None:
    user_id = payload.get("user_id")
    if not user_id:
        return
    row = conn.execute(
        """
        SELECT p.user_id, p.slug, p.full_name, p.headline, p.bio, p.location,
               COALESCE(array_agg(DISTINCT s.name) FILTER (WHERE s.name IS NOT NULL), '{}') AS skills,
               COALESCE(array_agg(DISTINCT i.name) FILTER (WHERE i.name IS NOT NULL), '{}') AS schools
        FROM profiles p
        LEFT JOIN user_skills us ON us.user_id = p.user_id
        LEFT JOIN skills s ON s.id = us.skill_id
        LEFT JOIN educations e ON e.user_id = p.user_id
        LEFT JOIN institutions i ON i.id = e.institution_id
        WHERE p.user_id = %s
        GROUP BY p.user_id, p.slug, p.full_name, p.headline, p.bio, p.location
        """,
        (user_id,),
    ).fetchone()
    if not row:
        return
    doc = {
        "user_id": str(row[0]),
        "slug": row[1],
        "full_name": row[2],
        "headline": row[3],
        "bio": row[4],
        "location": row[5],
        "skills": list(row[6] or []),
        "schools": list(row[7] or []),
    }
    _ensure_indices()
    if _es_request("PUT", f"/people/_doc/{user_id}", doc):
        log.info("indexed profile user_id=%s", user_id)


def index_post(conn: psycopg.Connection, payload: dict[str, Any]) -> None:
    post_id = payload.get("post_id")
    if not post_id:
        return
    row = conn.execute(
        """
        SELECT po.id, po.author_id, pr.full_name, po.body
        FROM posts po
        JOIN profiles pr ON pr.user_id = po.author_id
        WHERE po.id = %s AND po.deleted_at IS NULL
        """,
        (post_id,),
    ).fetchone()
    if not row:
        return
    doc = {
        "post_id": str(row[0]),
        "author_id": str(row[1]),
        "author_name": row[2],
        "body": row[3],
    }
    _ensure_indices()
    if _es_request("PUT", f"/posts/_doc/{post_id}", doc):
        log.info("indexed post post_id=%s", post_id)
