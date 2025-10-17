package builder

import (
	"encoding/json"
	"time"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/database/db"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/logging"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/models"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/utils"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// MapScholarship maps a DB row into an API response.
func MapScholarship(row db.GetAllScholarshipsRow) models.ScholarshipResponse {
	var deadline *time.Time
	if row.DeadlineEnd.Valid {
		deadline = &row.DeadlineEnd.Time
	}

	description := utils.NullStringToString(row.Description)
	extraNotes := utils.NullStringToString(row.ExtraNotes)

	var institutionInfo, requirements json.RawMessage
	if row.InstitutionInfo.Valid {
		institutionInfo = row.InstitutionInfo.RawMessage
	}
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
		DeadlineEnd: func() *models.DateOnly {
			if deadline != nil {
				d := models.DateOnly(*deadline)
				return &d
			}
			return nil
		}(),
		OfficialLink: utils.NullStringToPtr(row.OfficialLink),
		PhotoURL:     utils.NullStringToPtr(row.PhotoUrl),
		CreatedAt: func() models.DateOnly {
			if createdAt != nil {
				return models.DateOnly(*createdAt)
			}
			return models.DateOnly(time.Time{})
		}(),
	}
}

// BuildScholarshipResponses builds API responses from DB scholarships.
func BuildScholarshipResponses(
	scholarships []db.Scholarship,
	s3Client *s3.Client,
	bucketName string,
) []models.ScholarshipResponse {
	responses := make([]models.ScholarshipResponse, len(scholarships))

	for i, s := range scholarships {
		sr := models.ScholarshipResponse{
			ID:              int(s.ID),
			Title:           s.Title,
			Provider:        s.Provider,
			Description:     utils.NullStringToString(s.Description),
			InstitutionInfo: utils.NullRawMessageToJSON(s.InstitutionInfo),
			Requirements:    utils.NullRawMessageToJSON(s.Requirements),
			ExtraNotes:      utils.NullStringToString(s.ExtraNotes),
			DeadlineEnd: func() *models.DateOnly {
				t := utils.NullTimeToPtr(s.DeadlineEnd)
				if t != nil {
					d := models.DateOnly(*t)
					return &d
				}
				return nil
			}(),
			OfficialLink: utils.NullStringToPtr(s.OfficialLink),
			CreatedAt:    models.DateOnly(*utils.NullTimeToPtr(s.CreatedAt)),
		}

		// Generate presigned URL if PhotoUrl exists (will use cache automatically)
		if s.PhotoUrl.Valid && s.PhotoUrl.String != "" {
			if url, err := utils.GenerateScholarshipLogoURL(bucketName, s.PhotoUrl.String, s3Client); err == nil {
				sr.PhotoURL = &url
			} else {
				logging.Error("Failed to generate presigned URL for scholarship", s.ID, ":", err)
			}
		}
		responses[i] = sr
	}

	return responses
}

func MapScholarshipFromDBScholarship(s db.Scholarship) models.ScholarshipResponse {
	response := models.ScholarshipResponse{
		ID:        int(s.ID),
		Title:     s.Title,
		Provider:  s.Provider,
		CreatedAt: models.DateOnly(s.CreatedAt.Time),
		UpdatedAt: func() *models.DateOnly {
			d := models.DateOnly(s.UpdatedAt.Time)
			return &d
		}(),
	}

	// Handle nullable fields
	if s.Description.Valid {
		response.Description = s.Description.String
	}

	if s.ExtraNotes.Valid {
		response.ExtraNotes = s.ExtraNotes.String
	}

	if s.OfficialLink.Valid {
		response.OfficialLink = &s.OfficialLink.String
	}

	if s.DeadlineEnd.Valid {
		utcDeadline := s.DeadlineEnd.Time.UTC()
		d := models.DateOnly(utcDeadline)
		response.DeadlineEnd = &d
	}

	// Handle JSONB fields (institution_info)
	if s.InstitutionInfo.Valid {
		var institutionInfo json.RawMessage
		if err := json.Unmarshal(s.InstitutionInfo.RawMessage, &institutionInfo); err == nil {
			response.InstitutionInfo = institutionInfo
		}
	}

	// Handle JSONB fields (requirements)
	if s.Requirements.Valid {
		var requirements json.RawMessage
		if err := json.Unmarshal(s.Requirements.RawMessage, &requirements); err == nil {
			response.Requirements = requirements
		}
	}

	// PhotoURL will be set by the caller after generating presigned URL
	response.PhotoURL = nil

	return response
}
