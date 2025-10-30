-- +goose Up
-- +goose StatementBegin
CREATE TABLE scholarships (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    provider VARCHAR(255) NOT NULL,
    description TEXT,
    institution_info JSONB,
    requirements JSONB,
    extra_notes TEXT,
    deadline_end DATE,
    official_link VARCHAR(255),
    photo_url TEXT,
    categories JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);  
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP TABLE scholarships;
-- +goose StatementEnd
