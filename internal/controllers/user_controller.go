package controllers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/database/db"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/logging"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/models"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/tokens"
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

func getUserClaims(c *gin.Context) (*tokens.UserClaims, bool) {
	claims, ok := c.Get("claims")
	if !ok {
		return nil, false
	}

	userClaims, ok := claims.(*tokens.UserClaims)
	if !ok {
		return nil, false
	}

	return userClaims, true
}

// GetProfile godoc
// @Summary Get unified user profile with student information
// @Tags Users
// @Produce json
// @Success 200 {object} models.UnifiedUserProfileResponse "User profile fetched successfully"
// @Security BearerAuth
// @Router /profile [get]
func (u *UserController) GetProfile(c *gin.Context) {
	userClaims, ok := getUserClaims(c)
	if !ok {
		utils.JSONIndent(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	ctx := c.Request.Context()
	userWithProfile, err := u.Queries.GetUserWithStudentProfile(ctx, int32(userClaims.ID))
	if err != nil {
		if err == sql.ErrNoRows {
			utils.JSONIndent(c, http.StatusNotFound, "User not found", nil)
			return
		}
		go logging.Error("[User Controller]: Failed to get user profile: ", err)
		utils.JSONIndent(c, http.StatusInternalServerError, "Failed to fetch user profile", nil)
		return
	}

	response := buildUnifiedUserProfileResponse(userWithProfile)
	utils.JSONIndent(c, http.StatusOK, "User profile fetched successfully", response)
}

// UpdateProfile godoc
// @Summary Update user profile and student information atomically
// @Description Updates only the provided fields using atomic transaction. Supports clearing arrays by passing empty arrays.
// @Tags Users
// @Accept json
// @Produce json
// @Param body body models.UpdateUserProfileRequest true "Fields to update (partial update supported)"
// @Success 200 {object} models.UnifiedUserProfileResponse "Profile updated successfully"
// @Failure 400 {object} utils.Response "Invalid input"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Failure 500 {object} utils.Response "Internal server error"
// @Security BearerAuth
// @Router /profile [put]
func (u *UserController) UpdateProfile(c *gin.Context) {
	var input models.UpdateUserProfileRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		go logging.Error("[User Controller]: Failed to bind update profile input: ", err)
		utils.JSONIndent(c, http.StatusBadRequest, "Invalid input", err.Error())
		return
	}

	// Validate input using utils helper
	if err := utils.ValidateUserProfileInput(input); err != nil {
		utils.JSONIndent(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	userClaims, ok := getUserClaims(c)
	if !ok {
		utils.JSONIndent(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	ctx := c.Request.Context()
	userID := int32(userClaims.ID)

	// Start atomic transaction
	tx, err := u.DB.BeginTx(ctx, nil)
	if err != nil {
		go logging.Error("[User Controller]: Failed to begin transaction: ", err)
		utils.JSONIndent(c, http.StatusInternalServerError, "Failed to begin transaction", nil)
		return
	}
	defer tx.Rollback()

	queries := u.Queries.WithTx(tx)

	// Update user basic info and student profile atomically
	updatedProfile, err := u.updateProfileAtomic(ctx, queries, userID, input)
	if err != nil {
		go logging.Error("[User Controller]: Failed to update profile: ", err)
		utils.JSONIndent(c, http.StatusInternalServerError, "Failed to update profile", nil)
		return
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		go logging.Error("[User Controller]: Failed to commit transaction: ", err)
		utils.JSONIndent(c, http.StatusInternalServerError, "Failed to commit changes", nil)
		return
	}

	response := buildUnifiedUserProfileResponse(updatedProfile)
	utils.JSONIndent(c, http.StatusOK, "Profile updated successfully", response)
}

// updateProfileAtomic performs atomic update of user and student profile within transaction
func (u *UserController) updateProfileAtomic(ctx context.Context, queries *db.Queries, userID int32, input models.UpdateUserProfileRequest) (db.GetUserWithStudentProfileRow, error) {
	// Get current profile data
	currentProfile, err := queries.GetUserWithStudentProfile(ctx, userID)
	if err != nil {
		return db.GetUserWithStudentProfileRow{}, err
	}

	// Update user basic info if provided - using utils helper
	if utils.HasUserFields(input) {
		params := u.buildUserUpdateParams(userID, currentProfile, input)
		_, err := queries.UpdateUserProfile(ctx, params)
		if err != nil {
			return db.GetUserWithStudentProfileRow{}, err
		}
	}

	// Handle student profile - using utils helpers
	if utils.HasStudentProfileFields(input) {
		if utils.HasExistingStudentProfile(currentProfile) {
			// Update existing student profile
			params := u.buildStudentUpdateParams(userID, currentProfile, input)
			_, err := queries.UpdateStudentProfile(ctx, params)
			if err != nil {
				return db.GetUserWithStudentProfileRow{}, err
			}
		} else {
			// Create new student profile
			params := u.buildStudentCreateParams(userID, input)
			_, err := queries.CreateStudentProfile(ctx, params)
			if err != nil {
				return db.GetUserWithStudentProfileRow{}, err
			}
		}
	}

	// Return updated profile using RETURNING clause optimization
	return queries.GetUserWithStudentProfile(ctx, userID)
}

// Helper methods for building parameters
func (u *UserController) buildUserUpdateParams(userID int32, current db.GetUserWithStudentProfileRow, input models.UpdateUserProfileRequest) db.UpdateUserProfileParams {
	params := db.UpdateUserProfileParams{
		ID:       userID,
		Fullname: current.Fullname,
	}

	if input.Fullname != nil {
		params.Fullname = sql.NullString{String: *input.Fullname, Valid: *input.Fullname != ""}
	}

	return params
}

func (u *UserController) buildStudentUpdateParams(userID int32, current db.GetUserWithStudentProfileRow, input models.UpdateUserProfileRequest) db.UpdateStudentProfileParams {
	params := db.UpdateStudentProfileParams{
		StudentID:    userID,
		HighSchool:   current.HighSchool,
		GradeLevel:   current.GradeLevel,
		DiplomaYear:  current.DiplomaYear,
		DiplomaGrade: current.DiplomaGrade,
		SelectMajors: current.SelectMajors,
	}

	if input.HighSchool != nil {
		params.HighSchool = sql.NullString{String: *input.HighSchool, Valid: *input.HighSchool != ""}
	}
	if input.GradeLevel != nil {
		params.GradeLevel = sql.NullString{String: *input.GradeLevel, Valid: *input.GradeLevel != ""}
	}
	if input.DiplomaYear != nil {
		params.DiplomaYear = sql.NullInt32{Int32: *input.DiplomaYear, Valid: *input.DiplomaYear > 0}
	}
	if input.DiplomaGrade != nil {
		params.DiplomaGrade = sql.NullString{String: *input.DiplomaGrade, Valid: *input.DiplomaGrade != ""}
	}
	if input.SelectMajors != nil { // Handle empty arrays correctly
		params.SelectMajors = utils.SliceToNullRawMessage(input.SelectMajors)
	}

	return params
}

func (u *UserController) buildStudentCreateParams(userID int32, input models.UpdateUserProfileRequest) db.CreateStudentProfileParams {
	params := db.CreateStudentProfileParams{
		StudentID:    userID,
		HighSchool:   sql.NullString{Valid: false},
		GradeLevel:   sql.NullString{Valid: false},
		DiplomaYear:  sql.NullInt32{Valid: false},
		DiplomaGrade: sql.NullString{Valid: false},
		SelectMajors: pqtype.NullRawMessage{Valid: false},
	}

	if input.HighSchool != nil {
		params.HighSchool = sql.NullString{String: *input.HighSchool, Valid: *input.HighSchool != ""}
	}
	if input.GradeLevel != nil {
		params.GradeLevel = sql.NullString{String: *input.GradeLevel, Valid: *input.GradeLevel != ""}
	}
	if input.DiplomaYear != nil {
		params.DiplomaYear = sql.NullInt32{Int32: *input.DiplomaYear, Valid: *input.DiplomaYear > 0}
	}
	if input.DiplomaGrade != nil {
		params.DiplomaGrade = sql.NullString{String: *input.DiplomaGrade, Valid: *input.DiplomaGrade != ""}
	}
	if input.SelectMajors != nil {
		params.SelectMajors = utils.SliceToNullRawMessage(input.SelectMajors)
	}

	return params
}

// buildUnifiedUserProfileResponse constructs response with proper null handling
func buildUnifiedUserProfileResponse(data db.GetUserWithStudentProfileRow) models.UnifiedUserProfileResponse {
	response := models.UnifiedUserProfileResponse{
		ID:        data.UserID,
		Fullname:  data.Fullname.String,
		Email:     data.Email,
		CreatedAt: data.ProfileCreatedAt.Time,
		UpdatedAt: data.ProfileUpdatedAt.Time,
	}

	// Add student profile if any student data exists - using utils helper
	if utils.HasExistingStudentProfile(data) {
		var selectMajors []string
		if data.SelectMajors.Valid {
			if err := json.Unmarshal(data.SelectMajors.RawMessage, &selectMajors); err != nil {
				go logging.Error("[User Controller]: Failed to unmarshal select majors: ", err)
				selectMajors = []string{} // Default to empty array on error
			}
		}

		response.StudentProfile = &models.StudentProfileResponse{
			HighSchool:   data.HighSchool.String,
			GradeLevel:   data.GradeLevel.String,
			DiplomaYear:  data.DiplomaYear.Int32,
			DiplomaGrade: data.DiplomaGrade.String,
			SelectMajors: selectMajors,
			CreatedAt:    data.ProfileCreatedAt.Time,
			UpdatedAt:    data.ProfileUpdatedAt.Time,
		}
	}

	return response
}
