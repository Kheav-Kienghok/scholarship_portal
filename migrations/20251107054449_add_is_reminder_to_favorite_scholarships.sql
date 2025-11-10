-- +goose Up
-- +goose StatementBegin
ALTER TABLE favorite_scholarships
ADD COLUMN is_reminder BOOLEAN DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE favorite_scholarships
DROP COLUMN is_reminder;
-- +goose StatementEnd
