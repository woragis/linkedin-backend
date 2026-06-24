"""DAU, MAU, cohorts and post engagement rollups."""

from __future__ import annotations

import logging
from datetime import date, timedelta

import psycopg

log = logging.getLogger("linkedin-worker.analytics")


def run_batch(conn: psycopg.Connection) -> None:
    log.info("analytics_rollup batch started")
    today = date.today()

    # DAU — distinct users with any event today
    dau_row = conn.execute(
        """
        SELECT COUNT(DISTINCT user_id)::int FROM events
        WHERE user_id IS NOT NULL AND created_at::date = %s
        """,
        (today,),
    ).fetchone()
    dau = int(dau_row[0]) if dau_row else 0
    conn.execute(
        """
        INSERT INTO analytics.daily_active_users (day, dau, computed_at)
        VALUES (%s, %s, now())
        ON CONFLICT (day) DO UPDATE SET dau = EXCLUDED.dau, computed_at = now()
        """,
        (today, dau),
    )

    # Post engagement daily
    conn.execute(
        """
        INSERT INTO analytics.post_engagement_daily (day, post_id, views, reactions, comments, computed_at)
        SELECT %s, p.id,
               COALESCE(SUM(CASE WHEN e.event_type = 'post_viewed' THEN 1 ELSE 0 END), 0)::int,
               (SELECT COUNT(*)::int FROM content_reactions cr
                WHERE cr.target_type = 'post' AND cr.target_id = p.id),
               (SELECT COUNT(*)::int FROM comments c WHERE c.post_id = p.id AND c.deleted_at IS NULL),
               now()
        FROM posts p
        LEFT JOIN events e ON e.event_type = 'post_viewed'
          AND (e.payload->>'post_id')::uuid = p.id
          AND e.created_at::date = %s
        WHERE p.deleted_at IS NULL
        GROUP BY p.id
        ON CONFLICT (day, post_id) DO UPDATE SET
          views = EXCLUDED.views,
          reactions = EXCLUDED.reactions,
          comments = EXCLUDED.comments,
          computed_at = now()
        """,
        (today, today),
    )

    # Cohort retention — users by registration week, active in subsequent weeks
    conn.execute("DELETE FROM analytics.user_cohorts WHERE cohort_week >= %s - interval '12 weeks'", (today,))
    conn.execute(
        """
        INSERT INTO analytics.user_cohorts (cohort_week, week_offset, active_users, cohort_size, computed_at)
        SELECT date_trunc('week', u.created_at)::date AS cohort_week,
               FLOOR(EXTRACT(EPOCH FROM (e.created_at - u.created_at)) / 604800)::int AS week_offset,
               COUNT(DISTINCT u.id)::int,
               cohort_size.cnt,
               now()
        FROM users u
        JOIN events e ON e.user_id = u.id
        JOIN (
            SELECT date_trunc('week', created_at)::date AS cw, COUNT(*)::int AS cnt
            FROM users GROUP BY 1
        ) cohort_size ON cohort_size.cw = date_trunc('week', u.created_at)::date
        WHERE u.created_at >= now() - interval '12 weeks'
        GROUP BY 1, 2, cohort_size.cnt
        ON CONFLICT (cohort_week, week_offset) DO UPDATE SET
          active_users = EXCLUDED.active_users,
          cohort_size = EXCLUDED.cohort_size,
          computed_at = now()
        """,
    )

    conn.commit()
    log.info("analytics_rollup batch finished dau=%d", dau)

    _compute_ab_results(conn)


def _compute_ab_results(conn: psycopg.Connection) -> None:
    """Engagement rate by feed experiment variant with normal-approx CI."""
    import math

    exp = conn.execute(
        "SELECT id::text FROM ab_experiments WHERE name = 'feed_ranking_v1' AND status = 'active' LIMIT 1"
    ).fetchone()
    if not exp:
        return
    exp_id = exp[0]
    variants = conn.execute(
        "SELECT DISTINCT variant FROM ab_assignments WHERE experiment_id = %s::uuid", (exp_id,)
    ).fetchall()
    for (variant,) in variants:
        row = conn.execute(
            """
            SELECT COUNT(DISTINCT a.user_id)::int,
                   COUNT(*) FILTER (WHERE e.event_type IN ('post_viewed','post_liked'))::int
            FROM ab_assignments a
            LEFT JOIN events e ON e.user_id = a.user_id
              AND e.created_at >= now() - interval '7 days'
            WHERE a.experiment_id = %s::uuid AND a.variant = %s
            """,
            (exp_id, variant),
        ).fetchone()
        if not row or row[0] == 0:
            continue
        n, engaged = int(row[0]), int(row[1])
        rate = engaged / max(n, 1)
        # Wilson-ish normal approx 95% CI for proportion
        se = math.sqrt(rate * (1 - rate) / max(n, 1))
        ci_lo = max(0.0, rate - 1.96*se)
        ci_hi = min(1.0, rate + 1.96*se)
        conn.execute(
            """
            INSERT INTO ab_experiment_results
              (experiment_id, variant, sample_size, metric_value, ci_lower, ci_upper, computed_at)
            VALUES (%s::uuid, %s, %s, %s, %s, %s, now())
            """,
            (exp_id, variant, n, rate, ci_lo, ci_hi),
        )
    conn.commit()
