package controllers

import (
	"database/sql"
	"net/http"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/database/db"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/models"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/utils"
	"github.com/gin-gonic/gin"
)

type UserController struct {
	Queries *db.Queries
}

func UserControllerHandler(queries *db.Queries) *UserController {
	return &UserController{Queries: queries}
}

type UpdateProfileRequest struct {
	Fullname    string         `json:"fullname"`
	PhoneNumber sql.NullString `json:"phone_number"`
	HighSchool  sql.NullString `json:"high_school"`
	GradeLevel  sql.NullInt32  `json:"grade_level"`
	DiplomaYear sql.NullInt32  `json:"diploma_year"`
}

func (u *UserController) UpdateProfile(c *gin.Context) {
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

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONIndent(c, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	params := db.UpdateUserProfileParams{
		ID:          int32(userClaims.ID),
		Fullname:    req.Fullname,
		PhoneNumber: req.PhoneNumber,
		HighSchool:  req.HighSchool,
		GradeLevel:  req.GradeLevel,
		DiplomaYear: req.DiplomaYear,
	}

	_, err := u.Queries.UpdateUserProfile(c, params)
	if err != nil {
		utils.JSONIndent(c, http.StatusInternalServerError, "Failed to update profile", nil)
		return
	}

	utils.JSONIndent(c, http.StatusOK, "Profile updated successfully", nil)
}
