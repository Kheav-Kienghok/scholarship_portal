-- +goose Up
-- +goose StatementBegin
CREATE TABLE admins (
    id SERIAL PRIMARY KEY,
    fullname VARCHAR(100),
    email VARCHAR(100) UNIQUE NOT NULL
        CHECK (
            email ~ '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'
        ),
    password_hash VARCHAR(255) NOT NULL,
    totp_secret VARCHAR(32),
    is_two_factor BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Seed some admins
INSERT INTO admins (fullname, email, password_hash) 
VALUES
('Keanghok', 'admin@example.com', '$2a$10$vFO6bl5D83e5miAxbjL4V.f1nj6X1zs.al33oRUgaCuNTMXqDXLYu'),
('Tola', 'admin2@example.com', '$2a$10$vFO6bl5D83e5miAxbjL4V.f1nj6X1zs.al33oRUgaCuNTMXqDXLYu')
ON CONFLICT (email) DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS admins;
-- +goose StatementEnd
