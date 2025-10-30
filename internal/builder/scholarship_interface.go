package builder

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/database/db"
)

// ScholarshipSource defines a generic scholarship interface
type ScholarshipSource interface {
	GetID() int32
	GetTitle() string
	GetProvider() string
	GetDescription() string
	GetInstitutionInfo() json.RawMessage
	GetRequirements() json.RawMessage
	GetExtraNotes() string
	GetDeadlineEnd() *time.Time
	GetOfficialLink() *string
	GetPhotoUrl() *string
	GetCategories() json.RawMessage
	GetCreatedAt() *time.Time
	GetUpdatedAt() *time.Time
}

// =====================================================
// baseScholarshipWrapper - shared logic for all wrappers
// =====================================================
type baseScholarshipWrapper struct {
	ID              int32
	Title           string
	Provider        string
	Description     nullableString
	InstitutionInfo nullableJSON
	Requirements    nullableJSON
	ExtraNotes      nullableString
	DeadlineEnd     nullableTime
	OfficialLink    nullableString
	PhotoUrl        nullableString
	Categories      nullableJSON
	CreatedAt       nullableTime
	UpdatedAt       nullableTime
}

// --- Null Type Wrappers ---

type nullableString struct {
	Valid  bool
	String string
}

type nullableTime struct {
	Valid bool
	Time  time.Time
}

type nullableJSON struct {
	Valid      bool
	RawMessage json.RawMessage
}

// --- Converters from sql.Null* types ---

func fromNullString(ns sql.NullString) nullableString {
	return nullableString{
		Valid:  ns.Valid,
		String: ns.String,
	}
}

func fromNullTime(nt sql.NullTime) nullableTime {
	return nullableTime{
		Valid: nt.Valid,
		Time:  nt.Time,
	}
}

// You probably used a custom type for JSON columns (like pgtype.JSONB or sql.NullString).
// Assuming you stored JSON as json.RawMessage in a nullable column, handle it generically:
func fromNullableJSON(valid bool, raw json.RawMessage) nullableJSON {
	return nullableJSON{
		Valid:      valid,
		RawMessage: raw,
	}
}

// --------------------------
// Interface implementations
// --------------------------
func (s baseScholarshipWrapper) GetID() int32        { return s.ID }
func (s baseScholarshipWrapper) GetTitle() string    { return s.Title }
func (s baseScholarshipWrapper) GetProvider() string { return s.Provider }
func (s baseScholarshipWrapper) GetDescription() string {
	if s.Description.Valid {
		return s.Description.String
	}
	return ""
}
func (s baseScholarshipWrapper) GetInstitutionInfo() json.RawMessage {
	if s.InstitutionInfo.Valid {
		return s.InstitutionInfo.RawMessage
	}
	return nil
}
func (s baseScholarshipWrapper) GetRequirements() json.RawMessage {
	if s.Requirements.Valid {
		return s.Requirements.RawMessage
	}
	return nil
}
func (s baseScholarshipWrapper) GetExtraNotes() string {
	if s.ExtraNotes.Valid {
		return s.ExtraNotes.String
	}
	return ""
}
func (s baseScholarshipWrapper) GetDeadlineEnd() *time.Time {
	if s.DeadlineEnd.Valid {
		return &s.DeadlineEnd.Time
	}
	return nil
}
func (s baseScholarshipWrapper) GetOfficialLink() *string {
	if s.OfficialLink.Valid {
		return &s.OfficialLink.String
	}
	return nil
}
func (s baseScholarshipWrapper) GetPhotoUrl() *string {
	if s.PhotoUrl.Valid {
		return &s.PhotoUrl.String
	}
	return nil
}
func (s baseScholarshipWrapper) GetCategories() json.RawMessage {
	if s.Categories.Valid {
		return s.Categories.RawMessage
	}
	return nil
}
func (s baseScholarshipWrapper) GetCreatedAt() *time.Time {
	if s.CreatedAt.Valid {
		return &s.CreatedAt.Time
	}
	return nil
}
func (s baseScholarshipWrapper) GetUpdatedAt() *time.Time {
	if s.UpdatedAt.Valid {
		return &s.UpdatedAt.Time
	}
	return nil
}

// =====================================================
// Constructors for each wrapper type
// =====================================================

func NewScholarshipWrapper(s db.Scholarship) ScholarshipSource {
	return baseScholarshipWrapper{
		ID:              s.ID,
		Title:           s.Title,
		Provider:        s.Provider,
		Description:     fromNullString(s.Description),
		InstitutionInfo: fromNullableJSON(s.InstitutionInfo.Valid, s.InstitutionInfo.RawMessage),
		Requirements:    fromNullableJSON(s.Requirements.Valid, s.Requirements.RawMessage),
		ExtraNotes:      fromNullString(s.ExtraNotes),
		DeadlineEnd:     fromNullTime(s.DeadlineEnd),
		OfficialLink:    fromNullString(s.OfficialLink),
		PhotoUrl:        fromNullString(s.PhotoUrl),
		Categories:      fromNullableJSON(s.Categories.Valid, s.Categories.RawMessage),
		CreatedAt:       fromNullTime(s.CreatedAt),
		UpdatedAt:       fromNullTime(s.UpdatedAt),
	}
}

func NewAllScholarshipsWrapper(s db.GetAllScholarshipsRow) ScholarshipSource {
	return baseScholarshipWrapper{
		ID:              s.ID,
		Title:           s.Title,
		Provider:        s.Provider,
		Description:     fromNullString(s.Description),
		InstitutionInfo: fromNullableJSON(s.InstitutionInfo.Valid, s.InstitutionInfo.RawMessage),
		Requirements:    fromNullableJSON(s.Requirements.Valid, s.Requirements.RawMessage),
		ExtraNotes:      fromNullString(s.ExtraNotes),
		DeadlineEnd:     fromNullTime(s.DeadlineEnd),
		OfficialLink:    fromNullString(s.OfficialLink),
		Categories:      fromNullableJSON(s.Categories.Valid, s.Categories.RawMessage),
		PhotoUrl:        fromNullString(s.PhotoUrl),
		CreatedAt:       fromNullTime(s.CreatedAt),
	}
}

func NewActiveScholarshipWrapper(s db.GetActiveScholarshipsRow) ScholarshipSource {
	return baseScholarshipWrapper{
		ID:              s.ID,
		Title:           s.Title,
		Provider:        s.Provider,
		Description:     fromNullString(s.Description),
		InstitutionInfo: fromNullableJSON(s.InstitutionInfo.Valid, s.InstitutionInfo.RawMessage),
		Requirements:    fromNullableJSON(s.Requirements.Valid, s.Requirements.RawMessage),
		ExtraNotes:      fromNullString(s.ExtraNotes),
		DeadlineEnd:     fromNullTime(s.DeadlineEnd),
		OfficialLink:    fromNullString(s.OfficialLink),
		PhotoUrl:        fromNullString(s.PhotoUrl),
		CreatedAt:       fromNullTime(s.CreatedAt),
	}
}

func NewSearchScholarshipWrapper(s db.SearchScholarshipsByProgramsRow) ScholarshipSource {
	return baseScholarshipWrapper{
		ID:              s.ID,
		Title:           s.Title,
		Provider:        s.Provider,
		Description:     fromNullString(s.Description),
		InstitutionInfo: nullableJSON{Valid: false, RawMessage: nil}, // Not available in search
		Requirements:    fromNullableJSON(s.Requirements.Valid, s.Requirements.RawMessage),
		ExtraNotes:      fromNullString(s.ExtraNotes),
		DeadlineEnd:     fromNullTime(s.DeadlineEnd),
		OfficialLink:    fromNullString(s.OfficialLink),
		PhotoUrl:        fromNullString(s.PhotoUrl),
		CreatedAt:       fromNullTime(s.CreatedAt),
	}
}

func NewGetScholarshipByIDWrapper(s db.GetScholarshipByIDRow) ScholarshipSource {
	return baseScholarshipWrapper{
		ID:              s.ID,
		Title:           s.Title,
		Provider:        s.Provider,
		Description:     fromNullString(s.Description),
		InstitutionInfo: fromNullableJSON(s.InstitutionInfo.Valid, s.InstitutionInfo.RawMessage),
		Requirements:    fromNullableJSON(s.Requirements.Valid, s.Requirements.RawMessage),
		ExtraNotes:      fromNullString(s.ExtraNotes),
		DeadlineEnd:     fromNullTime(s.DeadlineEnd),
		OfficialLink:    fromNullString(s.OfficialLink),
		PhotoUrl:        fromNullString(s.PhotoUrl),
		Categories:      fromNullableJSON(s.Categories.Valid, s.Categories.RawMessage),
		CreatedAt:       fromNullTime(s.CreatedAt),
	}
}

func NewGetScholarshipByIDsWrapper(s db.Scholarship) ScholarshipSource {
	return NewScholarshipWrapper(s)
}
