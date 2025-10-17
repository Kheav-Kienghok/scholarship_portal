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