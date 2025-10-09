-- name: UpdateUserProfile :one
UPDATE users 
SET fullname = $2, phone_number = $3, updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, fullname, email, created_at, updated_at;