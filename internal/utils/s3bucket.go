package utils

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)


// base presign generator with dynamic expiry
func generatePresignedURL(bucketName, key, contentType string, expiry time.Duration, client *s3.Client) (string, error) {
	presignClient := s3.NewPresignClient(client)

	params := &s3.GetObjectInput{
		Bucket:                     &bucketName,
		Key:                        &key,
		ResponseContentDisposition: aws.String("inline"),
		ResponseContentType:        aws.String(contentType),
	}

	presignedReq, err := presignClient.PresignGetObject(
		context.TODO(),
		params,
		s3.WithPresignExpires(expiry),
	)
	if err != nil {
		return "", err
	}

	return presignedReq.URL, nil
}

// For scholarship logos (12h expiry)
func GenerateScholarshipLogoURL(bucketName, key string, client *s3.Client) (string, error) {
	return generatePresignedURL(bucketName, key, "image/png", 12*time.Hour, client)
}

// For institution logos (12h expiry)
func GenerateInstitutionLogoURL(bucketName, key string, client *s3.Client) (string, error) {
	return generatePresignedURL(bucketName, key, "image/png", 12*time.Hour, client)
}

// For QR codes (15 min expiry)
func GenerateQRCodeURL(bucketName, key string, client *s3.Client) (string, error) {
	return generatePresignedURL(bucketName, key, "image/png", 15*time.Minute, client)
}

// Sanitizer (unchanged)
func SanitizeString(s string) string {
	s = strings.TrimSpace(s)

	re := regexp.MustCompile(`\s+`)
	s = re.ReplaceAllString(s, " ")

	reAllowed := regexp.MustCompile(`[^a-zA-Z0-9 _-]`)
	s = reAllowed.ReplaceAllString(s, "")

	return s
}