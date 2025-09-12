-- name: AdminUpdateUserTOTPSecret :exec
-- description: Admin-only query to update a user's TOTP secret
UPDATE admins
SET totp_secret = $2
WHERE id = $1;

-- name: EnableAdmin2FA :exec
UPDATE admins
SET is_two_factor = TRUE
WHERE id = $1;