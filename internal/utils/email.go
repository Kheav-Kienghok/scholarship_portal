package utils

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
)

// VerificationRequest represents the payload for the email verification API
type VerificationRequest struct {
	Email      string `json:"email"`
	Name       string `json:"name"`
	VerifyLink string `json:"verify_link"`
}

// HTTPClient is a reusable HTTP client with a timeout
var HTTPClient = &http.Client{Timeout: 10 * time.Second}

// SendVerificationEmail triggers your API Gateway endpoint to send a verification email
func SendVerificationEmail(ctx context.Context, email, name, verifyLink string) error {
	// Load environment variables
	apiURL, region, err := getEnvVars()
	if err != nil {
		return err
	}

	// Load AWS config
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Resolve AWS credentials
	creds, err := cfg.Credentials.Retrieve(ctx)
	if err != nil {
		return fmt.Errorf("failed to retrieve AWS credentials: %w", err)
	}

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

	// Compute SHA256 hash of the payload (required by SigV4)
	hash := sha256.Sum256(jsonBody)
	bodyHash := hex.EncodeToString(hash[:])

	// Create the HTTP POST request
	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Sign the request with AWS SigV4
	signer := v4.NewSigner()
	if err := signer.SignHTTP(ctx, creds, req, bodyHash, "execute-api", region, time.Now()); err != nil {
		return fmt.Errorf("failed to sign HTTP request: %w", err)
	}

	// Send the HTTP request
	resp, err := HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Read and log the response body
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status %s: %s", resp.Status, string(respBody))
	}

	fmt.Println("Email verification request sent successfully")
	return nil
}

// getEnvVars validates and retrieves required environment variables
func getEnvVars() (string, string, error) {
	apiURL := os.Getenv("EMAIL_VERIFICATION_API_URL")
	region := os.Getenv("AWS_REGION")

	if apiURL == "" {
		return "", "", fmt.Errorf("missing required environment variable: EMAIL_VERIFICATION_API_URL")
	}
	if region == "" {
		return "", "", fmt.Errorf("missing required environment variable: AWS_REGION")
	}

	return apiURL, region, nil
}
