-- name: GetUserByIDOrEmail :one
SELECT 
    id,
    fullname,
    email,
    password_hash,
    role,
    phone_number,
    high_school,
    grade_level,
    diploma_year,
    created_at,
    updated_at
FROM users
WHERE id = $1 OR email = $2;
