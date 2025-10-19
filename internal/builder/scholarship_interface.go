package builder

import (
    "encoding/json"
    "time"

    "github.com/Kheav-Kienghok/scholarship_portal/internal/database/db"
)

// ScholarshipSource is an interface for different scholarship data sources
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
    GetCreatedAt() *time.Time
    GetUpdatedAt() *time.Time
}

// ============================================
// ScholarshipWrapper - wraps db.Scholarship
// ============================================
type ScholarshipWrapper struct {
    db.Scholarship
}

func (s ScholarshipWrapper) GetID() int32 {
    return s.ID
}

func (s ScholarshipWrapper) GetTitle() string {
    return s.Title
}

func (s ScholarshipWrapper) GetProvider() string {
    return s.Provider
}

func (s ScholarshipWrapper) GetDescription() string {
    if s.Description.Valid {
        return s.Description.String
    }
    return ""
}

func (s ScholarshipWrapper) GetInstitutionInfo() json.RawMessage {
    if s.InstitutionInfo.Valid {
        return s.InstitutionInfo.RawMessage
    }
    return nil
}

func (s ScholarshipWrapper) GetRequirements() json.RawMessage {
    if s.Requirements.Valid {
        return s.Requirements.RawMessage
    }
    return nil
}

func (s ScholarshipWrapper) GetExtraNotes() string {
    if s.ExtraNotes.Valid {
        return s.ExtraNotes.String
    }
    return ""
}

func (s ScholarshipWrapper) GetDeadlineEnd() *time.Time {
    if s.DeadlineEnd.Valid {
        return &s.DeadlineEnd.Time
    }
    return nil
}

func (s ScholarshipWrapper) GetOfficialLink() *string {
    if s.OfficialLink.Valid {
        return &s.OfficialLink.String
    }
    return nil
}

func (s ScholarshipWrapper) GetPhotoUrl() *string {
    if s.PhotoUrl.Valid {
        return &s.PhotoUrl.String
    }
    return nil
}

func (s ScholarshipWrapper) GetCreatedAt() *time.Time {
    if s.CreatedAt.Valid {
        return &s.CreatedAt.Time
    }
    return nil
}

func (s ScholarshipWrapper) GetUpdatedAt() *time.Time {
    if s.UpdatedAt.Valid {
        return &s.UpdatedAt.Time
    }
    return nil
}

// ============================================
// AllScholarshipsWrapper - wraps db.GetAllScholarshipsRow
// ============================================
type AllScholarshipsWrapper struct {
    db.GetAllScholarshipsRow
}

func (s AllScholarshipsWrapper) GetID() int32 {
    return s.ID
}

func (s AllScholarshipsWrapper) GetTitle() string {
    return s.Title
}

func (s AllScholarshipsWrapper) GetProvider() string {
    return s.Provider
}

func (s AllScholarshipsWrapper) GetDescription() string {
    if s.Description.Valid {
        return s.Description.String
    }
    return ""
}

func (s AllScholarshipsWrapper) GetInstitutionInfo() json.RawMessage {
    if s.InstitutionInfo.Valid {
        return s.InstitutionInfo.RawMessage
    }
    return nil
}

func (s AllScholarshipsWrapper) GetRequirements() json.RawMessage {
    if s.Requirements.Valid {
        return s.Requirements.RawMessage
    }
    return nil
}

func (s AllScholarshipsWrapper) GetExtraNotes() string {
    if s.ExtraNotes.Valid {
        return s.ExtraNotes.String
    }
    return ""
}

func (s AllScholarshipsWrapper) GetDeadlineEnd() *time.Time {
    if s.DeadlineEnd.Valid {
        return &s.DeadlineEnd.Time
    }
    return nil
}

func (s AllScholarshipsWrapper) GetOfficialLink() *string {
    if s.OfficialLink.Valid {
        return &s.OfficialLink.String
    }
    return nil
}

func (s AllScholarshipsWrapper) GetPhotoUrl() *string {
    if s.PhotoUrl.Valid {
        return &s.PhotoUrl.String
    }
    return nil
}

func (s AllScholarshipsWrapper) GetCreatedAt() *time.Time {
    if s.CreatedAt.Valid {
        return &s.CreatedAt.Time
    }
    return nil
}

func (s AllScholarshipsWrapper) GetUpdatedAt() *time.Time {
    return nil
}

// ============================================
// ActiveScholarshipWrapper - wraps db.GetActiveScholarshipsRow
// ============================================
type ActiveScholarshipWrapper struct {
    db.GetActiveScholarshipsRow
}

func (s ActiveScholarshipWrapper) GetID() int32 {
    return s.ID
}

func (s ActiveScholarshipWrapper) GetTitle() string {
    return s.Title
}

func (s ActiveScholarshipWrapper) GetProvider() string {
    return s.Provider
}

func (s ActiveScholarshipWrapper) GetDescription() string {
    if s.Description.Valid {
        return s.Description.String
    }
    return ""
}

func (s ActiveScholarshipWrapper) GetInstitutionInfo() json.RawMessage {
    if s.InstitutionInfo.Valid {
        return s.InstitutionInfo.RawMessage
    }
    return nil
}

func (s ActiveScholarshipWrapper) GetRequirements() json.RawMessage {
    if s.Requirements.Valid {
        return s.Requirements.RawMessage
    }
    return nil
}

func (s ActiveScholarshipWrapper) GetExtraNotes() string {
    if s.ExtraNotes.Valid {
        return s.ExtraNotes.String
    }
    return ""
}

func (s ActiveScholarshipWrapper) GetDeadlineEnd() *time.Time {
    if s.DeadlineEnd.Valid {
        return &s.DeadlineEnd.Time
    }
    return nil
}

func (s ActiveScholarshipWrapper) GetOfficialLink() *string {
    if s.OfficialLink.Valid {
        return &s.OfficialLink.String
    }
    return nil
}

func (s ActiveScholarshipWrapper) GetPhotoUrl() *string {
    if s.PhotoUrl.Valid {
        return &s.PhotoUrl.String
    }
    return nil
}

func (s ActiveScholarshipWrapper) GetCreatedAt() *time.Time {
    if s.CreatedAt.Valid {
        return &s.CreatedAt.Time
    }
    return nil
}

func (s ActiveScholarshipWrapper) GetUpdatedAt() *time.Time {
    // GetActiveScholarshipsRow does not include an UpdatedAt field
    return nil
}

// ============================================
// SearchScholarshipWrapper - wraps db.SearchScholarshipsByProgramsRow
// ============================================
type SearchScholarshipWrapper struct {
    db.SearchScholarshipsByProgramsRow
}

func (s SearchScholarshipWrapper) GetID() int32 {
    return s.ID
}

func (s SearchScholarshipWrapper) GetTitle() string {
    return s.Title
}

func (s SearchScholarshipWrapper) GetProvider() string {
    return s.Provider
}

func (s SearchScholarshipWrapper) GetDescription() string {
    if s.Description.Valid {
        return s.Description.String
    }
    return ""
}

func (s SearchScholarshipWrapper) GetInstitutionInfo() json.RawMessage {
    // SearchScholarshipsByProgramsRow does not include an InstitutionInfo field
    return nil
}

func (s SearchScholarshipWrapper) GetRequirements() json.RawMessage {
    if s.Requirements.Valid {
        return s.Requirements.RawMessage
    }
    return nil
}

func (s SearchScholarshipWrapper) GetExtraNotes() string {
    if s.ExtraNotes.Valid {
        return s.ExtraNotes.String
    }
    return ""
}

func (s SearchScholarshipWrapper) GetDeadlineEnd() *time.Time {
    if s.DeadlineEnd.Valid {
        return &s.DeadlineEnd.Time
    }
    return nil
}

func (s SearchScholarshipWrapper) GetOfficialLink() *string {
    if s.OfficialLink.Valid {
        return &s.OfficialLink.String
    }
    return nil
}

func (s SearchScholarshipWrapper) GetPhotoUrl() *string {
    if s.PhotoUrl.Valid {
        return &s.PhotoUrl.String
    }
    return nil
}

func (s SearchScholarshipWrapper) GetCreatedAt() *time.Time {
    if s.CreatedAt.Valid {
        return &s.CreatedAt.Time
    }
    return nil
}

func (s SearchScholarshipWrapper) GetUpdatedAt() *time.Time {
    // SearchScholarshipsByProgramsRow does not include an UpdatedAt field
    return nil
}