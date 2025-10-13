package utils

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/sqlc-dev/pqtype"
)

func SafeStringDeref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func SafeInt32Deref(i *int32) int32 {
	if i == nil {
		return 0
	}
	return *i
}

func SliceToNullRawMessage(slice []string) pqtype.NullRawMessage {
	if slice == nil {
		return pqtype.NullRawMessage{Valid: false}
	}
	b, _ := json.Marshal(slice)
	return pqtype.NullRawMessage{
		RawMessage: b,
		Valid:      true,
	}
}

func ParseDeadlineEnd(input *string, existing sql.NullTime) sql.NullTime {
	if input != nil && *input != "" {
		t, err := time.Parse("2006-01-02", *input) // parse YYYY-MM-DD
		if err != nil {
			t, err = time.Parse(time.RFC3339, *input) // fallback RFC3339
			if err != nil {
				return existing // leave existing if parse fails
			}
		}
		return sql.NullTime{Time: t, Valid: true}
	}
	return existing
}
