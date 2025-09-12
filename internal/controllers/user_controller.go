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

// UpdateProfile godoc
// @Summary Update user and profile information
// @Tags Users
// @Accept json
// @Produce json
// @Param body body models.UpdateUserRequest true "Update user profile"
// @Success 200 {object} utils.APIUpdateResponse "Update User Profile successfully"
// @Security BearerAuth
// @Router /update-profile [patch]
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

	// var req map[string]interface{}
	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logging.Error("Failed to bind JSON: ", err)
		utils.JSONIndent(c, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	ctx := c.Request.Context()

	// --- Check if any fields are present for update ---
	hasUserFields := req.Fullname != nil || req.PhoneNumber != nil || req.HighSchool != nil || req.GradeLevel != nil || req.DiplomaYear != nil
	hasStudentFields := req.DiplomaGrade != nil || len(req.SelectMajors) > 0

	if !hasUserFields && !hasStudentFields {
		utils.JSONIndent(c, http.StatusBadRequest, "No fields to update", nil)
		return
	}

	// Start transaction
	tx, err := u.DB.BeginTx(ctx, nil)
	if err != nil {
		logging.Error("Failed to start transaction: ", err)
		utils.JSONIndent(c, http.StatusInternalServerError, "Failed to start transaction", nil)
		return
	}
	defer tx.Rollback()

	qtx := db.New(tx)

	// --- Update User Table ---
	userParams := db.UpdateUserProfileParams{ID: int32(userClaims.ID)}

	if req.Fullname != nil {
		userParams.Fullname = sql.NullString{String: *req.Fullname, Valid: *req.Fullname != ""}
	}
	if req.PhoneNumber != nil {
		userParams.PhoneNumber = sql.NullString{String: *req.PhoneNumber, Valid: *req.PhoneNumber != ""}
	}
	if req.HighSchool != nil {
		userParams.HighSchool = sql.NullString{String: *req.HighSchool, Valid: *req.HighSchool != ""}
	}
	if req.GradeLevel != nil {
		userParams.GradeLevel = sql.NullInt32{Int32: int32(*req.GradeLevel), Valid: true}
	}
	if req.DiplomaYear != nil {
		userParams.DiplomaYear = sql.NullInt32{Int32: int32(*req.DiplomaYear), Valid: true}
	}

	if hasUserFields {
		if err := qtx.UpdateUserProfile(ctx, userParams); err != nil {
			logging.Error("Failed to update user profile: ", err)
			utils.JSONIndent(c, http.StatusInternalServerError, "Failed to update user", nil)
			return
		}
	}

	// --- Update Student Profile ---
	studentParams := db.UpdateStudentProfileParams{StudentID: int32(userClaims.ID)}

	if req.DiplomaGrade != nil {
		studentParams.DiplomaGrade = db.NullDiplomaGrade{
			DiplomaGrade: db.DiplomaGrade(*req.DiplomaGrade),
			Valid:        *req.DiplomaGrade != "",
		}
	}

	if len(req.SelectMajors) > 0 {
		studentParams.SelectMajors = sliceToNullRawMessage(req.SelectMajors)
	}

	if hasStudentFields {
		if err := qtx.UpdateStudentProfile(ctx, studentParams); err != nil {
			logging.Error("Failed to update student profile: ", err)
			utils.JSONIndent(c, http.StatusInternalServerError, "Failed to update student profile", nil)
			return
		}
	}

	// --- Commit Transaction ---
	if err := tx.Commit(); err != nil {
		logging.Error("Transaction commit failed: ", err)
		utils.JSONIndent(c, http.StatusInternalServerError, "Transaction commit failed", nil)
		return
	}

	utils.JSONIndent(c, http.StatusOK, "Profile updated successfully", nil)
}

// GetProfile godoc
// @Summary Get user profile information
// @Tags Users
// @Produce json
// @Success 200 {object} models.UserProfileResponse "Fetch User successfully"
// @Security BearerAuth
// @Router /profile [get]
func (u *UserController) GetUserProfile(c *gin.Context) {
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

	ctx := c.Request.Context()

	arg := db.GetUserWithProfileParams{
		ID:    int32(userClaims.ID),
		Email: userClaims.Email,
	}

	userWithProfile, err := u.Queries.GetUserWithProfile(ctx, arg)
	if err != nil {
		logging.Error("Failed to get user profile: ", err)
		utils.JSONIndent(c, http.StatusInternalServerError, "Failed to get user profile", nil)
		return
	}

	var selectMajors []string
	if userWithProfile.SelectMajors.Valid {
		if err := json.Unmarshal(userWithProfile.SelectMajors.RawMessage, &selectMajors); err != nil {
			logging.Error("Failed to unmarshal select majors: ", err)
			utils.JSONIndent(c, http.StatusInternalServerError, "Failed to parse select majors", nil)
			return
		}
	}

	response := models.UserProfileResponse{
		ID:          userWithProfile.ID,
		Fullname:    userWithProfile.Fullname.String,
		Email:       userWithProfile.Email,
		PhoneNumber: userWithProfile.PhoneNumber.String,
		HighSchool:  userWithProfile.HighSchool.String,
		GradeLevel:  int(userWithProfile.GradeLevel.Int32),
		DiplomaYear: int(userWithProfile.DiplomaYear.Int32),
		StudentProfile: models.StudentProfile{
			DiplomaGrade: string(userWithProfile.DiplomaGrade.DiplomaGrade),
			SelectMajors: selectMajors,
		},
		ProfileCreatedAt: userWithProfile.ProfileCreatedAt.Time,
		ProfileUpdatedAt: userWithProfile.ProfileUpdatedAt.Time,
	}

	utils.JSONIndent(c, http.StatusOK, "User profile fetched successfully", response)
}
