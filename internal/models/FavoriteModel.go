package models

type FavoriteScholarship struct {
	ID            int32 `json:"id"`
	UserID        int32 `json:"user_id"`
	ScholarshipID int32 `json:"scholarship_id"`
}

type CreateFavoriteRequest struct {
	UserID        int32 `json:"user_id,omitempty"`
	ScholarshipID int32 `json:"scholarship_id,omitempty"`
}

type ListFavoritesByUserRow struct {
    ID            int32
    UserID        int32
    ScholarshipID int32
}

type FavoriteScholarshipListResponse struct {
	Favorites []ScholarshipResponse
}

type DeleteFavoriteScholarshipRequest struct {
	UserID        int32 `json:"user_id" binding:"required"`
	ScholarshipID int32 `json:"scholarship_id" binding:"required"`
}

type UpdateFavoriteStatusRequest struct {
    IsFavorite *bool `json:"is_favorite" binding:"required"`
}
