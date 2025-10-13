-- name: UpsertOauthLogin :one
INSERT INTO oauth_logins (user_id, provider, provider_user_id, access_token, refresh_token)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (provider, provider_user_id)
DO UPDATE SET
    user_id = EXCLUDED.user_id,
    access_token = EXCLUDED.access_token,
    refresh_token = EXCLUDED.refresh_token,
    updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: GetOAuthLoginByProvider :one
SELECT ol.*, u.* FROM oauth_logins ol
JOIN users u ON ol.user_id = u.id
WHERE ol.provider = $1 AND ol.provider_user_id = $2
LIMIT 1;

-- name: CreateOAuthLogin :one
INSERT INTO oauth_logins (
    user_id,
    provider,
    provider_user_id,
    access_token,
    refresh_token,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, NOW(), NOW()
) RETURNING *;

-- name: UpdateOAuthLogin :one
UPDATE oauth_logins SET
    access_token = $3,
    refresh_token = $4,
    updated_at = NOW()
WHERE user_id = $1 AND provider = $2
RETURNING *;

-- name: GetOAuthLoginsByUserID :many
SELECT * FROM oauth_logins
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: DeleteOAuthLogin :exec
DELETE FROM oauth_logins
WHERE user_id = $1 AND provider = $2;

-- name: GetUserByOAuthProvider :one
SELECT u.* FROM users u
JOIN oauth_logins ol ON u.id = ol.user_id
WHERE ol.provider = $1 AND ol.provider_user_id = $2
LIMIT 1;