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
)

// VerificationRequest represents the payload for the email verification API
type VerificationRequest struct {
	Email      string `json:"email"`
	Name       string `json:"name"`
	VerifyLink string `json:"verify_link"`
}

// HTTPClient is a reusable HTTP client with a timeout
var HTTPClient = &http.Client{Timeout: 30 * time.Second}

const (
	maxRetries     = 3
	initialBackoff = 500 * time.Millisecond
)

// SendVerificationEmail triggers your API Gateway endpoint to send a verification email
func SendVerificationEmail(ctx context.Context, email, name, verifyLink string) error {
	// Load environment variables
	apiURL, _, apiKey, err := getEnvVars()
	if err != nil {
		logging.Error("Environment variables missing:", err)
		return err
	}

	logging.Info(fmt.Sprintf("Attempting to send email to %s via %s", email, apiURL))

	// Build the JSON payload
	payload := VerificationRequest{
		Email:      email,
		Name:       name,
		VerifyLink: verifyLink,
	}
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Retry logic with exponential backoff
	var lastErr error
	backoff := initialBackoff

	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Create HTTP request with context
		req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(jsonBody))
		if err != nil {
			return fmt.Errorf("failed to create HTTP request: %w", err)
		}

		// Set headers BEFORE signing (API key must be added AFTER signing)
		req.Header.Set("Content-Type", "application/json")

		// Add API key AFTER signing (not part of the signature)
		if apiKey != "" {
			req.Header.Set("x-api-key", apiKey)
		}

		// Send the HTTP request
		resp, err := HTTPClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("attempt %d failed: %w", attempt, err)
			logging.Warn(fmt.Sprintf("Email API request failed (attempt %d/%d): %v", attempt, maxRetries, err))

			// Retry on network errors
			if attempt < maxRetries {
				time.Sleep(backoff)
				backoff *= 2 // Exponential backoff
				continue
			}
			return lastErr
		}

		// Read response body for logging
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		// Accept all 2xx status codes as success
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			logging.Info(fmt.Sprintf("Email verification sent successfully to %s (status: %d)", email, resp.StatusCode))

			// Log response for debugging
			var respData map[string]interface{}
			if err := json.Unmarshal(respBody, &respData); err == nil {
				logging.Debug(fmt.Sprintf("Email API response: %v", respData))
			}

			return nil
		}

		// Handle 5xx errors with retry
		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("attempt %d failed with status %d: %s", attempt, resp.StatusCode, string(respBody))
			logging.Warn(fmt.Sprintf("Email API returned 5xx error (attempt %d/%d): %s", attempt, maxRetries, string(respBody)))

			if attempt < maxRetries {
				time.Sleep(backoff)
				backoff *= 2
				continue
			}
			return lastErr
		}

		// 4xx errors are client errors - don't retry
		logging.Error(fmt.Sprintf("Email API returned 4xx error: %s", string(respBody)))
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return lastErr
}

// getEnvVars validates and retrieves required environment variables
func getEnvVars() (string, string, string, error) {
	apiURL := os.Getenv("EMAIL_VERIFICATION_API_URL")
	region := os.Getenv("AWS_REGION")
	apiKey := os.Getenv("EMAIL_API_KEY") // Optional

	if apiURL == "" {
		return "", "", "", fmt.Errorf("missing required environment variable: EMAIL_VERIFICATION_API_URL")
	}
	if region == "" {
		return "", "", "", fmt.Errorf("missing required environment variable: AWS_REGION")
	}

	return apiURL, region, apiKey, nil
}
