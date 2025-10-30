package builder

import (
	"time"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/database/db"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/logging"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/models"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/utils"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// ----------------------
// Helpers
// ----------------------

// toDateOnly safely converts *time.Time → *models.DateOnly
func toDateOnly(t *time.Time) *models.DateOnly {
	if t == nil {
		return nil
	}
	d := models.DateOnly(*t)
	return &d
}

// toDateOnlyValue safely converts *time.Time → models.DateOnly (zero if nil)
func toDateOnlyValue(t *time.Time) models.DateOnly {
	if t == nil {
		return models.DateOnly(time.Time{})
	}
	return models.DateOnly(*t)
}

// ----------------------
// Builders
// ----------------------

// BuildScholarshipResponse converts any ScholarshipSource into an API response
func BuildScholarshipResponse(
	source ScholarshipSource,
	s3Client *s3.Client,
	bucketName string,
) models.ScholarshipResponse {
	resp := models.ScholarshipResponse{
		ID:              int(source.GetID()),
		Title:           source.GetTitle(),
		Provider:        source.GetProvider(),
		Description:     source.GetDescription(),
		InstitutionInfo: source.GetInstitutionInfo(),
		Requirements:    source.GetRequirements(),
		ExtraNotes:      source.GetExtraNotes(),
		DeadlineEnd:     toDateOnly(source.GetDeadlineEnd()),
		OfficialLink:    source.GetOfficialLink(),
		Categories:      source.GetCategories(),
		CreatedAt:       toDateOnlyValue(source.GetCreatedAt()),
		UpdatedAt:       toDateOnly(source.GetUpdatedAt()),
	}

	// If PhotoUrl exists, generate presigned S3 URL
	if photoKey := source.GetPhotoUrl(); photoKey != nil && *photoKey != "" {
		url, err := utils.GenerateScholarshipLogoURL(bucketName, *photoKey, s3Client)
		if err != nil {
			logging.Errorf("failed to generate presigned URL for scholarship %d: %v", source.GetID(), err)
		} else {
			resp.PhotoURL = &url
		}
	}

	return resp
}

// BuildScholarshipResponses converts any []ScholarshipSource into []ScholarshipResponse
func BuildScholarshipResponses(
	sources []ScholarshipSource,
	s3Client *s3.Client,
	bucketName string,
) []models.ScholarshipResponse {
	responses := make([]models.ScholarshipResponse, len(sources))
	for i, src := range sources {
		responses[i] = BuildScholarshipResponse(src, s3Client, bucketName)
	}
	return responses
}

// ----------------------
// Type-specific wrappers (optional convenience)
// ----------------------

// BuildAllScholarshipResponses converts []db.GetAllScholarshipsRow into []ScholarshipResponse
func BuildAllScholarshipResponses(
	scholarships []db.GetAllScholarshipsRow,
	s3Client *s3.Client,
	bucketName string,
) []models.ScholarshipResponse {
	sources := make([]ScholarshipSource, len(scholarships))
	for i, s := range scholarships {
		sources[i] = NewAllScholarshipsWrapper(s)
	}
	return BuildScholarshipResponses(sources, s3Client, bucketName)
}

func BuildScholarshipResponsesFromDB(
	scholarships []db.Scholarship,
	s3Client *s3.Client,
	bucketName string,
) []models.ScholarshipResponse {
	wrapped := make([]ScholarshipSource, len(scholarships))
	for i, s := range scholarships {
		wrapped[i] = NewScholarshipWrapper(s)
	}
	return BuildScholarshipResponses(wrapped, s3Client, bucketName)
}

func BuildActiveScholarshipResponsesFromDB(
	scholarships []db.GetActiveScholarshipsRow,
	s3Client *s3.Client,
	bucketName string,
) []models.ScholarshipResponse {
	wrapped := make([]ScholarshipSource, len(scholarships))
	for i, s := range scholarships {
		wrapped[i] = NewActiveScholarshipWrapper(s)
	}
	return BuildScholarshipResponses(wrapped, s3Client, bucketName)
}

func BuildGetScholarshipByIDResponseFromRow(
	scholarship db.GetScholarshipByIDRow,
	s3Client *s3.Client,
	bucketName string,
) models.ScholarshipResponse {
	return BuildScholarshipResponse(NewGetScholarshipByIDWrapper(scholarship), s3Client, bucketName)
}

func BuildSearchScholarshipResponsesFromDB(
	rows []db.SearchScholarshipsByProgramsRow,
	s3Client *s3.Client,
	bucketName string,
) []models.ScholarshipResponse {
	wrapped := make([]ScholarshipSource, len(rows))
	for i, s := range rows {
		wrapped[i] = NewSearchScholarshipWrapper(s)
	}
	return BuildScholarshipResponses(wrapped, s3Client, bucketName)
}
