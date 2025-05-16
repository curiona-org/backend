-- +goose Up
CREATE TABLE IF NOT EXISTS roadmap_topic_progressions (
    id SERIAL PRIMARY KEY,
    account_id INTEGER NOT NULL,
    roadmap_id INTEGER NOT NULL,
    topic_id INTEGER NOT NULL,
    is_finished BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT (now()),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT (now()),
    CONSTRAINT fk_roadmap_topic_progressions_account FOREIGN KEY (account_id) REFERENCES accounts (id) ON DELETE CASCADE,
    CONSTRAINT fk_roadmap_topic_progressions_roadmap FOREIGN KEY (roadmap_id) REFERENCES roadmaps (id) ON DELETE CASCADE,
    CONSTRAINT fk_roadmap_topic_progressions_topic FOREIGN KEY (topic_id) REFERENCES topics (id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_roadmap_topic_progressions_account_roadmap_topic ON roadmap_topic_progressions (account_id, roadmap_id, topic_id);

ALTER TABLE roadmaps DROP COLUMN IF EXISTS total_finished_topics;
ALTER TABLE topics DROP COLUMN IF EXISTS is_finished;

-- +goose Down
DROP TABLE IF EXISTS roadmap_topic_progressions;
