package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/logging"
	"github.com/robfig/cron/v3"
)

// ReminderRequest is the payload for sending reminder emails
type ReminderRequest struct {
	FullName        string `json:"name"`
	Email           string `json:"email"`
	ScholarshipName string `json:"scholarship_name"`
	Description     string `json:"description"`
	Deadline        string `json:"deadline"`
	ApplyLink       string `json:"apply_link"`
}

// ReminderStore is an interface for fetching pending reminders from DB
type ReminderStore interface {
	GetPendingReminders(ctx context.Context) ([]ReminderRequest, error)
}

// SendReminderEmail sends a reminder email via API Gateway
func SendReminderEmail(ctx context.Context, reminder ReminderRequest) error {
	apiURL := os.Getenv("REMINDER_EMAIL_API_URL")
	apiKey := os.Getenv("REMINDER_EMAIL_API_KEY")
	if apiURL == "" {
		return fmt.Errorf("missing REMINDER_EMAIL_API_URL")
	}

	// Wrap single reminder in `data` array for Lambda
	payload := map[string][]ReminderRequest{"data": {reminder}}
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	var lastErr error
	backoff := initialBackoff

	for attempt := 1; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(jsonBody))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		if apiKey != "" {
			req.Header.Set("x-api-key", apiKey)
		}

		resp, err := HTTPClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("attempt %d failed: %w", attempt, err)
			logging.Warn(fmt.Sprintf("Reminder request failed (attempt %d/%d): %v", attempt, maxRetries, err))
			if attempt < maxRetries {
				time.Sleep(backoff)
				backoff *= 2
				continue
			}
			return lastErr
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			logging.Info(fmt.Sprintf("✅ Reminder email sent to %s", reminder.Email))
			logging.Debug(fmt.Sprintf("Reminder API response: %s", string(body)))
			return nil
		}

		if resp.StatusCode >= 500 && attempt < maxRetries {
			lastErr = fmt.Errorf("attempt %d failed with status %d: %s", attempt, resp.StatusCode, string(body))
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		lastErr = fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
		break
	}

	return lastErr
}

func StartDailyEmailCheck(ctx context.Context, store ReminderStore, schedule string) {
	if schedule == "" {
		schedule = "* * * * *" // every minute for testing
	}

	c := cron.New(cron.WithLocation(time.FixedZone("Asia/Phnom_Penh", 7*60*60)))

	_, err := c.AddFunc(schedule, func() {
		start := time.Now()
		logging.Info("📅 Running reminder cron job...")

		reminders, err := store.GetPendingReminders(ctx)
		if err != nil {
			logging.Error("Failed to fetch reminders:", err)
			return
		}

		if len(reminders) == 0 {
			logging.Info("No pending reminders found")
			return
		}

		for _, r := range reminders {
			if err := SendReminderEmail(ctx, r); err != nil {
				logging.Error(fmt.Sprintf("❌ Failed to send reminder to %s: %v", r.Email, err))
				continue
			}
			logging.Info(fmt.Sprintf("✅ Reminder sent to: %s", r.Email))
		}

		logging.Info(fmt.Sprintf("📨 Cron finished in %s", time.Since(start)))
	})

	if err != nil {
		logging.Error("Failed to schedule cron job:", err)
		return
	}

	c.Start()
	logging.Info(fmt.Sprintf("✅ Reminder cron started with schedule: %s", schedule))

	// graceful shutdown
	go func() {
		<-ctx.Done()
		logging.Info("🛑 Stopping reminder cron...")
		c.Stop()
	}()
}
