package utils

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/database/db"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/models"
	"github.com/sqlc-dev/pqtype"
)

func ParseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}

func ToNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}

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

func ToNullRawMessage(m json.RawMessage) pqtype.NullRawMessage {
	if len(m) == 0 {
		return pqtype.NullRawMessage{Valid: false}
	}
	return pqtype.NullRawMessage{RawMessage: m, Valid: true}
}

func MapScholarship(row db.GetAllScholarshipsRow) models.ScholarshipResponse {
	var deadline *time.Time
	if row.DeadlineEnd.Valid {
		deadline = &row.DeadlineEnd.Time
	}

	description := ""
	if row.Description.Valid {
		description = row.Description.String
	}

	extraNotes := ""
	if row.ExtraNotes.Valid {
		extraNotes = row.ExtraNotes.String
	}

	officialLink := ""
	if row.OfficialLink.Valid {
		officialLink = row.OfficialLink.String
	}

	institutionInfo := json.RawMessage(nil)
	if row.InstitutionInfo.Valid {
		institutionInfo = row.InstitutionInfo.RawMessage
	}

	requirements := json.RawMessage(nil)
	if row.Requirements.Valid {
		requirements = row.Requirements.RawMessage
	}

	var createdAt *time.Time
	if row.CreatedAt.Valid {
		createdAt = &row.CreatedAt.Time
	}

	return models.ScholarshipResponse{
		ID:              int(row.ID),
		Title:           row.Title,
		Provider:        row.Provider,
		Description:     description,
		InstitutionInfo: institutionInfo,
		Requirements:    requirements,
		ExtraNotes:      extraNotes,
		DeadlineEnd:     deadline,
		OfficialLink:    officialLink,
		CreatedAt:       func() time.Time {
			if createdAt != nil {
				return *createdAt
			}
			return time.Time{}
		}(),
	}
}
