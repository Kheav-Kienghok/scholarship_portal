package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/database/db"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/logging"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/models"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/sqlc-dev/pqtype"
)

type UserController struct {
	DB      *sql.DB
	Queries *db.Queries
}

func UserControllerHandler(dbConn *sql.DB, queries *db.Queries) *UserController {
	return &UserController{
		DB:      dbConn,
		Queries: queries,
	}
}

func sliceToNullRawMessage(slice []string) pqtype.NullRawMessage {
	if len(slice) == 0 {
		return pqtype.NullRawMessage{Valid: false}
	}
	b, _ := json.Marshal(slice)
	return pqtype.NullRawMessage{
		RawMessage: b,
		Valid:      true,
	}
}

func (u *UserController) UpdateUserAndProfile(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		utils.JSONIndent(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	userClaims, ok := claims.(*models.Claims)
	if !ok {
		utils.JSONIndent(c, http.StatusUnauthorized, "Invalid claims", nil)
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		logging.Error("Failed to bind JSON: ", err)
		utils.JSONIndent(c, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	ctx := c.Request.Context()

	// Start transaction
	tx, err := u.DB.BeginTx(ctx, nil)
	if err != nil {
		logging.Error("Failed to start transaction: ", err)
		utils.JSONIndent(c, http.StatusInternalServerError, "Failed to start transaction", nil)
		return
	}
	defer tx.Rollback()

	qtx := db.New(tx)

	hasUserFields := false

	// --- Update User Table ---
	userParams := db.UpdateUserProfileParams{
		ID: int32(userClaims.ID),
	}

	if val, ok := req["fullname"].(string); ok {
		userParams.Fullname = sql.NullString{String: val, Valid: val != ""}
		hasUserFields = true
	}
	if val, ok := req["phone_number"].(string); ok {
		userParams.PhoneNumber = sql.NullString{String: val, Valid: val != ""}
		hasUserFields = true
	}
	if val, ok := req["high_school"].(string); ok {
		userParams.HighSchool = sql.NullString{String: val, Valid: val != ""}
		hasUserFields = true
	}
	if val, ok := req["grade_level"].(float64); ok {
		userParams.GradeLevel = sql.NullInt32{Int32: int32(val), Valid: true}
		hasUserFields = true
	}
	if val, ok := req["diploma_year"].(float64); ok {
		userParams.DiplomaYear = sql.NullInt32{Int32: int32(val), Valid: true}
		hasUserFields = true
	}

	if hasUserFields {
		logging.Info("Prepared user update params: ", userParams)
		if err := qtx.UpdateUserProfile(ctx, userParams); err != nil {
			logging.Error("Failed to update user profile: ", err)
			utils.JSONIndent(c, http.StatusInternalServerError, "Failed to update user", nil)
			return
		}
	}

	if err := qtx.UpdateUserProfile(ctx, userParams); err != nil {
		logging.Error("Failed to update user profile: ", err)
		utils.JSONIndent(c, http.StatusInternalServerError, "Failed to update user", nil)
		return
	}

	// --- Update Student Profile ---
	studentParams := db.UpdateStudentProfileParams{
		StudentID: int32(userClaims.ID),
	}

	// Only set fields if they exist in JSON
	hasStudentFields := false

	if val, ok := req["diploma_grade"].(string); ok {
		studentParams.DiplomaGrade = db.NullDiplomaGrade{
			DiplomaGrade: db.DiplomaGrade(val),
			Valid:        val != "",
		}
		hasStudentFields = true
	}
	if majors, ok := req["select_majors"].([]interface{}); ok {
		var majorsStr []string
		for _, m := range majors {
			if s, ok := m.(string); ok {
				majorsStr = append(majorsStr, s)
			}
		}
		studentParams.SelectMajors = sliceToNullRawMessage(majorsStr)
		hasStudentFields = true
	}

	// Skip update if nothing to update
	if hasStudentFields {
		logging.Info("Prepared student update params: ", studentParams)

		if err := qtx.UpdateStudentProfile(ctx, studentParams); err != nil {
			logging.Error("Failed to update student profile: ", err)
			utils.JSONIndent(c, http.StatusInternalServerError, "Failed to update student profile", nil)
			return
		}
	}

	if err := qtx.UpdateStudentProfile(ctx, studentParams); err != nil {
		logging.Error("Failed to update student profile: ", err)
		utils.JSONIndent(c, http.StatusInternalServerError, "Failed to update student profile", nil)
		return
	}

	// --- Commit Transaction ---
	if err := tx.Commit(); err != nil {
		logging.Error("Transaction commit failed: ", err)
		utils.JSONIndent(c, http.StatusInternalServerError, "Transaction commit failed", nil)
		return
	}

	utils.JSONIndent(c, http.StatusOK, "Profile updated successfully", nil)
}
