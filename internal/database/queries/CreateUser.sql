-- name: CreateUser :one
INSERT INTO users (
    fullname, 
    email,
    password_hash,
    email_verified
) VALUES ($1, $2, $3, $4)
RETURNING id, fullname, email, email_verified, created_at, updated_at;

-- name: GetUserByIDOrEmail :one
SELECT 
    id,
    fullname,
    email,
    password_hash,
    email_verified,
    created_at,
    updated_at
FROM users
WHERE ($1::int IS NOT NULL AND id = $1) OR ($2::text IS NOT NULL AND email = $2);

-- name: GetUserByEmail :one
SELECT id, fullname, email, password_hash, email_verified, created_at, updated_at
FROM users 
WHERE email = $1;

-- name: CheckUserExistByEmail :one
SELECT email, email_verified 
FROM users 
WHERE email = $1;

-- name: GetUserByID :one
SELECT id, fullname, email, password_hash, email_verified, created_at, updated_at
FROM users 
WHERE id = $1;

-- name: VerifyUserEmail :exec
UPDATE users
SET email_verified = true, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;