-- +goose Up
CREATE TYPE account_method AS ENUM (
    'email',
    'google'
);
CREATE TABLE IF NOT EXISTS accounts (
    id SERIAL PRIMARY KEY,
    method account_method NOT NULL DEFAULT 'email',
    email VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    is_suspended boolean NOT NULL DEFAULT FALSE,
    is_admin boolean NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT (now()),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT (now()),
    deleted_at TIMESTAMPTZ
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_accounts_email ON accounts (email);
CREATE INDEX IF NOT EXISTS idx_accounts_deleted_at ON accounts (deleted_at);

CREATE TABLE IF NOT EXISTS sessions (
    id SERIAL PRIMARY KEY,
    account_id INTEGER NOT NULL,
    refresh_token VARCHAR(255) NOT NULL,
    user_agent VARCHAR(255) NOT NULL,
    client_ip VARCHAR(39) NOT NULL,
    is_blocked BOOLEAN NOT NULL DEFAULT FALSE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT (now()),
    CONSTRAINT fk_sessions_account FOREIGN KEY (
        account_id
    ) REFERENCES accounts (id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_sessions_account_id ON sessions (account_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_sessions_refresh_token ON sessions (refresh_token);

CREATE TABLE IF NOT EXISTS profiles (
    id INTEGER PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    avatar VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT (now()),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT (now()),
    CONSTRAINT fk_profiles FOREIGN KEY (id) REFERENCES accounts (
        id
    ) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_profiles_account_id ON profiles (id);

CREATE TABLE IF NOT EXISTS roadmaps (
    id SERIAL PRIMARY KEY,
    account_id INTEGER NOT NULL,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    total_topics INTEGER NOT NULL DEFAULT 0,
    total_finished_topics INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT (now()),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT (now()),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT fk_roadmaps_account FOREIGN KEY (
        account_id
    ) REFERENCES accounts (id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_roadmaps_account_id ON roadmaps (account_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_roadmaps_slug ON roadmaps (slug);
CREATE INDEX IF NOT EXISTS idx_roadmaps_deleted_at ON roadmaps (deleted_at);

CREATE TABLE IF NOT EXISTS topics (
    id SERIAL PRIMARY KEY,
    account_id INTEGER NOT NULL,
    roadmap_id INTEGER NOT NULL,
    parent_id INTEGER,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    "order" INTEGER NOT NULL,
    is_finished BOOLEAN NOT NULL DEFAULT FALSE,
    completed_at TIMESTAMPTZ,
    external_search_query TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT (now()),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT (now()),
    CONSTRAINT fk_topics FOREIGN KEY (roadmap_id) REFERENCES roadmaps (
        id
    ) ON DELETE CASCADE,
    CONSTRAINT fk_topics_parent FOREIGN KEY (parent_id) REFERENCES topics (
        id
    ) ON DELETE CASCADE
);
-- CREATE INDEX idx_topics_account_id ON topics (account_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_topics_slug ON topics (slug);
CREATE INDEX IF NOT EXISTS idx_topics_roadmap_id ON topics (roadmap_id);
CREATE INDEX IF NOT EXISTS idx_topics_parent_id ON topics (parent_id);

CREATE TYPE personalization_option_skill_level AS ENUM (
    'beginner',
    'intermediate',
    'advanced'
);
CREATE TABLE IF NOT EXISTS personalization_options (
    id SERIAL PRIMARY KEY,
    account_id INTEGER,
    roadmap_id INTEGER,
    daily_time_availability INTERVAL NOT NULL,
    total_duration INTERVAL NOT NULL,
    skill_level personalization_option_skill_level NOT NULL,
    additional_info TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT (now()),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT (now()),
    CONSTRAINT fk_personalization_options_account FOREIGN KEY (
        account_id
    ) REFERENCES accounts (id) ON DELETE CASCADE,
    CONSTRAINT fk_personalization_options_roadmap FOREIGN KEY (
        roadmap_id
    ) REFERENCES roadmaps (id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_personalization_options_account_id ON personalization_options (account_id);
CREATE INDEX IF NOT EXISTS idx_personalization_options_roadmap_id ON personalization_options (roadmap_id);

CREATE TYPE external_resource_type AS ENUM (
    'youtube',
    'article',
    'book'
);
CREATE TABLE IF NOT EXISTS external_resources (
    id SERIAL PRIMARY KEY,
    topic_id INTEGER NOT NULL,
    title VARCHAR(255) NOT NULL,
    author VARCHAR(255),
    url VARCHAR(255) NOT NULL,
    cover_url VARCHAR(255),
    length VARCHAR(255), --- duration for videos and pages for books.
    type external_resource_type NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT (now()),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT (now()),
    CONSTRAINT fk_external_resources_topic FOREIGN KEY (
        topic_id
    ) REFERENCES topics (id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_external_resources_topic_id ON external_resources (topic_id);

-- +goose Down
DROP TABLE IF EXISTS external_resources;
DROP TYPE IF EXISTS external_resource_type;

DROP TABLE IF EXISTS personalization_options;
DROP TYPE IF EXISTS personalization_option_skill_level;

DROP TABLE IF EXISTS topics;

DROP TABLE IF EXISTS roadmaps;

DROP TABLE IF EXISTS profiles;

DROP TABLE IF EXISTS sessions;

DROP TABLE IF EXISTS accounts;
DROP TYPE IF EXISTS account_method;
