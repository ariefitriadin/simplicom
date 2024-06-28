-- migrate:up

CREATE TABLE IF NOT EXISTS oauth2_tokens (
    id SERIAL PRIMARY KEY,
    access_token TEXT NOT NULL,
    refresh_token TEXT,
    expires_at TIMESTAMP NOT NULL,
    client_id UUID NOT NULL,
    user_id UUID NOT NULL
);

CREATE TABLE IF NOT EXISTS oauth2_clients (
    id SERIAL PRIMARY KEY,
    client_id UUID NOT NULL UNIQUE,
    client_secret UUID NOT NULL,
    domain TEXT NOT NULL,
    redirect_url TEXT NOT NULL,
    scope JSONB NOT NULL
);

-- migrate:down

-- Drop oauth2_token table
DROP TABLE IF EXISTS oauth2_tokens;

-- Drop oauth2_clients table
DROP INDEX IF EXISTS oauth2_clients;


