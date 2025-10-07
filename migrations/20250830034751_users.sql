-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    fullname VARCHAR(100),

    -- Email validation: case-insensitive, allows normal characters, avoids trailing space issues
    email VARCHAR(100) UNIQUE NOT NULL
        CHECK (
            TRIM(email) ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'
        ),

    password_hash VARCHAR(255),

    -- Phone number: supports +855-XX-XXX-XXX or +855-XX-XXX-XXXX
    phone_number VARCHAR(20) UNIQUE
        CHECK (
            TRIM(phone_number) ~ '^\+855-[1-9][0-9]-[0-9]{3}-[0-9]{3,4}$'
        ),

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users CASCADE;
-- +goose StatementEnd
