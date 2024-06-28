-- migrate:up

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    LIKE template_table INCLUDING ALL,
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20) UNIQUE NOT NULL,
    role SMALLINT NOT NULL DEFAULT 0,
    google_id varchar(255) DEFAULT NULL
);

-- Add index on google_id
CREATE INDEX IF NOT EXISTS idx_users_google_id ON users(google_id);

-- migrate:down

-- Drop users table
DROP TABLE IF EXISTS users;

-- Drop index on google_id
DROP INDEX IF EXISTS idx_users_google_id;
