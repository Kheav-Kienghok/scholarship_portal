-- +goose Up
-- +goose StatementBegin
CREATE TYPE diploma_grade AS ENUM ('A', 'B', 'C', 'D', 'F');

CREATE TABLE student_profiles (
    student_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    diploma_grade diploma_grade,
    select_majors JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (student_id)
);

-- Auto-update updated_at on row changes
CREATE OR REPLACE FUNCTION update_student_profiles_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_update_student_profiles_updated_at
BEFORE UPDATE ON student_profiles
FOR EACH ROW
EXECUTE FUNCTION update_student_profiles_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_update_student_profiles_updated_at ON student_profiles;
DROP FUNCTION IF EXISTS update_student_profiles_updated_at;
DROP TABLE student_profiles;
DROP TYPE diploma_grade;
-- +goose StatementEnd
