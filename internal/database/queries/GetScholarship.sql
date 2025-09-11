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

