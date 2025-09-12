package utils

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func GeneratePresignedURL(bucketName, key string, client *s3.Client) (string, error) {
	presignClient := s3.NewPresignClient(client)

	params := &s3.GetObjectInput{
		Bucket:                     &bucketName,
		Key:                        &key,
		ResponseContentDisposition: aws.String("inline"),    // <-- This forces browser to display
		ResponseContentType:        aws.String("image/png"), // optional but good
	}

	// The URL expires after 12 hours
	presignedReq, err := presignClient.PresignGetObject(
		context.TODO(),
		params,
		s3.WithPresignExpires(12*time.Hour),
	)
	if err != nil {
		return "", err
	}

	return presignedReq.URL, nil
}

func SanitizeString(s string) string {
	s = strings.TrimSpace(s)

	re := regexp.MustCompile(`\s+`)
	s = re.ReplaceAllString(s, " ")

	reAllowed := regexp.MustCompile(`[^a-zA-Z0-9 _-]`)
	s = reAllowed.ReplaceAllString(s, "")

	return s
}
