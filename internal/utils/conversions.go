package utils

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/sqlc-dev/pqtype"
)

// ParseDate converts YYYY-MM-DD string into time.Time.
func ParseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}

// ToNullString converts *string into sql.NullString.
func ToNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}

// ToNullTime converts YYYY-MM-DD string into sql.NullTime.
func ToNullTime(s string) sql.NullTime {
	if s == "" {
		return sql.NullTime{Valid: false}
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: t, Valid: true}
}

// ToNullRawMessage converts JSON RawMessage into pqtype.NullRawMessage.
func ToNullRawMessage(m json.RawMessage) pqtype.NullRawMessage {
	if len(m) == 0 {
		return pqtype.NullRawMessage{Valid: false}
	}
	return pqtype.NullRawMessage{RawMessage: m, Valid: true}
}

// ---- Null-safe extractors ----

func NullStringToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func NullStringToPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

func NullRawMessageToJSON(n pqtype.NullRawMessage) json.RawMessage {
	if n.Valid {
		return n.RawMessage
	}
	return json.RawMessage("null")
}

func NullTimeToPtr(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}
