-- +goose Up
CREATE TABLE sessions(
  id SERIAL PRIMARY KEY,
  account_id INTEGER NOT NULL,
  refresh_token VARCHAR(255) NOT NULL UNIQUE,
  user_agent VARCHAR(255) NOT NULL,
  client_ip VARCHAR(39) NOT NULL,
  blocked BOOLEAN NOT NULL DEFAULT FALSE,
  expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT (now()),
  CONSTRAINT sessions_fk_account FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS sessions;
