-- name: CreateUser :one
INSERT INTO users (
    fullname, 
    email, 
    password_hash, 
    phone_number, 
    high_school, 
    grade_level, 
    diploma_year
) VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, fullname, email, created_at, updated_at;


-- name: CreateStudentProfile :exec
INSERT INTO student_profiles (student_id)
VALUES ($1);