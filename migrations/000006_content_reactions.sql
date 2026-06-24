-- R2: unified content reactions + comment threads (live realm).

CREATE TABLE IF NOT EXISTS content_reactions (
    target_type TEXT NOT NULL,
    target_id   UUID NOT NULL,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    kind        TEXT NOT NULL DEFAULT 'like',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (target_type, target_id, user_id),
    CONSTRAINT content_reactions_target_type_valid CHECK (target_type IN ('post', 'comment')),
    CONSTRAINT content_reactions_kind_valid CHECK (
        kind IN ('like', 'celebrate', 'support', 'insightful', 'love', 'funny')
    )
);

CREATE INDEX IF NOT EXISTS idx_content_reactions_target
    ON content_reactions(target_type, target_id);
CREATE INDEX IF NOT EXISTS idx_content_reactions_user
    ON content_reactions(user_id);

INSERT INTO content_reactions (target_type, target_id, user_id, kind, created_at)
SELECT 'post', post_id, user_id, kind, created_at
FROM reactions
ON CONFLICT DO NOTHING;

ALTER TABLE reactions DROP CONSTRAINT IF EXISTS reactions_kind_valid;
ALTER TABLE reactions ADD CONSTRAINT reactions_kind_valid CHECK (
    kind IN ('like', 'celebrate', 'support', 'insightful', 'love', 'funny')
);

ALTER TABLE comments
    ADD COLUMN IF NOT EXISTS parent_comment_id UUID REFERENCES comments(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_comments_parent
    ON comments(parent_comment_id)
    WHERE parent_comment_id IS NOT NULL;
