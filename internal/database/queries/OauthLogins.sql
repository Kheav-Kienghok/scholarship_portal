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