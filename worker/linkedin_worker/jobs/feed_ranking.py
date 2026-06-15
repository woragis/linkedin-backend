"""Precomputed feed ranking per user."""

from __future__ import annotations

import logging
import math
from datetime import datetime, timezone

import psycopg

log = logging.getLogger("linkedin-worker.feed")


def run_batch(conn: psycopg.Connection) -> None:
    log.info("feed_ranking batch started")
    now = datetime.now(timezone.utc)
    users = conn.execute("SELECT user_id::text FROM profiles").fetchall()

    for (viewer_id,) in users:
        peers = conn.execute(
            """
            SELECT CASE WHEN requester_id = %s::uuid THEN addressee_id ELSE requester_id END
            FROM connections
            WHERE status = 'accepted' AND %s::uuid IN (requester_id, addressee_id)
            """,
            (viewer_id, viewer_id),
        ).fetchall()
        author_ids = [viewer_id] + [str(p[0]) for p in peers]

        posts = conn.execute(
            """
            SELECT po.id::text, po.author_id::text, po.created_at,
                   COALESCE(gm.pagerank, 0) AS pagerank,
                   (SELECT COUNT(*)::int FROM reactions r WHERE r.post_id = po.id) AS rxn,
                   (SELECT COUNT(*)::int FROM comments c WHERE c.post_id = po.id AND c.deleted_at IS NULL) AS cmts
            FROM posts po
            LEFT JOIN user_graph_metrics gm ON gm.user_id = po.author_id
            WHERE po.deleted_at IS NULL AND po.author_id = ANY(%s::uuid[])
            ORDER BY po.created_at DESC
            LIMIT 200
            """,
            (author_ids,),
        ).fetchall()

        conn.execute("DELETE FROM user_feed_scores WHERE user_id = %s::uuid", (viewer_id,))
        for rank, row in enumerate(posts, start=1):
            post_id, _author, created_at, pagerank, rxn, cmts = row
            age_hours = max(1.0, (now - created_at).total_seconds() / 3600)
            recency = 1.0 / math.log2(age_hours + 2)
            engagement = math.log1p(rxn + cmts * 2)
            score = recency * 2.0 + engagement + float(pagerank) * 5.0

            conn.execute(
                """
                INSERT INTO user_feed_scores (user_id, post_id, score, computed_at)
                VALUES (%s::uuid, %s::uuid, %s, now())
                """,
                (viewer_id, post_id, score),
            )
            if rank >= 100:
                break

    conn.commit()
    log.info("feed_ranking batch finished users=%d", len(users))
