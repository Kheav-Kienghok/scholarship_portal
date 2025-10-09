-- +goose Up
-- +goose StatementBegin
CREATE TABLE student_profiles (
    id SERIAL PRIMARY KEY,
    student_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    high_school VARCHAR(200),
    grade_level VARCHAR(50),
    diploma_year INTEGER,
    diploma_grade VARCHAR(10),
    select_majors JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(student_id)
);

CREATE INDEX idx_student_profiles_student_id ON student_profiles(student_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE student_profiles;
-- +goose StatementEnd