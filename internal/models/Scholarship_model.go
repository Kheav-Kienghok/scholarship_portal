package models

import (
	"encoding/json"
	"fmt"
	"time"
)

type DateOnly time.Time

func (d DateOnly) MarshalJSON() ([]byte, error) {
	t := time.Time(d)
	formatted := fmt.Sprintf("\"%s\"", t.UTC().Format("2006-01-02"))
	return []byte(formatted), nil
}

type Institution struct {
	Institution     string   `json:"institution"`
	ProgramsOffered []string `json:"programs_offered,omitempty"` // omit empty
}

type ScholarshipSearchResponse struct {
	ID              int           `json:"id"`
	Title           string        `json:"title"`
	Provider        string        `json:"provider"`
	Description     string        `json:"description"`
	InstitutionInfo []Institution `json:"institution_info,omitempty"`
	Requirements    interface{}   `json:"requirements"`
	ExtraNotes      string        `json:"extra_notes"`
	DeadlineEnd     *DateOnly     `json:"deadline_end"`
	OfficialLink    string        `json:"official_link"`
	Categories      []string      `json:"categories"`
	PhotoURL        string        `json:"photo_url"`
	CreatedAt       DateOnly      `json:"created_at"`
}

type CreateScholarshipRequest struct {
	Title           string          `form:"title" json:"title"`
	Provider        string          `form:"provider" json:"provider"`
	Description     *string         `form:"description" json:"description"`
	InstitutionInfo json.RawMessage `form:"institution_info" json:"institution_info"`
	Requirements    json.RawMessage `form:"requirements" json:"requirements"`
	ExtraNotes      *string         `form:"extra_notes" json:"extra_notes"`
	DeadlineEnd     *string         `form:"deadline_end" json:"deadline_end" required:"true"`
	OfficialLink    *string         `form:"official_link" json:"official_link"`
	Categories      json.RawMessage `form:"categories" json:"categories"`
	PhotoURL        *string         `form:"photo_url" json:"photo_url"`
}

type ScholarshipResponse struct {
	ID              int             `json:"id"`
	Title           string          `json:"title"`
	Provider        string          `json:"provider"`
	Description     string          `json:"description,omitempty"`
	InstitutionInfo json.RawMessage `json:"institution_info,omitempty"`
	Requirements    json.RawMessage `json:"requirements,omitempty"`
	ExtraNotes      string          `json:"extra_notes,omitempty"`
	OfficialLink    *string         `json:"official_link,omitempty"`
	DeadlineEnd     *DateOnly       `json:"deadline_end,omitempty"`
	PhotoURL        *string         `json:"photo_url,omitempty"`
	Categories      json.RawMessage `json:"categories,omitempty"`
	CreatedAt       DateOnly        `json:"created_at"`
	UpdatedAt       *DateOnly       `json:"updated_at,omitempty"`
}

// Add these to your models package
type UpdateScholarshipRequest struct {
	Title           string          `json:"title,omitempty"`
	Provider        string          `json:"provider,omitempty"`
	Description     *string         `json:"description,omitempty"`
	InstitutionInfo json.RawMessage `json:"institution_info,omitempty"`
	Requirements    json.RawMessage `json:"requirements,omitempty"`
	ExtraNotes      *string         `json:"extra_notes,omitempty"`
	OfficialLink    *string         `json:"official_link,omitempty"`
	PhotoURL        *string         `json:"photo_url,omitempty"`
	DeadlineEnd     *string         `json:"deadline_end,omitempty"`
}

type ReminderData struct {
    Name            string `json:"name"`
    Email           string `json:"email"`
    ScholarshipName string `json:"scholarship_name"`
    Deadline        string `json:"deadline"`
    ApplyLink       string `json:"apply_link"`
}