-- +goose Up
-- +goose StatementBegin
ALTER TABLE scholarships
ADD COLUMN institution_code TEXT GENERATED ALWAYS AS (
    CASE
        WHEN institution_info->0->>'institution' ~ '\(.*\)' THEN
            -- Extract text inside parentheses
            regexp_replace(institution_info->0->>'institution', '.*\((.*)\).*', '\1')
        ELSE
            -- Take first letter of each word
            regexp_replace(institution_info->0->>'institution', '\b(\w)\w*', '\1', 'g')
    END
) STORED;

CREATE INDEX idx_scholarships_institution_code ON scholarships (institution_code);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_scholarships_institution_code;
ALTER TABLE scholarships
DROP COLUMN IF EXISTS institution_code;
-- +goose StatementEnd
