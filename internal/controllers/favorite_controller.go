package controllers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/database/db"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/logging"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/models"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/tokens"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/sqlc-dev/pqtype"
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

func (ctrl *FavoriteController) getUserIDOrAbort(c *gin.Context) int64 {
	userID, err := getUserIDFromClaims(c)
	if err != nil {
		utils.RespondUnauthorized(c, err.Error())
		return 0
	}
	return userID
}

func nullStringToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func nullRawMessageToJSON(n pqtype.NullRawMessage) json.RawMessage {
	if n.Valid {
		return n.RawMessage
	}
	return json.RawMessage("null")
}

func nullTimeToPtr(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}

func nullStringToPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

// AddFavorite adds a scholarship to the user's favorites
func (ctrl *FavoriteController) AddFavorite(c *gin.Context) {

	userID := ctrl.getUserIDOrAbort(c)
	if userID == 0 {
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
		logging.Error("Failed to add favorite: ", err)
		utils.RespondInternalError(c, "Failed to add favorite")
		return
	}

	utils.RespondOK(c, "Favorite added", nil)
}

// RemoveFavorite removes a scholarship from the user's favorites
func (ctrl *FavoriteController) RemoveFavorite(c *gin.Context) {

	userID := ctrl.getUserIDOrAbort(c)
	if userID == 0 {
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

	userID := ctrl.getUserIDOrAbort(c)
	if userID == 0 {
		return
	}

	// Fetch favorite entries
	favorites, err := ctrl.Queries.ListFavoritesByUser(c, userID)
	if err != nil {
		logging.Error("Failed to fetch favorites: ", err)
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
		logging.Error("Failed to fetch scholarships: ", err)
		utils.RespondInternalError(c, "Failed to fetch scholarships")
		return
	}

	// Map to response
	scholarships := make([]models.ScholarshipResponse, len(dbScholarships))
	for i, s := range dbScholarships {
		scholarships[i] = models.ScholarshipResponse{
			ID:              int(s.ID),
			Title:           s.Title,
			Provider:        s.Provider,
			Description:     nullStringToString(s.Description),
			InstitutionInfo: nullRawMessageToJSON(s.InstitutionInfo),
			Requirements:    nullRawMessageToJSON(s.Requirements),
			ExtraNotes:      nullStringToString(s.ExtraNotes),
			DeadlineEnd:     nullTimeToPtr(s.DeadlineEnd),
			OfficialLink:    nullStringToPtr(s.OfficialLink),
			PhotoURL:        nullStringToPtr(s.PhotoUrl),
			CreatedAt:       *nullTimeToPtr(s.CreatedAt),
		}
	}

	utils.RespondOK(c, "Favorites fetched", models.FavoriteScholarshipListResponse{
		Favorites: scholarships,
	})
}
