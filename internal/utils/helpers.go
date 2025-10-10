package utils

import (
	"encoding/json"

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
