"""Normalize DATABASE_URL for psycopg3 and hosted Postgres (e.g. Railway)."""

from __future__ import annotations

import os
from urllib.parse import parse_qsl, urlencode, urlparse, urlunparse


def normalize_database_url(url: str) -> str:
    """Use postgresql:// scheme and sslmode=require on Railway when missing."""
    if url.startswith("postgres://"):
        url = "postgresql://" + url[len("postgres://") :]

    if os.getenv("RAILWAY_ENVIRONMENT"):
        parsed = urlparse(url)
        query = dict(parse_qsl(parsed.query, keep_blank_values=True))
        if "sslmode" not in query:
            query["sslmode"] = "require"
            url = urlunparse(parsed._replace(query=urlencode(query)))

    return url
