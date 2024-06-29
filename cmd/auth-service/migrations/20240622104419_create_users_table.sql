-- migrate:up

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    LIKE template_table INCLUDING ALL,
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role SMALLINT NOT NULL DEFAULT 0
);

-- migrate:down

-- Drop users table
DROP TABLE IF EXISTS users;