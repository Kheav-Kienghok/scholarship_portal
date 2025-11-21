package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/builder"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/database/db"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/logging"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/models"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/otpstore"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/storage"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/tokens"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/utils"
	"github.com/gin-gonic/gin"
)

type FavoriteController struct {
	Queries       *db.Queries
	ActionLimiter *otpstore.OTPStore
}

func FavoriteControllerHandler(queries *db.Queries) *FavoriteController {
	return &FavoriteController{
		Queries:       queries,
		ActionLimiter: otpstore.NewOTPStore(10, 1*time.Minute), // Max 10 actions per minute
	}
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

func (ctrl *FavoriteController) getUserIDOrAbort(c *gin.Context) int64 {
	userID, err := getUserIDFromClaims(c)
	if err != nil {
		utils.RespondUnauthorized(c, err.Error())
		return 0
	}
	return userID
}

func (ctrl *FavoriteController) checkActionLimit(c *gin.Context, userID int64, action string) bool {
	key := fmt.Sprintf("fav:%s:%d", action, userID)

	remaining, locked := ctrl.ActionLimiter.Remaining(key)
	if locked {
		utils.RespondTooManyRequests(c, "Too many favorite actions. Please wait.",
			int(ctrl.ActionLimiter.Window().Seconds()))
		return false
	}

	if remaining <= 2 {
		go logging.Warn(fmt.Sprintf("User %d approaching rate limit for %s: %d remaining", userID, action, remaining))
	}

	ctrl.ActionLimiter.Increment(key)
	return true
}

// AddFavorite adds a scholarship to the user's favorites
func (ctrl *FavoriteController) AddFavorite(c *gin.Context) {

	userID := ctrl.getUserIDOrAbort(c)
	if userID == 0 {
		return
	}

	// Check rate limit
	if !ctrl.checkActionLimit(c, userID, "add") {
		return
	}

	var req models.CreateFavoriteRequest
	if !utils.BindJSONOrFail(c, &req) {
		return
	}

	err := ctrl.Queries.AddFavorite(c, db.AddFavoriteParams{
		UserID:        userID,
		ScholarshipID: int64(req.ScholarshipID),
	})
	if err != nil {
		errMsg := err.Error()

		// --- 1. Scholarship not found (foreign key violation) ---
		if strings.Contains(errMsg, "foreign key") || strings.Contains(errMsg, "23503") {
			utils.JSONIndent(c, http.StatusNotFound, "Scholarship not found or no longer available", nil)
			return
		}

		go logging.Error("Failed to add favorite:", err)
		utils.RespondInternalError(c, "Failed to add favorite")
		return
	}

	go logging.Info(fmt.Sprintf("User %d added favorite %d", userID, req.ScholarshipID))
	utils.RespondOK(c, "Favorite added", nil)
}

// RemoveFavorite removes a scholarship from the user's favorites
func (ctrl *FavoriteController) RemoveFavorite(c *gin.Context) {

	userID := ctrl.getUserIDOrAbort(c)
	if userID == 0 {
		return
	}

	// Check rate limit
	if !ctrl.checkActionLimit(c, userID, "remove") {
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
		go logging.Error("Failed to remove favorite:", err)
		utils.RespondInternalError(c, "Failed to remove favorite")
		return
	}

	utils.RespondOK(c, "Favorite removed", nil)
}

// ListFavorites lists all favorites for the user
func (ctrl *FavoriteController) ListFavorites(c *gin.Context) {

	userID := ctrl.getUserIDOrAbort(c)
	if userID == 0 {
		return
	}

	// Fetch favorite entries
	favorites, err := ctrl.Queries.ListFavoritesByUser(c, userID)
	if err != nil {
		go logging.Error("Failed to fetch favorites: ", err)
		utils.RespondInternalError(c, "Failed to fetch favorites")
		return
	}

	if len(favorites) == 0 {
		utils.RespondOK(c, "No favorites found", models.FavoriteScholarshipListResponse{Favorites: []models.ScholarshipResponse{}})
		return
	}

	// Prepare IDs for batch query
	scholarshipIDs := make([]int32, len(favorites))
	for i, fav := range favorites {
		scholarshipIDs[i] = int32(fav.ScholarshipID)
	}

	// Batch fetch scholarships
	dbScholarships, err := ctrl.Queries.GetScholarshipsByIDs(c, scholarshipIDs)
	if err != nil {
		go logging.Error("Failed to fetch scholarships: ", err)
		utils.RespondInternalError(c, "Failed to fetch scholarships")
		return
	}

	wrapped := make([]builder.ScholarshipSource, len(dbScholarships))
	for i, s := range dbScholarships {
		wrapped[i] = builder.NewScholarshipWrapper(s)
	}

	responses := builder.BuildScholarshipResponses(wrapped, storage.S3Client, storage.BucketName)
	utils.RespondOK(c, "Favorites fetched", models.FavoriteScholarshipListResponse{
		Favorites: responses,
	})
}

func (ctrl *FavoriteController) UpdateFavoriteStatus(c *gin.Context) {

	userID := ctrl.getUserIDOrAbort(c)
	if userID == 0 {
		return
	}

	var req models.UpdateFavoriteStatusRequest
	if !utils.BindJSONOrFail(c, &req) {
		return
	}

	scholarshipIDStr := c.Param("scholarship_id")
	scholarshipID, err := strconv.Atoi(scholarshipIDStr)
	if err != nil {
		utils.RespondBadRequest(c, "Invalid scholarship ID", nil)
		return
	}

	err = ctrl.Queries.UpdateFavoriteStatus(c, db.UpdateFavoriteStatusParams{
		IsFavorite:    *req.IsFavorite,
		UserID:        userID,
		ScholarshipID: int64(scholarshipID),
	})
	if err != nil {
		go logging.Error("Failed to update favorite status:", err)
		utils.RespondInternalError(c, "Failed to update favorite status")
		return
	}

	utils.RespondOK(c, "Favorite status updated", nil)
}
