package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/database/db"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/logging"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/models"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/tokens"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/utils"
	"github.com/gin-gonic/gin"
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

// func sliceToNullRawMessage(slice []string) pqtype.NullRawMessage {
// 	if len(slice) == 0 {
// 		return pqtype.NullRawMessage{Valid: false}
// 	}
// 	b, _ := json.Marshal(slice)
// 	return pqtype.NullRawMessage{
// 		RawMessage: b,
// 		Valid:      true,
// 	}
// }

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
// @Summary Get user profile information
// @Tags Users
// @Produce json
// @Success 200 {object} models.UserProfileResponse "Fetch User successfully"
// @Security BearerAuth
// @Router /profile [get]
func (u *UserController) GetUserProfile(c *gin.Context) {

	userClaims, ok := getUserClaims(c)
	if !ok {
		utils.JSONIndent(c, http.StatusUnauthorized, "Something went wrong!", nil)
		return
	}

	ctx := c.Request.Context()

	arg := db.GetUserWithProfileParams{
		ID:    int32(userClaims.ID),
		Email: userClaims.Email,
	}

	userWithProfile, err := u.Queries.GetUserWithProfile(ctx, arg)
	if err != nil {
		logging.Error("[User Controller]: Failed to get user profile: ", err)
		utils.JSONIndent(c, http.StatusInternalServerError, "Failed to fetch user profile", nil)
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
		ID:               userWithProfile.ID,
		Fullname:         userWithProfile.Fullname.String,
		Email:            userWithProfile.Email,
		PhoneNumber:      userWithProfile.PhoneNumber.String,
		ProfileCreatedAt: userWithProfile.ProfileCreatedAt.Time,
		ProfileUpdatedAt: userWithProfile.ProfileUpdatedAt.Time,
	}

	utils.JSONIndent(c, http.StatusOK, "User profile fetched successfully", response)
}
