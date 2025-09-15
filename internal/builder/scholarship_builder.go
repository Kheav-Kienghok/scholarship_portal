package builder

import (
	"encoding/json"
	"time"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/database/db"
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
		DeadlineEnd:     deadline,
		OfficialLink:    utils.NullStringToPtr(row.OfficialLink),
		PhotoURL:        utils.NullStringToPtr(row.PhotoUrl),
		CreatedAt: func() time.Time {
			if createdAt != nil {
				return *createdAt
			}
			return time.Time{}
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
			DeadlineEnd:     utils.NullTimeToPtr(s.DeadlineEnd),
			OfficialLink:    utils.NullStringToPtr(s.OfficialLink),
			CreatedAt:       *utils.NullTimeToPtr(s.CreatedAt),
		}

		// Generate presigned URL if PhotoUrl exists
		if s.PhotoUrl.Valid && s.PhotoUrl.String != "" {
			if url, err := utils.GenerateScholarshipLogoURL(bucketName, s.PhotoUrl.String, s3Client); err == nil {
				sr.PhotoURL = &url
			}
		}
		responses[i] = sr
	}

	return responses
}
