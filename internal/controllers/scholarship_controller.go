package controllers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/builder"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/database/db"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/logging"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/models"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/storage"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/utils"
	"github.com/gin-gonic/gin"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// ScholarshipController handles scholarship-related requests
type ScholarshipController struct {
	Queries *db.Queries
}

func ScholarshipControllerHandler(queries *db.Queries) *ScholarshipController {
	return &ScholarshipController{
		Queries: queries,
	}
}

// CreateScholarship godoc
// @Summary Create a new scholarship
// @Tags Scholarships
// @Accept json
// @Produce json
// @Param body body models.CreateScholarshipRequest true "Create scholarship"
// @Success 201 {object} utils.Response{data=models.Scholarship} "Scholarship created successfully"
// @Router /scholarships [post]
func (ctrl *ScholarshipController) CreateScholarship(c *gin.Context) {

	// Get JSON string from form field
	jsonStr := c.PostForm("data")
	if jsonStr == "" {
		utils.JSONIndent(c, http.StatusBadRequest, "Missing JSON payload", nil)
		return
	}

	var input models.CreateScholarshipRequest
	if err := json.Unmarshal([]byte(jsonStr), &input); err != nil {
		utils.JSONIndent(c, http.StatusBadRequest, "Invalid JSON", err.Error())
		return
	}

	title_name := utils.SanitizeString(input.Title)
	if title_name == "" {
		utils.JSONIndent(c, http.StatusBadRequest, "Title cannot be empty or whitespace", nil)
		return
	}

	// replace spaces with underscores and optionally lowercase
	title_name = strings.ReplaceAll(title_name, " ", "_")
	title_name = strings.ToLower(title_name)

	// Handle file upload
	file, handler, err := c.Request.FormFile("photo")
	if err == nil {
		defer file.Close()

		// Extract the extension (e.g. ".png", ".jpg")
		ext := filepath.Ext(handler.Filename)

		// validate extension to avoid weird uploads
		allowed := map[string]bool{".png": true, ".jpg": true, ".jpeg": true}
		if ext != "" && !allowed[ext] {
			utils.JSONIndent(c, http.StatusBadRequest, "Unsupported file type", nil)
			return
		}

		// Create a unique key for the file in S3
		key := fmt.Sprintf("scholarship_logo/%s%s", title_name, ext)

		_, err = storage.S3Client.PutObject(context.TODO(), &s3.PutObjectInput{
			Bucket: &storage.BucketName,
			Key:    &key,
			Body:   file,
		})
		if err != nil {
			utils.JSONIndent(c, http.StatusInternalServerError, "Failed to upload photo", err.Error())
			return
		}

		// Store only the key in the database
		input.PhotoURL = &key
	}

	scholarship, err := ctrl.Queries.CreateScholarship(c, db.CreateScholarshipParams{
		Title:           input.Title,
		Provider:        input.Provider,
		Description:     utils.ToNullString(input.Description),
		InstitutionInfo: utils.ToNullRawMessage(input.InstitutionInfo),
		Requirements:    utils.ToNullRawMessage(input.Requirements),
		ExtraNotes:      utils.ToNullString(input.ExtraNotes),
		DeadlineEnd:     utils.ToNullTime(*input.DeadlineEnd),
		OfficialLink:    utils.ToNullString(input.OfficialLink),
		PhotoUrl:        utils.ToNullString(input.PhotoURL),
	})
	if err != nil {
		logging.Error("Failed to create scholarship: ", err)
		utils.JSONIndent(c, http.StatusInternalServerError, "Failed to perform the operation", nil)
		return
	}

	utils.JSONIndent(c, http.StatusCreated, "Scholarship created successfully", scholarship)
}

// GetScholarships godoc
// @Summary Get all scholarships
// @Tags Scholarships
// @Produce json
// @Success 200 {object} utils.Response{data=[]models.Scholarship} "List of scholarships"
// @Router /scholarships [get]
func (ctrl *ScholarshipController) GetScholarships(c *gin.Context) {

	// responses := utils.BuildScholarshipResponses(scholarships, storage.S3Client, storage.BucketName)
	// utils.RespondOK(c, "Scholarships fetched", responses)

	scholarships, err := ctrl.Queries.GetAllScholarships(c)
	if err != nil {
		utils.JSONIndent(c, http.StatusInternalServerError, "Could not fetch scholarships", err.Error())
		return
	}

	var response []models.ScholarshipResponse
	for _, s := range scholarships {
		sr := builder.MapScholarship(s)

		// Generate presigned URL if PhotoUrl is present
		if s.PhotoUrl.Valid && s.PhotoUrl.String != "" {
			url, err := utils.GenerateScholarshipLogoURL(storage.BucketName, s.PhotoUrl.String, storage.S3Client)
			if err != nil {
				logging.Error("Failed to generate presigned URL for scholarship ", s.ID, ": ", err)
				sr.PhotoURL = nil // or keep original key
			} else {
				sr.PhotoURL = &url
			}
		}

		response = append(response, sr)
	}

	utils.JSONIndent(c, http.StatusOK, "List of scholarships", response)
}

func (ctrl *ScholarshipController) SearchScholarships(c *gin.Context) {
	code := c.Query("code")
	name := c.Query("name")
	program := c.Query("program")

	var scholarships []db.Scholarship
	var err error

	// Case 1: only code
	if code != "" && name == "" && program == "" {
		scholarships, err = ctrl.Queries.GetScholarshipsByInstitutionCodeLike(c, sql.NullString{String: code, Valid: true})
	} else {
		// Flexible search with multiple conditions
		params := db.SearchScholarshipsParams{
			Code:    sql.NullString{String: code, Valid: code != ""},
			Name:    sql.NullString{String: name, Valid: name != ""},
			Program: sql.NullString{String: program, Valid: program != ""},
		}
		scholarships, err = ctrl.Queries.SearchScholarships(c, params)
	}

	if err != nil || len(scholarships) == 0 {
		utils.JSONIndent(c, http.StatusNotFound, "No scholarships found", nil)
		return
	}

	// Use the correct builder for db.Scholarship
	response := builder.BuildScholarshipResponses(scholarships, storage.S3Client, storage.BucketName)

	utils.JSONIndent(c, http.StatusOK, "Search results", response)
}

// DeleteScholarship godoc
// @Summary Delete a scholarship by ID
// @Tags Scholarships
// @Produce json
// @Param id path int true "Scholarship ID"
// @Success 200 {object} utils.Response{data=string} "Scholarship deleted successfully"
// @Router /scholarships/{id} [delete]
func (ctrl *ScholarshipController) DeleteScholarship(c *gin.Context) {

	id, err := utils.GetIDParam(c, "id")
	if err != nil || id <= 0 {
		utils.JSONIndent(c, http.StatusBadRequest, "Invalid scholarship ID", nil)
		return
	}

	err = ctrl.Queries.DeleteScholarshipByID(c, int32(id))
	if err != nil {
		logging.Error("Failed to delete scholarship: ", err)
		utils.JSONIndent(c, http.StatusInternalServerError, "Could not delete scholarship", err.Error())
		return
	}

	utils.JSONIndent(c, http.StatusOK, "Scholarship deleted successfully", nil)
}

// UpdateScholarship godoc
// @Summary Update a scholarship by ID
// @Tags Scholarships
// @Accept json
// @Produce json
// @Param id path int true "Scholarship ID"
// @Param body body models.UpdateScholarshipRequest true "Update scholarship"
// @Success 200 {object} utils.Response{data=models.Scholarship} "Scholarship updated successfully"
// @Router /scholarships/{id} [put]
func (ctrl *ScholarshipController) UpdateScholarship(c *gin.Context) {
	id, err := utils.GetIDParam(c, "id")
	if err != nil || id <= 0 {
		utils.JSONIndent(c, http.StatusBadRequest, "Invalid scholarship ID", nil)
		return
	}

	// Get JSON string from form field (similar to create)
	jsonStr := c.PostForm("data")
	if jsonStr == "" {
		utils.JSONIndent(c, http.StatusBadRequest, "Missing JSON payload", nil)
		return
	}

	var input models.UpdateScholarshipRequest
	if err := json.Unmarshal([]byte(jsonStr), &input); err != nil {
		utils.JSONIndent(c, http.StatusBadRequest, "Invalid JSON", err.Error())
		return
	}

	// Check if scholarship exists
	existing, err := ctrl.Queries.GetScholarshipByID(c, int32(id))
	if err != nil {
		utils.JSONIndent(c, http.StatusNotFound, "Scholarship not found", nil)
		return
	}

	// Handle photo upload if provided
	var photoKey *string
	file, handler, err := c.Request.FormFile("photo")
	if err == nil {
		defer file.Close()

		title_name := utils.SanitizeString(input.Title)
		if title_name == "" {
			title_name = utils.SanitizeString(existing.Title) // fallback to existing title
		}
		title_name = strings.ReplaceAll(strings.ToLower(title_name), " ", "_")

		ext := filepath.Ext(handler.Filename)
		allowed := map[string]bool{".png": true, ".jpg": true, ".jpeg": true}
		if ext != "" && !allowed[ext] {
			utils.JSONIndent(c, http.StatusBadRequest, "Unsupported file type", nil)
			return
		}

		key := fmt.Sprintf("scholarship_logo/%s%s", title_name, ext)

		_, err = storage.S3Client.PutObject(context.TODO(), &s3.PutObjectInput{
			Bucket: &storage.BucketName,
			Key:    &key,
			Body:   file,
		})
		if err != nil {
			utils.JSONIndent(c, http.StatusInternalServerError, "Failed to upload photo", err.Error())
			return
		}

		photoKey = &key
	}

	// Prepare update parameters
	updateParams := db.UpdateScholarshipParams{
		ID:              int32(id),
		Title:           utils.GetStringOrDefault(input.Title, existing.Title),
		Provider:        utils.GetStringOrDefault(input.Provider, existing.Provider),
		Description:     utils.GetNullStringOrExisting(input.Description, existing.Description),
		InstitutionInfo: utils.GetNullRawMessageOrExisting(input.InstitutionInfo, existing.InstitutionInfo),
		Requirements:    utils.GetNullRawMessageOrExisting(input.Requirements, existing.Requirements),
		ExtraNotes:      utils.GetNullStringOrExisting(input.ExtraNotes, existing.ExtraNotes),
		OfficialLink:    utils.GetNullStringOrExisting(input.OfficialLink, existing.OfficialLink),
	}

	// Handle deadline
	if input.DeadlineEnd != nil {
		updateParams.DeadlineEnd = utils.ToNullTime(input.DeadlineEnd.Format("2006-01-02T15:04:05Z07:00"))
	} else {
		updateParams.DeadlineEnd = existing.DeadlineEnd
	}

	// Handle photo URL
	if photoKey != nil {
		updateParams.PhotoUrl = utils.ToNullString(photoKey)
	} else if input.PhotoURL != nil {
		updateParams.PhotoUrl = utils.ToNullString(input.PhotoURL)
	} else {
		updateParams.PhotoUrl = existing.PhotoUrl
	}

	scholarship, err := ctrl.Queries.UpdateScholarship(c, updateParams)
	if err != nil {
		logging.Error("Failed to update scholarship: ", err)
		utils.JSONIndent(c, http.StatusInternalServerError, "Internal Server Error", nil)
		return
	}

	utils.JSONIndent(c, http.StatusOK, "Scholarship updated successfully", scholarship)
}

// UpdateScholarshipJSONB godoc
// @Summary Update specific JSONB fields of a scholarship
// @Tags Scholarships
// @Accept json
// @Produce json
// @Param id path int true "Scholarship ID"
// @Param body body models.UpdateJSONBRequest true "Update JSONB fields"
// @Success 200 {object} utils.Response{data=models.Scholarship} "JSONB fields updated successfully"
// @Router /scholarships/{id}/jsonb [patch]
func (ctrl *ScholarshipController) UpdateScholarshipJSONB(c *gin.Context) {
	id, err := utils.GetIDParam(c, "id")
	if err != nil || id <= 0 {
		utils.JSONIndent(c, http.StatusBadRequest, "Invalid scholarship ID", nil)
		return
	}

	var input models.UpdateJSONBRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.JSONIndent(c, http.StatusBadRequest, "Invalid JSON", err.Error())
		return
	}

	// Check if scholarship exists
	_, err = ctrl.Queries.GetScholarshipByID(c, int32(id))
	if err != nil {
		utils.JSONIndent(c, http.StatusNotFound, "Scholarship not found", nil)
		return
	}

	// Update specific JSONB fields
	params := db.UpdateScholarshipJSONBParams{
		ID: int32(id),
	}

	if input.InstitutionInfo != nil {
		params.InstitutionInfo = utils.ToNullRawMessage(input.InstitutionInfo)
	}

	if input.Requirements != nil {
		params.Requirements = utils.ToNullRawMessage(input.Requirements)
	}

	scholarship, err := ctrl.Queries.UpdateScholarshipJSONB(c, params)
	if err != nil {
		logging.Error("Failed to update scholarship JSONB: ", err)
		utils.JSONIndent(c, http.StatusInternalServerError, "Internal Server Error", nil)
		return
	}

	utils.JSONIndent(c, http.StatusOK, "JSONB fields updated successfully", scholarship)
}
