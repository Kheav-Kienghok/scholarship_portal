-- name: UpdateUserProfile :one
UPDATE users
SET fullname = COALESCE(sqlc.narg('fullname'), fullname),
    phone_number = COALESCE(sqlc.narg('phone_number'), phone_number)
WHERE id = sqlc.arg('id')
RETURNING *;