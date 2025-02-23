-- +goose Up
CREATE TYPE external_resource_type AS ENUM(
  'youtube',
  'article',
  'book'
);

CREATE TABLE IF NOT EXISTS external_resources(
  id SERIAL PRIMARY KEY,
  topic_id INTEGER NOT NULL,
  title VARCHAR(255) NOT NULL,
  url VARCHAR(255) NOT NULL,
  type external_resource_type NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT (now()),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT (now()),
  CONSTRAINT external_resources_fk_topic FOREIGN KEY (topic_id) REFERENCES topics(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS external_resources;

DROP TYPE IF EXISTS external_resource_type;
