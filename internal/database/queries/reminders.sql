-- name: GetPendingReminders :many
SELECT id, name, email, scholarship_name, deadline, apply_link, reminder_sent
FROM reminders
WHERE reminder_sent = false
  AND deadline <= (NOW() + INTERVAL '7 days');

-- name: MarkReminderSent :exec
UPDATE reminders
SET reminder_sent = true
WHERE id = $1;
