package errors

import (
	"fmt"
	"strings"
	"time"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/logging"
	"github.com/gin-gonic/gin"
)

type ErrorType string

const (
	DatabaseError   ErrorType = "database_error"
	ValidationError ErrorType = "validation_error"
	AuthError       ErrorType = "auth_error"
	NotFoundError   ErrorType = "not_found_error"
	InternalError   ErrorType = "internal_error"
)

type AppError struct {
	Type        ErrorType
	Message     string
	InternalErr error
	StatusCode  int
}

func (e *AppError) Error() string {
	return e.Message
}

// SanitizedErrorResponse handles errors safely for production
func SanitizedErrorResponse(c *gin.Context, err error, statusCode int, userMessage string) {
	// Generate unique error ID for tracking
	errorID := generateErrorID()

	// Log full error details internally with error ID
	logging.Error(fmt.Sprintf("Error ID: %s | %v", errorID, err))

	// Determine sanitized message
	sanitizedMsg := sanitizeErrorMessage(err, userMessage)

	// Return clean response to user
	response := gin.H{
		"success":  false,
		"message":  sanitizedMsg,
		"error_id": errorID, // Clients can reference this for support
	}

	c.JSON(statusCode, response)
}

// HandleAppError handles typed application errors
func HandleAppError(c *gin.Context, appErr *AppError) {
	errorID := generateErrorID()

	// Log with context
	logging.Error(fmt.Sprintf("Error ID: %s | Type: %s | Internal: %v",
		errorID, appErr.Type, appErr.InternalErr))

	var userMessage string
	switch appErr.Type {
	case DatabaseError:
		userMessage = "Database service temporarily unavailable. Please try again later."
	case ValidationError:
		userMessage = appErr.Message // Validation errors are safe to show
	case AuthError:
		userMessage = "Authentication failed. Please check your credentials."
	case NotFoundError:
		userMessage = "Requested resource not found."
	default:
		userMessage = "An unexpected error occurred. Please try again later."
	}

	c.JSON(appErr.StatusCode, gin.H{
		"success":  false,
		"message":  userMessage,
		"error_id": errorID,
	})
}

// sanitizeErrorMessage removes sensitive information
func sanitizeErrorMessage(err error, fallbackMessage string) string {
	if err == nil {
		return fallbackMessage
	}

	errStr := strings.ToLower(err.Error())

	// Check for common sensitive error patterns
	sensitivePatterns := []string{
		"broken pipe",
		"connection refused",
		"tcp",
		"dial",
		"network",
		"sql:",
		"database/sql",
		"pq:",
		"postgres",
		"mysql",
		"password",
		"connection reset",
		"timeout",
		"no route to host",
		"permission denied",
	}

	for _, pattern := range sensitivePatterns {
		if strings.Contains(errStr, pattern) {
			return getDatabaseFriendlyMessage(errStr, fallbackMessage)
		}
	}

	// If error doesn't contain sensitive info, might be safe to show
	// But still be cautious - when in doubt, use fallback
	if len(err.Error()) > 100 || strings.Contains(errStr, "panic") {
		return fallbackMessage
	}

	return fallbackMessage // Default to safe message
}

func getDatabaseFriendlyMessage(errStr, fallback string) string {
	if strings.Contains(errStr, "broken pipe") ||
		strings.Contains(errStr, "connection") {
		return "Database connection error. Please try again in a moment."
	}
	if strings.Contains(errStr, "timeout") {
		return "Request timed out. Please try again."
	}
	return fallback
}

func generateErrorID() string {
	// Simple UUID-like error ID for tracking
	return fmt.Sprintf("ERR_%d", time.Now().UnixNano())
}
