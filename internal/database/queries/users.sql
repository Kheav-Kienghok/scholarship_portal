-- name: CreateUser :one
INSERT INTO users (fullname, email, password_hash, phone_number, high_school, grade_level, diploma_year)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, fullname, email, created_at, updated_at;

-- name: FindUserByEmail :one
SELECT id, fullname, email, password_hash, role, created_at, updated_at
FROM users
WHERE email = $1;

-- name: UpdateUserProfile :one
UPDATE users
SET fullname = $2,
    phone_number = $3,
    high_school = $4,
    grade_level = $5,
    diploma_year = $6,
    updated_at = NOW()
WHERE id = $1
RETURNING id, fullname, email, phone_number, high_school, grade_level, diploma_year, role, created_at, updated_at;