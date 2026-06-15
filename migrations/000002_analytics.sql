-- Precomputed analytics scores and daily rollups (populated by worker-batch).

CREATE SCHEMA IF NOT EXISTS analytics;

-- ---------------------------------------------------------------------------
-- Graph metrics (PageRank, centrality, communities)
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS user_graph_metrics (
    user_id         UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    pagerank        DOUBLE PRECISION NOT NULL DEFAULT 0,
    degree          INT NOT NULL DEFAULT 0,
    in_degree       INT NOT NULL DEFAULT 0,
    out_degree      INT NOT NULL DEFAULT 0,
    community_id    INT,
    computed_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_user_graph_metrics_pagerank ON user_graph_metrics(pagerank DESC);

-- ---------------------------------------------------------------------------
-- Pairwise affinity (viewer → target)
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS user_pair_affinity (
    viewer_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    score           DOUBLE PRECISION NOT NULL,
    reasons         JSONB NOT NULL DEFAULT '[]',
    computed_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (viewer_id, target_id),
    CONSTRAINT user_pair_affinity_distinct CHECK (viewer_id <> target_id)
);

CREATE INDEX IF NOT EXISTS idx_user_pair_affinity_viewer_score ON user_pair_affinity(viewer_id, score DESC);

-- ---------------------------------------------------------------------------
-- Connection suggestions (top-K per viewer)
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS user_connection_suggestions (
    viewer_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    suggested_user_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    score               DOUBLE PRECISION NOT NULL,
    reasons             JSONB NOT NULL DEFAULT '[]',
    rank                INT NOT NULL,
    computed_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (viewer_id, suggested_user_id),
    CONSTRAINT user_connection_suggestions_distinct CHECK (viewer_id <> suggested_user_id)
);

CREATE INDEX IF NOT EXISTS idx_connection_suggestions_viewer_rank ON user_connection_suggestions(viewer_id, rank);

-- ---------------------------------------------------------------------------
-- Feed ranking (precomputed post scores per user)
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS user_feed_scores (
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    post_id         UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    score           DOUBLE PRECISION NOT NULL,
    computed_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, post_id)
);

CREATE INDEX IF NOT EXISTS idx_user_feed_scores_user_score ON user_feed_scores(user_id, score DESC);

-- ---------------------------------------------------------------------------
-- Churn prediction
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS user_churn_scores (
    user_id             UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    churn_probability   DOUBLE PRECISION NOT NULL,
    risk_tier           TEXT NOT NULL DEFAULT 'low',
    features            JSONB NOT NULL DEFAULT '{}',
    computed_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT user_churn_risk_tier_valid CHECK (risk_tier IN ('low', 'medium', 'high'))
);

CREATE INDEX IF NOT EXISTS idx_user_churn_scores_probability ON user_churn_scores(churn_probability DESC);

-- ---------------------------------------------------------------------------
-- ML model versions
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS model_versions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_name      TEXT NOT NULL,
    version         TEXT NOT NULL,
    metrics         JSONB NOT NULL DEFAULT '{}',
    artifact_path   TEXT,
    is_active       BOOLEAN NOT NULL DEFAULT false,
    trained_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (model_name, version)
);

CREATE INDEX IF NOT EXISTS idx_model_versions_active ON model_versions(model_name) WHERE is_active;

-- ---------------------------------------------------------------------------
-- Daily rollups
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS analytics.daily_active_users (
    day             DATE NOT NULL,
    dau             INT NOT NULL DEFAULT 0,
    computed_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (day)
);

CREATE TABLE IF NOT EXISTS analytics.post_engagement_daily (
    day             DATE NOT NULL,
    post_id         UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    views           INT NOT NULL DEFAULT 0,
    reactions       INT NOT NULL DEFAULT 0,
    comments        INT NOT NULL DEFAULT 0,
    computed_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (day, post_id)
);

CREATE INDEX IF NOT EXISTS idx_post_engagement_daily_day ON analytics.post_engagement_daily(day);

CREATE TABLE IF NOT EXISTS analytics.user_cohorts (
    cohort_week     DATE NOT NULL,
    week_offset     INT NOT NULL,
    active_users    INT NOT NULL DEFAULT 0,
    cohort_size     INT NOT NULL DEFAULT 0,
    computed_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (cohort_week, week_offset)
);

-- ---------------------------------------------------------------------------
-- In-app notifications (worker-realtime)
-- ---------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS notifications (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    kind            TEXT NOT NULL,
    payload         JSONB NOT NULL DEFAULT '{}',
    read_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_notifications_user_created ON notifications(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_user_unread ON notifications(user_id) WHERE read_at IS NULL;
