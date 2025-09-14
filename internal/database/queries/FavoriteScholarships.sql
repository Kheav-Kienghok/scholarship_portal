-- name: AddFavorite :exec
INSERT INTO favorite_scholarships (user_id, scholarship_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemoveFavorite :exec
DELETE FROM favorite_scholarships
WHERE user_id = $1 AND scholarship_id = $2;

-- name: ListFavoritesByUser :many
SELECT s.*
FROM favorite_scholarships f
JOIN scholarships s ON f.scholarship_id = s.id
WHERE f.user_id = $1
ORDER BY f.created_at DESC;
