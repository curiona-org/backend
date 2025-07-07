-- +goose Up
CREATE TABLE IF NOT EXISTS bookmarks (
    account_id INTEGER NOT NULL,
    roadmap_id INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT (now()),
    PRIMARY KEY (account_id, roadmap_id),
    CONSTRAINT fk_bookmarks_account
    FOREIGN KEY (account_id)
    REFERENCES accounts (id)
    ON DELETE CASCADE,
    CONSTRAINT fk_bookmarks_roadmap
    FOREIGN KEY (roadmap_id)
    REFERENCES roadmaps (id)
    ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS bookmarks;
