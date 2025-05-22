-- +goose Up
ALTER TABLE topics
ADD COLUMN pro_tips TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE topics
DROP COLUMN pro_tips;
