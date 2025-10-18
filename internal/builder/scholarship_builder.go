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

// BuildScholarshipResponse builds a single scholarship response from any source
func BuildScholarshipResponse(
    source ScholarshipSource,
    s3Client *s3.Client,
    bucketName string,
) models.ScholarshipResponse {
    sr := models.ScholarshipResponse{
        ID:              int(source.GetID()),
        Title:           source.GetTitle(),
        Provider:        source.GetProvider(),
        Description:     source.GetDescription(),
        InstitutionInfo: source.GetInstitutionInfo(),
        Requirements:    source.GetRequirements(),
        ExtraNotes:      source.GetExtraNotes(),
        DeadlineEnd: func() *models.DateOnly {
            if t := source.GetDeadlineEnd(); t != nil {
                d := models.DateOnly(*t)
                return &d
            }
            return nil
        }(),
        OfficialLink: source.GetOfficialLink(),
        CreatedAt: func() models.DateOnly {
            if t := source.GetCreatedAt(); t != nil {
                return models.DateOnly(*t)
            }
            return models.DateOnly(time.Time{})
        }(),
        UpdatedAt: func() *models.DateOnly {
            if t := source.GetUpdatedAt(); t != nil {
                d := models.DateOnly(*t)
                return &d
            }
            return nil
        }(),
    }

    // Generate presigned URL if PhotoUrl exists (will use cache automatically)
    if photoUrl := source.GetPhotoUrl(); photoUrl != nil && *photoUrl != "" {
        if url, err := utils.GenerateScholarshipLogoURL(bucketName, *photoUrl, s3Client); err == nil {
            sr.PhotoURL = &url
        } else {
            logging.Error("Failed to generate presigned URL for scholarship", source.GetID(), ":", err)
        }
    }

    return sr
}

// BuildScholarshipResponses builds API responses from GetAllScholarshipsRow
func BuildScholarshipResponses(
    scholarships []db.GetAllScholarshipsRow,
    s3Client *s3.Client,
    bucketName string,
) []models.ScholarshipResponse {
    responses := make([]models.ScholarshipResponse, len(scholarships))

    for i, s := range scholarships {
        wrapper := AllScholarshipsWrapper{s}
        responses[i] = BuildScholarshipResponse(wrapper, s3Client, bucketName)
    }

    return responses
}

// BuildActiveScholarshipResponses builds API responses from active scholarships
func BuildActiveScholarshipResponses(
    scholarships []db.GetActiveScholarshipsRow,
    s3Client *s3.Client,
    bucketName string,
) []models.ScholarshipResponse {
    responses := make([]models.ScholarshipResponse, len(scholarships))

    for i, s := range scholarships {
        wrapper := ActiveScholarshipWrapper{s}
        responses[i] = BuildScholarshipResponse(wrapper, s3Client, bucketName)
    }

    return responses
}

// BuildScholarshipResponsesFromDB builds API responses from db.Scholarship
func BuildScholarshipResponsesFromDB(
    scholarships []db.Scholarship,
    s3Client *s3.Client,
    bucketName string,
) []models.ScholarshipResponse {
    responses := make([]models.ScholarshipResponse, len(scholarships))

    for i, s := range scholarships {
        wrapper := ScholarshipWrapper{s}
        responses[i] = BuildScholarshipResponse(wrapper, s3Client, bucketName)
    }

    return responses
}

// ============================================
// Legacy Functions (for backward compatibility)
// ============================================

// MapScholarship - legacy function for backward compatibility
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

// MapScholarshipFromDBScholarship - legacy function
func MapScholarshipFromDBScholarship(s db.Scholarship) models.ScholarshipResponse {
    wrapper := ScholarshipWrapper{s}
    // Use nil for s3Client and bucketName since this function doesn't generate URLs
    return BuildScholarshipResponse(wrapper, nil, "")
}