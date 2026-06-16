"""Normalize DATABASE_URL for psycopg3 and hosted Postgres (e.g. Railway)."""

from __future__ import annotations

import os
from urllib.parse import parse_qsl, urlencode, urlparse, urlunparse


def _hosted_postgres(url: str) -> bool:
    host = (urlparse(url).hostname or "").lower()
    if host in ("localhost", "127.0.0.1", "postgres"):
        return False
    if os.getenv("RAILWAY_ENVIRONMENT") or os.getenv("RAILWAY_SERVICE_NAME"):
        return True
    return "railway" in host


def normalize_database_url(url: str) -> str:
    """Use postgresql:// scheme and sslmode on hosted Postgres when missing."""
    if url.startswith("postgres://"):
        url = "postgresql://" + url[len("postgres://") :]

    parsed = urlparse(url)
    query = dict(parse_qsl(parsed.query, keep_blank_values=True))

    sslmode = os.getenv("DATABASE_SSLMODE", "").strip()
    if sslmode:
        query["sslmode"] = sslmode
    elif "sslmode" not in query and _hosted_postgres(url):
        query["sslmode"] = "require"

    return urlunparse(parsed._replace(query=urlencode(query)))
