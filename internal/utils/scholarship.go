package utils

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sqlc-dev/pqtype"
)

// GetStringOrDefault returns the input string if not empty, otherwise returns the default value
func GetStringOrDefault(input, defaultValue string) string {
	if input == "" {
		return defaultValue
	}
	return input
}

// GetNullStringOrExisting returns a new NullString if input is provided, otherwise returns existing
func GetNullStringOrExisting(input *string, existing sql.NullString) sql.NullString {
	if input != nil {
		return ToNullString(input)
	}
	return existing
}

// GetNullRawMessageOrExisting returns a new NullRawMessage if input is provided, otherwise returns existing
func GetNullRawMessageOrExisting(input interface{}, existing pqtype.NullRawMessage) pqtype.NullRawMessage {
	if input == nil {
		return existing
	}

	// Handle pointers properly
	switch v := input.(type) {
	case *json.RawMessage:
		if v == nil {
			return existing
		}
		return ToNullRawMessage(*v)

	case json.RawMessage:
		if len(v) == 0 {
			return existing
		}
		return ToNullRawMessage(v)

	default:
		return existing
	}
}

// JSONNoEscape sends a JSON response with HTML escaping disabled
func JSONNoEscape(c *gin.Context, status int, message string, data interface{}) {
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.Status(status)

	response := gin.H{
		"success": true,
		"message": message,
		"data":    data,
	}

	encoder := json.NewEncoder(c.Writer)
	encoder.SetEscapeHTML(false) // disable & < > escaping
	encoder.SetIndent("", "  ")  // optional: pretty print
	if err := encoder.Encode(response); err != nil {
		// fallback if encoding fails
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to encode response",
			"data":    nil,
		})
	}
}
