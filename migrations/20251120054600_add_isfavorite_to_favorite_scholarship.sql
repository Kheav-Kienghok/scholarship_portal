-- +goose Up
-- +goose StatementBegin
ALTER TABLE favorite_scholarships
ADD COLUMN is_favorite BOOLEAN NOT NULL DEFAULT TRUE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE favorite_scholarships
DROP COLUMN is_favorite;
-- +goose StatementEnd