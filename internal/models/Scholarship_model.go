package models

import (
	"encoding/json"
	"time"
)

type CreateScholarshipRequest struct {
	Title           string          `json:"title" binding:"required"`
	Provider        string          `json:"provider" binding:"required"`
	Description     *string         `json:"description"`
	InstitutionInfo json.RawMessage `json:"institution_info"`
	Requirements    json.RawMessage `json:"requirements"`
	ExtraNotes      *string         `json:"extra_notes"`
	DeadlineEnd     *string         `json:"deadline_end" binding:"required"`
	OfficialLink    *string         `json:"official_link"`
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
	CreatedAt       time.Time       `json:"created_at"`
}

// type Scholarship struct {
// 	ID              int64  `json:"id"`
// 	Title           string `json:"title"`
// 	Provider        string `json:"provider"`
// 	Description     string `json:"description"`
// 	InstitutionInfo string `json:"institution_info"`
// 	Requirements    string `json:"requirements"`
// 	ExtraNotes      string `json:"extra_notes"`
// 	DeadlineEnd     string `json:"deadline_end"`
// 	OfficialLink    string `json:"official_link"`
// }
