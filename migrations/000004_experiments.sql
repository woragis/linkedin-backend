-- A/B experiments and feed variant assignments (phase 6).

CREATE TABLE IF NOT EXISTS ab_experiments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            TEXT NOT NULL UNIQUE,
    description     TEXT NOT NULL DEFAULT '',
    status          TEXT NOT NULL DEFAULT 'active',
    variants        JSONB NOT NULL DEFAULT '["control","treatment"]',
    primary_metric  TEXT NOT NULL DEFAULT 'engagement_rate',
    started_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    ended_at        TIMESTAMPTZ,
    CONSTRAINT ab_experiments_status_valid CHECK (status IN ('active', 'paused', 'completed'))
);

CREATE TABLE IF NOT EXISTS ab_assignments (
    experiment_id   UUID NOT NULL REFERENCES ab_experiments(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    variant         TEXT NOT NULL,
    assigned_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (experiment_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_ab_assignments_user ON ab_assignments(user_id);

CREATE TABLE IF NOT EXISTS ab_experiment_results (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    experiment_id   UUID NOT NULL REFERENCES ab_experiments(id) ON DELETE CASCADE,
    variant         TEXT NOT NULL,
    sample_size     INT NOT NULL DEFAULT 0,
    metric_value    DOUBLE PRECISION NOT NULL DEFAULT 0,
    ci_lower        DOUBLE PRECISION,
    ci_upper        DOUBLE PRECISION,
    computed_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_ab_results_experiment ON ab_experiment_results(experiment_id, computed_at DESC);

-- Default feed ranking experiment
INSERT INTO ab_experiments (name, description, variants, primary_metric)
VALUES (
    'feed_ranking_v1',
    'Chronological feed (control) vs precomputed ranked feed (treatment)',
    '["chronological","ranked"]',
    'engagement_rate'
) ON CONFLICT (name) DO NOTHING;
