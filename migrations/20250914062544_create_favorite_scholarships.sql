-- +goose Up
-- +goose StatementBegin
CREATE TABLE favorite_scholarships (
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    scholarship_id BIGINT REFERENCES scholarships(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_id, scholarship_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE favorite_scholarships;
-- +goose StatementEnd

