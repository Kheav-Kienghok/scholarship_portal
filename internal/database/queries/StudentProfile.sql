-- name: CreateStudentProfile :one
INSERT INTO student_profiles (
    student_id,
    high_school,
    grade_level,
    diploma_year,
    diploma_grade,
    select_majors
) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, student_id, high_school, grade_level, diploma_year, diploma_grade, select_majors, created_at, updated_at;

-- name: GetStudentProfile :one
SELECT id, student_id, high_school, grade_level, diploma_year, diploma_grade, select_majors, created_at, updated_at
FROM student_profiles
WHERE student_id = $1;

-- name: UpdateStudentProfile :one
UPDATE student_profiles 
SET 
    high_school = $2,
    grade_level = $3,
    diploma_year = $4,
    diploma_grade = $5,
    select_majors = $6,
    updated_at = CURRENT_TIMESTAMP
WHERE student_id = $1
RETURNING id, student_id, high_school, grade_level, diploma_year, diploma_grade, select_majors, created_at, updated_at;

-- name: DeleteStudentProfile :exec
DELETE FROM student_profiles WHERE student_id = $1;

-- name: GetUserWithStudentProfile :one
SELECT 
    u.id as user_id,
    u.fullname,
    u.email,               
    sp.high_school,
    sp.grade_level,
    sp.diploma_year,
    sp.diploma_grade,
    sp.select_majors,
    sp.created_at as profile_created_at,
    sp.updated_at as profile_updated_at
FROM users u
LEFT JOIN student_profiles sp ON u.id = sp.student_id
WHERE u.id = $1;