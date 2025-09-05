-- name: CreateUser :one
INSERT INTO users (fullname, email, password_hash, phone_number, high_school, grade_level, diploma_year)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, fullname, email, created_at, updated_at;

-- name: FindUserByEmail :one
SELECT id, fullname, email, password_hash, role, created_at, updated_at
FROM users
WHERE email = $1;