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

type FavoriteScholarshipListResponse struct {
	Favorites []FavoriteScholarship `json:"favorites"`
}

type CreateFavoriteScholarshipRequest struct {
	UserID        int32 `json:"user_id" binding:"required"`
	ScholarshipID int32 `json:"scholarship_id" binding:"required"`
}

type DeleteFavoriteScholarshipRequest struct {
	UserID        int32 `json:"user_id" binding:"required"`
	ScholarshipID int32 `json:"scholarship_id" binding:"required"`
}

type FavoriteScholarshipResponse struct {
	ID            int32 `json:"id"`
	UserID        int32 `json:"user_id"`
	ScholarshipID int32 `json:"scholarship_id"`
}
