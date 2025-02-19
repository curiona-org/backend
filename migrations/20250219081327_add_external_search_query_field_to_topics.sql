-- +goose Up
ALTER TABLE topics
  ADD COLUMN external_search_query TEXT;

-- +goose Down
ALTER TABLE topics
  DROP COLUMN external_search_query;
