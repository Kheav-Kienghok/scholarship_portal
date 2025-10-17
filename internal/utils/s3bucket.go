package utils

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/cache"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/logging"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// base presign generator with dynamic expiry
func generatePresignedURL(bucketName, key, contentType string, expiry time.Duration, client *s3.Client) (string, error) {
	presignClient := s3.NewPresignClient(client)

	params := &s3.GetObjectInput{
		Bucket:              &bucketName,
		Key:                 &key,
		ResponseContentType: aws.String(contentType),
		// Remove ResponseContentDisposition or set it properly
		ResponseContentDisposition: aws.String("inline; filename=\"" + key + "\""),
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

// For scholarship logos (24h expiry) with cache
func GenerateScholarshipLogoURL(bucketName, key string, client *s3.Client) (string, error) {
	// Create cache key
	cacheKey := fmt.Sprintf("presigned:scholarship:%s:%s", bucketName, key)

	// Try to get from cache first
	if cachedURL, found := cache.URLCache.Get(cacheKey); found {
		return cachedURL, nil
	}

	// Generate new presigned URL
	url, err := generatePresignedURL(bucketName, key, "image/png", 24*time.Hour, client)
	if err != nil {
		logging.Error("Failed to generate presigned URL for scholarship logo:", key, "-", err)
		return "", err
	}

	// Cache the URL for 23 hours (1 hour buffer)
	cache.URLCache.Set(cacheKey, url, 23*time.Hour)

	return url, nil
}

// For institution logos (12h expiry) with cache
func GenerateInstitutionLogoURL(bucketName, key string, client *s3.Client) (string, error) {
	// Create cache key
	cacheKey := fmt.Sprintf("presigned:institution:%s:%s", bucketName, key)

	// Try to get from cache first
	if cachedURL, found := cache.URLCache.Get(cacheKey); found {
		logging.Info("✅ Cache HIT for institution logo:", key)
		return cachedURL, nil
	}

	logging.Info("❌ Cache MISS for institution logo:", key, "- generating new URL")

	// Generate new presigned URL
	url, err := generatePresignedURL(bucketName, key, "image/png", 12*time.Hour, client)
	if err != nil {
		logging.Error("Failed to generate presigned URL for institution logo:", key, "-", err)
		return "", err
	}

	// Cache the URL for 11 hours (1 hour buffer)
	cache.URLCache.Set(cacheKey, url, 11*time.Hour)

	return url, nil
}

// For QR codes (15 min expiry) - no cache for short-lived URLs
func GenerateQRCodeURL(bucketName, key string, client *s3.Client) (string, error) {
	logging.Info("Generating QR code URL (no cache):", key)
	return generatePresignedURL(bucketName, key, "image/png", 15*time.Minute, client)
}

// InvalidateScholarshipLogoCache invalidates the cache for a specific scholarship logo
func InvalidateScholarshipLogoCache(bucketName, key string) {
	cacheKey := fmt.Sprintf("presigned:scholarship:%s:%s", bucketName, key)
	cache.URLCache.Delete(cacheKey)
	logging.Info("Invalidated cache for scholarship logo:", key)
}

// InvalidateInstitutionLogoCache invalidates the cache for a specific institution logo
func InvalidateInstitutionLogoCache(bucketName, key string) {
	cacheKey := fmt.Sprintf("presigned:institution:%s:%s", bucketName, key)
	cache.URLCache.Delete(cacheKey)
	logging.Info("Invalidated cache for institution logo:", key)
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