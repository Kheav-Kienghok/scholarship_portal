-- name: UpdateUserProfile :one
UPDATE users
SET fullname = COALESCE(sqlc.narg('fullname'), fullname)
WHERE id = sqlc.arg('id')
RETURNING *;