-- +goose Up
ALTER TABLE profiles
ADD COLUMN max_generated_roadmaps INTEGER NOT NULL DEFAULT 5;

-- +goose Down
ALTER TABLE profiles
DROP COLUMN max_generated_roadmaps;
