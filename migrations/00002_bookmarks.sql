-- +goose Up
CREATE TABLE IF NOT EXISTS bookmarks (
    account_id INTEGER NOT NULL,
    roadmap_id INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT (now()),
    PRIMARY KEY (account_id, roadmap_id)
);

-- +goose Down
DROP TABLE IF EXISTS bookmarks;
