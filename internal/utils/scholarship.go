package utils

import (
	"database/sql"
	"encoding/json"

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
