-- name: CreateEmailVerification :one
INSERT INTO email_verifications (user_id, token, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetEmailVerificationByToken :one
SELECT * FROM email_verifications
WHERE token = $1 AND verified_at IS NULL;

-- name: MarkEmailVerificationAsUsed :exec
UPDATE email_verifications
SET verified_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: DeleteEmailVerificationsByUserID :exec
DELETE FROM email_verifications
WHERE user_id = $1;

-- name: GetEmailVerificationByUserID :one
SELECT * FROM email_verifications
WHERE user_id = $1 AND verified_at IS NULL
ORDER BY created_at DESC
LIMIT 1;

-- name: CleanupExpiredVerifications :exec
DELETE FROM email_verifications
WHERE expires_at < CURRENT_TIMESTAMP;

-- name: GetLatestVerificationByEmail :one
SELECT ev.* 
FROM email_verifications ev
JOIN users u ON ev.user_id = u.id
WHERE u.email = $1
ORDER BY ev.created_at DESC
LIMIT 1;

-- name: DeleteExpiredVerifications :exec
DELETE FROM email_verifications
WHERE user_id = $1 AND verified_at IS NULL;

-- name: GetUnverifiedUserByEmail :one
SELECT id, email, email_verified, created_at
FROM users
WHERE email = $1 AND email_verified = false;