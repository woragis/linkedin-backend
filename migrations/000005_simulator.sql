-- Simulator agent metadata (extends users/profiles for synthetic population).

CREATE TABLE IF NOT EXISTS simulator_agents (
    user_id         UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    archetype       TEXT NOT NULL,
    age             INT NOT NULL,
    gender          TEXT NOT NULL,
    city            TEXT NOT NULL,
    latitude        DOUBLE PRECISION,
    longitude       DOUBLE PRECISION,
    extraversion    DOUBLE PRECISION NOT NULL DEFAULT 0.5,
    activity_level  DOUBLE PRECISION NOT NULL DEFAULT 0.5,
    interests       JSONB NOT NULL DEFAULT '[]',
    markov_state    TEXT NOT NULL DEFAULT 'offline',
    rng_offset      INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_simulator_agents_archetype ON simulator_agents(archetype);
CREATE INDEX IF NOT EXISTS idx_simulator_agents_city ON simulator_agents(city);
