package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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
	jsonStr := c.PostForm("json")
	if jsonStr == "" {
		utils.JSONIndent(c, http.StatusBadRequest, "Missing JSON payload", nil)
		return
	}

	var input models.CreateScholarshipRequest
	if err := json.Unmarshal([]byte(jsonStr), &input); err != nil {
		utils.JSONIndent(c, http.StatusBadRequest, "Invalid JSON", err.Error())
		return
	}

	// Parse JSON string into struct
	// var input models.CreateScholarshipRequest
	// if err := c.ShouldBindJSON(&input); err != nil {
	// 	utils.JSONIndent(c, http.StatusBadRequest, "Missing parameter", err.Error())
	// 	return
	// }

	// Handle file upload (optional)
	file, handler, err := c.Request.FormFile("photo")
	if err == nil { // photo provided
		defer file.Close()

		key := fmt.Sprintf("scholarships/%d-%s", time.Now().UnixNano(), handler.Filename)

		_, err = storage.S3Client.PutObject(context.TODO(), &s3.PutObjectInput{
			Bucket: &storage.BucketName,
			Key:    &key,
			Body:   file,
		})
		if err != nil {
			utils.JSONIndent(c, http.StatusInternalServerError, "Failed to upload photo", err.Error())
			return
		}

		url := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", storage.BucketName, key)
		input.PhotoURL = &url
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
	scholarships, err := ctrl.Queries.GetAllScholarships(c)
	if err != nil {
		utils.JSONIndent(c, http.StatusInternalServerError, "Could not fetch scholarships", err.Error())
		return
	}

	var response []models.ScholarshipResponse
	for _, s := range scholarships {
		response = append(response, utils.MapScholarship(s))
	}

	utils.JSONIndent(c, http.StatusOK, "List of scholarships", response)
}

// // GetScholarshipByID godoc
// // @Summary Get a scholarship by ID
// // @Tags Scholarships
// // @Produce json
// // @Param id path int true "Scholarship ID"
// // @Success 200 {object} utils.Response{data=models.Scholarship} "Scholarship details"
// // @Router /scholarships/{id} [get]
// func (ctrl *ScholarshipController) GetScholarshipByID(c *gin.Context) {
// 	id, err := utils.GetIDParam(c, "id")
// 	if err != nil {
// 		utils.JSONIndent(c, http.StatusBadRequest, "Invalid scholarship ID", err.Error())
// 		return
// 	}

// 	scholarship, err := ctrl.Queries.GetScholarshipByID(c, int32(id))
// 	if err != nil {
// 		utils.JSONIndent(c, http.StatusNotFound, "Scholarship not found", nil)
// 		return
// 	}

// 	utils.JSONIndent(c, http.StatusOK, "Scholarship details", scholarship)
// }

// // DeleteScholarship godoc
// // @Summary Delete a scholarship by ID
// // @Tags Scholarships
// // @Produce json
// // @Param id path int true "Scholarship ID"
// // @Success 200 {object} utils.Response{data=string} "Scholarship deleted successfully"
// // @Router /scholarships/{id} [delete]
// func (ctrl *ScholarshipController) DeleteScholarship(c *gin.Context) {
// 	id, err := utils.GetIDParam(c, "id")
// 	if err != nil {
// 		utils.JSONIndent(c, http.StatusBadRequest, "Invalid scholarship ID", err.Error())
// 		return
// 	}

// 	err = ctrl.Queries.DeleteScholarship(c, int32(id))
// 	if err != nil {
// 		utils.JSONIndent(c, http.StatusInternalServerError, "Could not delete scholarship", err.Error())
// 		return
// 	}

// 	utils.JSONIndent(c, http.StatusOK, "Scholarship deleted successfully", nil)
// }
