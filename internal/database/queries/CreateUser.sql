-- name: CreateUser :one
INSERT INTO users (
    fullname, 
    email,
    password_hash
) VALUES ($1, $2, $3)
RETURNING id, fullname, email, created_at, updated_at;

-- name: GetUserByIDOrEmail :one
SELECT 
    id,
    fullname,
    email,
    password_hash,
    created_at,
    updated_at
FROM users
WHERE ($1::int IS NOT NULL AND id = $1) OR ($2::text IS NOT NULL AND email = $2);

-- name: GetUserByEmail :one
SELECT id, fullname, email, password_hash, created_at, updated_at
FROM users 
WHERE email = $1;

-- name: GetUserByID :one
SELECT id, fullname, email, password_hash, created_at, updated_at
FROM users 
WHERE id = $1;