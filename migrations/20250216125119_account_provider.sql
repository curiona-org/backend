-- +goose Up
CREATE TYPE account_provider AS ENUM(
  'email',
  'google'
);

ALTER TABLE accounts
  ADD COLUMN provider account_provider NOT NULL DEFAULT 'email';

-- +goose Down
ALTER TABLE accounts
  DROP COLUMN provider;

DROP TYPE IF EXISTS account_provider;
