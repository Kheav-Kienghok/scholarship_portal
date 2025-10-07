-- +goose Up
-- +goose StatementBegin
-- Create ENUM type only if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'oauth_provider') THEN
        CREATE TYPE oauth_provider AS ENUM ('google', 'facebook');
    END IF;
END$$;


CREATE TABLE oauth_logins (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    provider oauth_provider NOT NULL,
    provider_user_id VARCHAR(100) NOT NULL,
    access_token VARCHAR(255) NOT NULL,
    refresh_token VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (provider, provider_user_id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS oauth_logins CASCADE;
DROP TYPE IF EXISTS oauth_provider CASCADE;
-- +goose StatementEnd
