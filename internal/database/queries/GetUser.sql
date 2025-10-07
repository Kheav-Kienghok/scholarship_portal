-- -- name: GetUserByIDOrEmail :one
-- SELECT 
--     id,
--     fullname,
--     email,
--     password_hash,
--     phone_number,
--     high_school,
--     grade_level,
--     diploma_year,
--     created_at,
--     updated_at
-- FROM users
-- WHERE id = $1 OR email = $2;


-- name: GetAdminByIDOrEmail :one
SELECT 
    id,
    fullname,
    email,
    password_hash,
    totp_secret,
    is_two_factor,
    created_at,
    updated_at
FROM admins
WHERE id = $1 OR email = $2;


