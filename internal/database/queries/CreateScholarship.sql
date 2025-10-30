-- name: CreateScholarship :one
INSERT INTO scholarships (
    title,
    provider,
    description,
    institution_info,
    requirements,
    extra_notes,
    deadline_end,
    official_link,
    categories,
    photo_url
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, title, provider;

-- name: CreateScholarshipWithDetails :one
INSERT INTO scholarships (
    title,
    provider,
    description,
    institution_info,
    requirements,
    extra_notes,
    deadline_end,
    official_link,
    categories
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;


-- name: DeleteScholarshipByID :exec
DELETE FROM scholarships
WHERE id = $1;