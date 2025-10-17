-- +goose Up
-- +goose StatementBegin

-- Add email verification column to users table
ALTER TABLE users ADD COLUMN email_verified BOOLEAN DEFAULT FALSE;

-- Create email_verifications table
CREATE TABLE email_verifications (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    verified_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create index for faster lookups
CREATE INDEX idx_email_verifications_token ON email_verifications(token);
CREATE INDEX idx_email_verifications_user_id ON email_verifications(user_id);
CREATE INDEX idx_email_verifications_expires_at ON email_verifications(expires_at);

-- Add trigger to update updated_at column
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_email_verifications_updated_at 
    BEFORE UPDATE ON email_verifications 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop trigger and function
DROP TRIGGER IF EXISTS update_email_verifications_updated_at ON email_verifications;
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_email_verifications_expires_at;
DROP INDEX IF EXISTS idx_email_verifications_user_id;
DROP INDEX IF EXISTS idx_email_verifications_token;

-- Drop email_verifications table
DROP TABLE IF EXISTS email_verifications;

-- Remove email_verified column from users table
ALTER TABLE users DROP COLUMN IF EXISTS email_verified;

-- +goose StatementEnd
