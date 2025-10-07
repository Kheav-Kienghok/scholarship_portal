-- name: GetUserWithProfile :one
SELECT 
    u.id,
    u.fullname,
    u.email,
    u.phone_number,
    -- u.high_school,
    -- u.grade_level,
    -- u.diploma_year,
    -- s.diploma_grade,
    s.select_majors,
    s.created_at AS profile_created_at,
    s.updated_at AS profile_updated_at
FROM users u
LEFT JOIN student_profiles s ON s.student_id = u.id
WHERE u.id = $1 OR u.email = $2;