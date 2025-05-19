-- +goose Up
CREATE TABLE IF NOT EXISTS roadmap_progressions (
    id SERIAL PRIMARY KEY,
    account_id INTEGER NOT NULL,
    roadmap_id INTEGER NOT NULL,
    total_topics INTEGER NOT NULL DEFAULT 0,
    total_finished_topics INTEGER NOT NULL DEFAULT 0,
    is_finished BOOLEAN NOT NULL DEFAULT FALSE,
    finished_at TIMESTAMP DEFAULT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT fk_roadmap_progressions_roadmap_id UNIQUE (account_id, roadmap_id),
    CONSTRAINT fk_roadmap_progressions_account FOREIGN KEY (account_id) REFERENCES accounts (id) ON DELETE CASCADE,
    CONSTRAINT fk_roadmap_progressions_roadmap FOREIGN KEY (roadmap_id) REFERENCES roadmaps (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_roadmap_progressions_account_id_roadmap_id ON roadmap_progressions (account_id, roadmap_id);

CREATE TABLE IF NOT EXISTS roadmap_topic_progressions (
    progression_id INTEGER NOT NULL,
    topic_id INTEGER NOT NULL,
    account_id INTEGER NOT NULL,
    is_finished BOOLEAN NOT NULL DEFAULT FALSE,
    finished_at TIMESTAMP DEFAULT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (progression_id, topic_id),
    CONSTRAINT fk_roadmap_topic_progressions_progression FOREIGN KEY (progression_id) REFERENCES roadmap_progressions (id) ON DELETE CASCADE,
    CONSTRAINT fk_roadmap_topic_progressions_account FOREIGN KEY (account_id) REFERENCES accounts (id) ON DELETE CASCADE,
    CONSTRAINT fk_roadmap_topic_progressions_topic FOREIGN KEY (topic_id) REFERENCES topics (id) ON DELETE CASCADE
);

ALTER TABLE roadmaps DROP COLUMN IF EXISTS total_finished_topics;
ALTER TABLE topics DROP COLUMN IF EXISTS is_finished;

-- +goose Down
DROP TABLE IF EXISTS roadmap_topic_progressions;
DROP TABLE IF EXISTS roadmap_progressions;
