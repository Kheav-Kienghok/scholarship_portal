-- +goose Up
-- +goose StatementBegin

-- Create ENUM type for role (idempotent)
DO $$ BEGIN
    CREATE TYPE user_role AS ENUM ('student', 'admin');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Create users table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    fullname VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE,
    password VARCHAR(255),
    role user_role NOT NULL DEFAULT 'student',
    phone_number VARCHAR(20) UNIQUE
        CHECK (
            phone_number ~ '^(\+855(-[0-9]{1,3}){3,4}|(\+855[0-9]{8,12}))$'
        ),
    high_school VARCHAR(100),
    grade_level INTEGER CHECK (grade_level BETWEEN 1 AND 12),
    diploma_year INTEGER,
    diploma_grade VARCHAR(1),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert the two initial admins
INSERT INTO users (fullname, email, password, role) 
VALUES
('Keanghok', 'admin@example.com', '$2a$10$vFO6bl5D83e5miAxbjL4V.f1nj6X1zs.al33oRUgaCuNTMXqDXLYu', 'admin'),
('Tola', 'admin2@example.com', '$2a$10$vFO6bl5D83e5miAxbjL4V.f1nj6X1zs.al33oRUgaCuNTMXqDXLYu', 'admin')
ON CONFLICT (email) DO NOTHING;

-- Trigger function to enforce max 2 admins
CREATE OR REPLACE FUNCTION enforce_max_two_admins()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.role = 'admin' THEN
        IF (SELECT COUNT(*) FROM users WHERE role = 'admin') >= 2 THEN
            RAISE EXCEPTION 'Maximum number of admins reached';
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Attach trigger
CREATE TRIGGER trg_max_two_admins
BEFORE INSERT OR UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION enforce_max_two_admins();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_max_two_admins ON users;
DROP FUNCTION IF EXISTS enforce_max_two_admins;
DROP TABLE IF EXISTS users CASCADE;
DROP TYPE IF EXISTS user_role CASCADE;
-- +goose StatementEnd
