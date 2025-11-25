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
	"github.com/Kheav-Kienghok/scholarship_portal/internal/errors"
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
	var input models.CreateScholarshipRequest
	var err error

	contentType := c.Request.Header.Get("Content-Type")

	if strings.HasPrefix(contentType, "application/json") {
		// Handle raw JSON body
		if err := c.ShouldBindJSON(&input); err != nil {
			utils.JSONIndent(c, http.StatusBadRequest, "Invalid JSON", err.Error())
			return
		}
	} else {
		// Assume multipart/form-data
		jsonStr := c.PostForm("data")
		if jsonStr == "" {
			utils.JSONIndent(c, http.StatusBadRequest, "Missing JSON payload", nil)
			return
		}

		if err = json.Unmarshal([]byte(jsonStr), &input); err != nil {
			utils.JSONIndent(c, http.StatusBadRequest, "Invalid JSON", err.Error())
			return
		}

		// Handle file upload
		file, handler, err := c.Request.FormFile("photo_url")
		if err == nil {
			defer file.Close()

			ext := filepath.Ext(handler.Filename)
			allowed := map[string]bool{".png": true, ".jpg": true, ".jpeg": true}
			if ext != "" && !allowed[ext] {
				utils.JSONIndent(c, http.StatusBadRequest, "Unsupported file type", nil)
				return
			}

			key := fmt.Sprintf("scholarship_logo/%s%s", strings.ToLower(strings.ReplaceAll(utils.SanitizeString(input.Title), " ", "_")), ext)

			_, err = storage.S3Client.PutObject(context.TODO(), &s3.PutObjectInput{
				Bucket: &storage.BucketName,
				Key:    &key,
				Body:   file,
			})
			if err != nil {
				utils.JSONIndent(c, http.StatusInternalServerError, "Failed to upload photo", err.Error())
				return
			}

			input.PhotoURL = &key
		}
	}

	// Sanitize title
	title_name := utils.SanitizeString(input.Title)
	if title_name == "" {
		utils.JSONIndent(c, http.StatusBadRequest, "Title cannot be empty or whitespace", nil)
		return
	}

	if input.Provider == "" {
		utils.JSONIndent(c, http.StatusBadRequest, "Provider is required", nil)
		return
	}

	if input.DeadlineEnd == nil {
		utils.JSONIndent(c, http.StatusBadRequest, "DeadlineEnd is required", nil)
		return
	}

	// Save to DB
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
		Categories:      utils.ToNullRawMessage(input.Categories),
	})
	if err != nil {
		go logging.Error("Failed to create scholarship: ", err)
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

	scholarships, err := ctrl.Queries.GetAllScholarships(c)
	if err != nil {
		if err == sql.ErrNoRows {
			// Return empty array instead of error for no rows
			utils.JSONIndent(c, http.StatusOK, "No scholarships found", []models.ScholarshipResponse{})
			return
		}
		// Apply sanitized error handling
		errors.SanitizedErrorResponse(c, err, http.StatusInternalServerError, "Could not fetch scholarships")
		return
	}
	response := builder.BuildAllScholarshipResponses(scholarships, storage.S3Client, storage.BucketName)
	utils.JSONIndent(c, http.StatusOK, "List of scholarships", response)
}

func (cr *ScholarshipController) GetScholarshipByID(c *gin.Context) {
	id, err := utils.GetIDParam(c, "id")
	if err != nil || id <= 0 {
		utils.JSONIndent(c, http.StatusBadRequest, "Invalid scholarship ID", nil)
		return
	}

	scholarship, err := cr.Queries.GetScholarshipByID(c, int32(id))
	if err != nil {
		if err == sql.ErrNoRows {
			utils.JSONIndent(c, http.StatusNotFound, "Scholarship not found", nil)
			return
		}
		errors.SanitizedErrorResponse(c, err, http.StatusInternalServerError, "Could not fetch scholarship")
		return
	}

	response := builder.BuildGetScholarshipByIDResponseFromRow(scholarship, storage.S3Client, storage.BucketName)

	// Generate presigned URL if PhotoUrl is present
	if scholarship.PhotoUrl.Valid && scholarship.PhotoUrl.String != "" {
		url, err := utils.GenerateScholarshipLogoURL(storage.BucketName, scholarship.PhotoUrl.String, storage.S3Client)
		if err != nil {
			go logging.Error("Failed to generate presigned URL for scholarship ", scholarship.ID, ": ", err)
			response.PhotoURL = nil // or keep original key
		} else {
			response.PhotoURL = &url
		}
	}

	utils.JSONIndent(c, http.StatusOK, "Scholarship details", response)
}

// GetActiveScholarships godoc
// @Summary Get active scholarships (deadline not passed)
// @Description Returns scholarships where deadline_end is greater than current date
// @Tags Scholarships
// @Produce json
// @Success 200 {object} utils.Response{data=[]models.Scholarship} "List of active scholarships"
// @Router /scholarships/active [get]
func (ctrl *ScholarshipController) GetActiveScholarship(c *gin.Context) {
	scholarships, err := ctrl.Queries.GetActiveScholarships(c)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.JSONIndent(c, http.StatusOK, "No active scholarships found", []models.ScholarshipResponse{})
			return
		}
		errors.SanitizedErrorResponse(c, err, http.StatusInternalServerError, "Could not fetch active scholarships")
		return
	}

	response := builder.BuildActiveScholarshipResponsesFromDB(scholarships, storage.S3Client, storage.BucketName)

	utils.JSONIndent(c, http.StatusOK, "List of active scholarships", response)
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

	response := builder.BuildScholarshipResponsesFromDB(scholarships, storage.S3Client, storage.BucketName)
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
		go logging.Error("Failed to delete scholarship: ", err)
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

	// Check if scholarship exists
	existing, err := ctrl.Queries.GetScholarshipByID(c, int32(id))
	if err != nil {
		utils.JSONIndent(c, http.StatusNotFound, "Scholarship not found", nil)
		return
	}

	contentType := c.ContentType()
	var input models.UpdateScholarshipRequest

	// --- Handle multipart/form-data (photo + JSON fields) ---
	if strings.HasPrefix(contentType, "multipart/form-data") {
		jsonStr := c.PostForm("data")
		if jsonStr == "" {
			utils.JSONIndent(c, http.StatusBadRequest, "Missing JSON payload", nil)
			return
		}
		if err := json.Unmarshal([]byte(jsonStr), &input); err != nil {
			utils.JSONIndent(c, http.StatusBadRequest, "Invalid JSON", err.Error())
			return
		}

		// Handle optional photo upload
		file, handler, err := c.Request.FormFile("photo_url")
		if err == nil {
			defer file.Close()
			titleName := utils.SanitizeString(input.Title)
			if titleName == "" {
				titleName = utils.SanitizeString(existing.Title)
			}
			titleName = strings.ReplaceAll(strings.ToLower(titleName), " ", "_")

			ext := filepath.Ext(handler.Filename)
			allowed := map[string]bool{".png": true, ".jpg": true, ".jpeg": true}
			if ext != "" && !allowed[ext] {
				utils.JSONIndent(c, http.StatusBadRequest, "Unsupported file type", nil)
				return
			}

			key := fmt.Sprintf("scholarship_logo/%s%s", titleName, ext)
			_, err = storage.S3Client.PutObject(context.TODO(), &s3.PutObjectInput{
				Bucket: &storage.BucketName,
				Key:    &key,
				Body:   file,
			})
			if err != nil {
				utils.JSONIndent(c, http.StatusInternalServerError, "Failed to upload photo", err.Error())
				return
			}

			input.PhotoURL = &key
		}
	}

	// --- Handle plain JSON requests ---
	if strings.HasPrefix(contentType, "application/json") {
		if err := c.ShouldBindJSON(&input); err != nil {
			utils.JSONIndent(c, http.StatusBadRequest, "Invalid JSON", err.Error())
			return
		}
	}

	// --- Prepare update parameters ---
	updateParams := db.UpdateScholarshipParams{
		ID:              int32(id),
		Title:           utils.GetStringOrDefault(input.Title, existing.Title),
		Provider:        utils.GetStringOrDefault(input.Provider, existing.Provider),
		Description:     utils.GetNullStringOrExisting(input.Description, existing.Description),
		InstitutionInfo: utils.GetNullRawMessageOrExisting(input.InstitutionInfo, existing.InstitutionInfo),
		Requirements:    utils.GetNullRawMessageOrExisting(input.Requirements, existing.Requirements),
		ExtraNotes:      utils.GetNullStringOrExisting(input.ExtraNotes, existing.ExtraNotes),
		OfficialLink:    utils.GetNullStringOrExisting(input.OfficialLink, existing.OfficialLink),
		DeadlineEnd:     utils.ParseDeadlineEnd(input.DeadlineEnd, existing.DeadlineEnd),
		PhotoUrl:        existing.PhotoUrl, // will update later if input.PhotoURL != nil
	}

	// --- Handle photo ---
	if input.PhotoURL != nil {
		updateParams.PhotoUrl = utils.ToNullString(input.PhotoURL)
	} else {
		updateParams.PhotoUrl = existing.PhotoUrl
	}

	// --- Execute update ---
	scholarship, err := ctrl.Queries.UpdateScholarship(c, updateParams)
	if err != nil {
		go logging.Error("Failed to update scholarship: ", err)
		utils.JSONIndent(c, http.StatusInternalServerError, "Internal Server Error", nil)
		return
	}

	// --- Build response using scholarship model ---
	response := builder.BuildScholarshipResponse(
		builder.NewScholarshipWrapper(scholarship),
		storage.S3Client,
		storage.BucketName,
	)

	// Generate presigned URL if PhotoUrl is present
	if scholarship.PhotoUrl.Valid && scholarship.PhotoUrl.String != "" {
		url, err := utils.GenerateScholarshipLogoURL(storage.BucketName, scholarship.PhotoUrl.String, storage.S3Client)
		if err != nil {
			go logging.Error("Failed to generate presigned URL for scholarship ", scholarship.ID, ": ", err)
			response.PhotoURL = nil // or keep original key
		} else {
			response.PhotoURL = &url
		}
	}

	utils.JSONIndent(c, http.StatusOK, "Scholarship updated successfully", response)
}

// FilterByCategory godoc
// @Summary Filter scholarships by program category
// @Description Filter scholarships by program categories (IT, Engineering, Business, etc.)
// @Tags Scholarships
// @Produce json
// @Param category query string true "Category name (IT, Engineering, Business, Economics & Finance, Medicine & Health, Science, Arts & Humanities)"
// @Success 200 {object} utils.Response{data=[]models.Scholarship} "Filtered scholarships"
// @Router /scholarships/filter [get]
func (ctrl *ScholarshipController) FilterByCategory(c *gin.Context) {
	category := c.Query("category")
	if category == "" {
		utils.JSONIndent(c, http.StatusBadRequest, "Category parameter is required", nil)
		return
	}

	// Get all programs for this category
	programs := utils.ExpandCategoryToPrograms(category)
	if len(programs) == 0 {
		utils.JSONIndent(c, http.StatusBadRequest, "Invalid category", gin.H{
			"valid_categories": utils.ValidCategories(),
		})
		return
	}

	// Query database with all programs
	scholarships, err := ctrl.Queries.SearchScholarshipsByPrograms(c, programs)
	if err != nil {
		errors.SanitizedErrorResponse(c, err, http.StatusInternalServerError, "Could not filter scholarships")
		return
	}

	if len(scholarships) == 0 {
		utils.JSONIndent(c, http.StatusOK, "No scholarships found for this category", []models.ScholarshipResponse{})
		return
	}

	// response := builder.BuildSearchScholarshipResponses(scholarships, storage.S3Client, storage.BucketName)
	response := builder.BuildSearchScholarshipResponsesFromDB(scholarships, storage.S3Client, storage.BucketName)

	utils.JSONIndent(c, http.StatusOK, fmt.Sprintf("Found %d scholarship(s) in %s category", len(scholarships), category), response)
}
