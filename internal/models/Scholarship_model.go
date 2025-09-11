package models

import (
	"encoding/json"
	"time"
)

type CreateScholarshipRequest struct {
	Title           string          `form:"title" json:"title"`
	Provider        string          `form:"provider" json:"provider"`
	Description     *string         `form:"description" json:"description"`
	InstitutionInfo json.RawMessage `form:"institution_info" json:"institution_info"`
	Requirements    json.RawMessage `form:"requirements" json:"requirements"`
	ExtraNotes      *string         `form:"extra_notes" json:"extra_notes"`
	DeadlineEnd     *string         `form:"deadline_end" json:"deadline_end"`
	OfficialLink    *string         `form:"official_link" json:"official_link"`
	PhotoURL        *string         `form:"photo_url" json:"photo_url"` // new field
}

type ScholarshipResponse struct {
	ID              int             `json:"id"`
	Title           string          `json:"title"`
	Provider        string          `json:"provider"`
	Description     string          `json:"description,omitempty"`
	InstitutionInfo json.RawMessage `json:"institution_info,omitempty"`
	Requirements    json.RawMessage `json:"requirements,omitempty"`
	ExtraNotes      string          `json:"extra_notes,omitempty"`
	DeadlineEnd     *time.Time      `json:"deadline_end,omitempty"`
	OfficialLink    string          `json:"official_link,omitempty"`
	PhotoURL        string          `json:"photo_url,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
}
