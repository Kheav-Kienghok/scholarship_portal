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

-- name: GetActiveScholarships :many
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
WHERE deadline_end > CURRENT_DATE
ORDER BY deadline_end ASC;

-- name: GetScholarshipsByIDs :many
SELECT *
FROM scholarships
WHERE id = ANY($1::int[]);


-- name: GetScholarshipsByInstitutionCodeLike :many
SELECT *
FROM scholarships
WHERE institution_code ILIKE '%' || sqlc.arg('code') || '%';

-- name: SearchScholarships :many
SELECT *
FROM scholarships
WHERE (institution_code ILIKE '%' || sqlc.arg('code') || '%' OR sqlc.arg('code') IS NULL)
  AND (institution_info->0->>'institution' ILIKE '%' || sqlc.arg('name') || '%' OR sqlc.arg('name') IS NULL)
  AND (
        EXISTS (
            SELECT 1 FROM jsonb_array_elements(institution_info->0->'programs_offered') AS p
            WHERE p->>0 ILIKE '%' || sqlc.arg('program') || '%'
        ) OR sqlc.arg('program') IS NULL
      );

-- name: UpdateScholarship :one
UPDATE scholarships SET
    title = $2,
    provider = $3,
    description = $4,
    institution_info = $5,
    requirements = $6,
    extra_notes = $7,
    deadline_end = $8,
    official_link = $9,
    photo_url = $10,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateScholarshipJSONB :one
UPDATE scholarships SET
    institution_info = COALESCE($2, institution_info),
    requirements = COALESCE($3, requirements),
    updated_at = NOW()
WHERE id = $1
RETURNING *;