package controllers

import (
	"errors"
	"strconv"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/database/db"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/logging"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/models"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/tokens"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/utils"
	"github.com/gin-gonic/gin"
)

type FavoriteController struct {
	Queries *db.Queries
}

func FavoriteControllerHandler(queries *db.Queries) *FavoriteController {
	return &FavoriteController{Queries: queries}
}

// getUserIDFromClaims extracts the user ID from claims in Gin context
func getUserIDFromClaims(c *gin.Context) (int64, error) {
	claimsVal, exists := c.Get("claims")
	if !exists {
		return 0, errors.New("unauthorized: no claims found")
	}

	userClaims, ok := claimsVal.(tokens.ClaimsInterface)
	if !ok {
		return 0, errors.New("unauthorized: invalid claims type")
	}

	return userClaims.GetID(), nil
}

// AddFavorite adds a scholarship to the user's favorites
func (ctrl *FavoriteController) AddFavorite(c *gin.Context) {

	userID, err := getUserIDFromClaims(c)
	if err != nil {
		utils.RespondUnauthorized(c, err.Error())
		return
	}

	var req models.CreateFavoriteRequest
	if !utils.BindJSONOrFail(c, &req) {
		return
	}

	err = ctrl.Queries.AddFavorite(c, db.AddFavoriteParams{
		UserID:        userID,
		ScholarshipID: int64(req.ScholarshipID),
	})
	if err != nil {
		logging.Error("Failed to add favorite: ", err)
		utils.RespondInternalError(c, "Failed to add favorite")
		return
	}

	utils.RespondOK(c, "Favorite added", nil)
}

// RemoveFavorite removes a scholarship from the user's favorites
func (ctrl *FavoriteController) RemoveFavorite(c *gin.Context) {

	userID, err := getUserIDFromClaims(c)
	if err != nil {
		utils.RespondUnauthorized(c, err.Error())
		return
	}

	scholarshipIDStr := c.Param("scholarship_id")
	scholarshipID, err := strconv.Atoi(scholarshipIDStr)
	if err != nil {
		utils.RespondBadRequest(c, "Invalid scholarship ID", nil)
		return
	}

	err = ctrl.Queries.RemoveFavorite(c, db.RemoveFavoriteParams{
		UserID:        userID,
		ScholarshipID: int64(scholarshipID),
	})
	if err != nil {
		utils.RespondInternalError(c, "Failed to remove favorite")
		return
	}

	utils.RespondOK(c, "Favorite removed", nil)
}

// ListFavorites lists all favorites for the user
func (ctrl *FavoriteController) ListFavorites(c *gin.Context) {

	userID, err := getUserIDFromClaims(c)
	if err != nil {
		utils.RespondUnauthorized(c, err.Error())
		return
	}

	favorites, err := ctrl.Queries.ListFavoritesByUser(c, userID)
	if err != nil {
		utils.RespondInternalError(c, "Failed to fetch favorites")
		return
	}

	utils.RespondOK(c, "Favorites fetched", favorites)
}
