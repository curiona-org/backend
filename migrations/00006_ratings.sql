-- +goose Up
CREATE TABLE roadmap_ratings (
    roadmap_id INTEGER NOT NULL,
    account_id INTEGER NOT NULL,
    progression_total_topics INTEGER NOT NULL DEFAULT 0,
    progression_total_finished_topics INTEGER NOT NULL DEFAULT 0,
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comment TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT (now()),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT (now()),
    CONSTRAINT fk_roadmap_ratings_roadmap FOREIGN KEY (
        roadmap_id
    ) REFERENCES roadmaps (id) ON DELETE CASCADE,
    CONSTRAINT fk_roadmap_ratings_account FOREIGN KEY (
        account_id
    ) REFERENCES accounts (id) ON DELETE CASCADE,
    PRIMARY KEY (roadmap_id, account_id)
);

-- +goose Down
DROP TABLE IF EXISTS roadmap_ratings;
