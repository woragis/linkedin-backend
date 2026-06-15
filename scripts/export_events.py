# Export events to CSV for data lake / offline analysis.
# Usage: python scripts/export_events.py > events_export.csv

from __future__ import annotations

import csv
import os
import sys

import psycopg

DATABASE_URL = os.getenv(
    "DATABASE_URL",
    "postgres://linkedin:linkedin@127.0.0.1:5432/linkedin?sslmode=disable",
)


def main() -> None:
    conn = psycopg.connect(DATABASE_URL)
    rows = conn.execute(
        """
        SELECT id::text, user_id::text, event_type, payload::text, created_at::text
        FROM events ORDER BY created_at
        """
    ).fetchall()
    writer = csv.writer(sys.stdout)
    writer.writerow(["id", "user_id", "event_type", "payload", "created_at"])
    for row in rows:
        writer.writerow(row)
    conn.close()


if __name__ == "__main__":
    main()
