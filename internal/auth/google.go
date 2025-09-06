package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/url"
	"os"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/database/db"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/tokens"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// GoogleAuthHandler holds the db queries for Google OAuth
type GoogleAuthHandler struct {
	Queries *db.Queries
}

func NewGoogleAuthHandler(queries *db.Queries) *GoogleAuthHandler {
	return &GoogleAuthHandler{Queries: queries}
}

func getGoogleOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  "http://localhost:8080/api/v1/auth/google/callback",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}

func (h *GoogleAuthHandler) GoogleLogin(c *gin.Context) {
	cfg := getGoogleOAuthConfig()
	// TODO: generate a random state and store it in session for security
	url := cfg.AuthCodeURL("state-token", oauth2.AccessTypeOnline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *GoogleAuthHandler) GoogleCallback(c *gin.Context) {
	cfg := getGoogleOAuthConfig()

	code := c.Query("code")
	token, err := cfg.Exchange(context.Background(), code)
	if err != nil {
		utils.JSONIndent(c, http.StatusBadRequest, "Failed to exchange token", err.Error())
		return
	}

	client := cfg.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		utils.JSONIndent(c, http.StatusBadRequest, "Failed to get user info", err.Error())
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		utils.JSONIndent(c, http.StatusInternalServerError, "Failed to decode user info", nil)
		return
	}

	// Check if user exists
	if _, err := h.Queries.FindUserByEmail(c, userInfo.Email); err != nil && err != sql.ErrNoRows {
		utils.JSONIndent(c, http.StatusInternalServerError, "Database error", err.Error())
		return
	} else if err == sql.ErrNoRows {
		// Create user if not exists
		_, err := h.Queries.CreateUser(c, db.CreateUserParams{
			Fullname:     userInfo.Name,
			Email:        userInfo.Email,
			PasswordHash: sql.NullString{Valid: false},
			PhoneNumber:  sql.NullString{Valid: false},
			HighSchool:   sql.NullString{Valid: false},
			GradeLevel:   sql.NullInt32{Valid: false},
			DiplomaYear:  sql.NullInt32{Valid: false},
		})
		if err != nil {
			utils.JSONIndent(c, http.StatusInternalServerError, "Failed to create user", err.Error())
			return
		}
	}

	// Find user to get ID and role for JWT
	user, err := h.Queries.FindUserByEmail(c, userInfo.Email)
	if err != nil {
		utils.JSONIndent(c, http.StatusInternalServerError, "Failed to fetch data", err.Error())
		return
	}

	// Generate JWT token
	tokenString, err := tokens.GenerateToken(user.ID, user.Fullname, user.Email, user.Role.(string))
	if err != nil {
		utils.JSONIndent(c, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}

	// Redirect to frontend with JWT token as query param
	frontendURL := os.Getenv("FRONTEND_URL")
	redirectURL := frontendURL + "?token=" + url.QueryEscape(tokenString)
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

func (h *GoogleAuthHandler) GetLoginURL(c *gin.Context) {
	cfg := getGoogleOAuthConfig()
	loginURL := cfg.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	c.JSON(http.StatusOK, gin.H{
		"login_url": loginURL,
	})
}
