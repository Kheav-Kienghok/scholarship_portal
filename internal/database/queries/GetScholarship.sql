-- name: GetAllScholarships :many
SELECT 
    id,
    title,
    provider,
    description,
    institution_info,
    requirements,
    extra_notes,
    deadline_end,
    official_link,
    photo_url,
    created_at
FROM scholarships
ORDER BY created_at DESC;

-- name: GetScholarshipByID :one
SELECT 
    id AS scholarship_id,
    title,
    provider,
    description,
    institution_info,
    requirements,
    extra_notes,
    deadline_end,
    official_link,
    photo_url,
    created_at
FROM scholarships
WHERE id = $1;

-- name: GetScholarshipsByIDs :many
SELECT *
FROM scholarships
WHERE id = ANY($1::int[]);