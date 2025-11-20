-- Add a favorite
-- name: AddFavorite :exec
INSERT INTO favorite_scholarships (user_id, scholarship_id)
VALUES ($1, $2)
ON CONFLICT (user_id, scholarship_id) DO NOTHING;

-- Remove a favorite
-- name: RemoveFavorite :exec
DELETE FROM favorite_scholarships
WHERE user_id = $1 AND scholarship_id = $2;

-- List all favorites for a user (joined with scholarships)
-- name: ListFavoritesByUser :many
SELECT 
    s.id AS scholarship_id,
    s.title,
    s.provider,
    s.description,
    s.institution_info,
    s.requirements,
    s.extra_notes,
    s.deadline_end,
    s.official_link,
    s.categories,
    s.photo_url,
    s.created_at
FROM favorite_scholarships f
JOIN scholarships s ON f.scholarship_id = s.id
WHERE f.user_id = $1 AND f.is_favorite = TRUE
ORDER BY f.created_at DESC;

-- name: GetRemindersForToday :many
SELECT 
    u.fullname,
    u.email,
    s.title,
    s.description,
    s.deadline_end,
    s.official_link,
    fs.is_reminder
FROM favorite_scholarships fs
LEFT JOIN users u 
    ON u.id = fs.user_id
LEFT JOIN scholarships s
    ON s.id = fs.scholarship_id;

-- name: UpdateFavoriteStatus :exec
UPDATE favorite_scholarships
SET is_favorite = $1
WHERE user_id = $2
  AND scholarship_id = $3;