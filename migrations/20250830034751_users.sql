-- +goose Up
-- +goose StatementBegin

DO $$ BEGIN
    CREATE TYPE user_role AS ENUM ('student', 'admin');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    fullname VARCHAR(100),
    email VARCHAR(100) UNIQUE NOT NULL
        CHECK (
            email ~ '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'
        ),
    password_hash VARCHAR(255),
    role user_role DEFAULT 'student',
    phone_number VARCHAR(20) UNIQUE
        CHECK (
            phone_number ~ '^\+855-[1-9][0-9]-[0-9]{3}-[0-9]{3}[0-9]?$'
        ),
    high_school VARCHAR(100),
    grade_level INTEGER CHECK (grade_level BETWEEN 1 AND 12),
    diploma_year INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO users (fullname, email, password_hash, role) 
VALUES
('Keanghok', 'admin@example.com', '$2a$10$vFO6bl5D83e5miAxbjL4V.f1nj6X1zs.al33oRUgaCuNTMXqDXLYu', 'admin'),
('Tola', 'admin2@example.com', '$2a$10$vFO6bl5D83e5miAxbjL4V.f1nj6X1zs.al33oRUgaCuNTMXqDXLYu', 'admin')
ON CONFLICT (email) DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users CASCADE;
DROP TYPE IF EXISTS user_role CASCADE;
-- +goose StatementEnd
