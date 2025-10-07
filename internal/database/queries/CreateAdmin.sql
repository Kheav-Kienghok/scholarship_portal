-- name: GetAdminByIDOrEmail :one
SELECT 
    id,
    fullname,
    email,
    password_hash,
    totp_secret,
    is_two_factor,
    created_at,
    updated_at
FROM admins
WHERE ($1::bigint IS NOT NULL AND id = $1) OR ($2::text IS NOT NULL AND email = $2);

-- name: GetAdminByEmail :one
SELECT id, fullname, email, password_hash, totp_secret, is_two_factor, created_at, updated_at
FROM admins 
WHERE email = $1;

-- name: GetAdminByID :one
SELECT id, fullname, email, password_hash, totp_secret, is_two_factor, created_at, updated_at
FROM admins 
WHERE id = $1;