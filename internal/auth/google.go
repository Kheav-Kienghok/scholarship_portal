package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"os"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/database/db"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/errors"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/logging"
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

// generateRandomState creates a URL-safe random string
func GenerateRandomState() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err) // or handle gracefully
	}
	return base64.URLEncoding.EncodeToString(b)
}

func getGoogleOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}

func (h *GoogleAuthHandler) GoogleLogin(c *gin.Context) {

	cfg := getGoogleOAuthConfig()
	state := GenerateRandomState()

	url := cfg.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	// url := cfg.AuthCodeURL(state, oauth2.AccessTypeOnline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *GoogleAuthHandler) GoogleCallback(c *gin.Context) {
	cfg := getGoogleOAuthConfig()

	code := c.Query("code")
	token, err := cfg.Exchange(context.Background(), code)
	if err != nil {
		logging.Error("GOOGLE: Failed to exchange token: ", err)
		errors.SanitizedErrorResponse(c, err, http.StatusInternalServerError, "Failed to authenticate with Google")
		return
	}

	client := cfg.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		logging.Error("GOOGLE: Failed to get user info:", err)
		utils.JSONIndent(c, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		logging.Error("GOOGLE: Failed to decode user info:", err)
		utils.JSONIndent(c, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}

	// Check if user exists
	user, err := h.Queries.GetUserByEmail(c, userInfo.Email)
	if err != nil && err != sql.ErrNoRows {
		logging.Error("DB: Failed to find user:", err)
		utils.JSONIndent(c, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}

	if err == sql.ErrNoRows {
		// Create user if not exists
		_, err = h.Queries.CreateUser(c, db.CreateUserParams{
			Email:        userInfo.Email,
			Fullname:     sql.NullString{String: userInfo.Name, Valid: true},
			PasswordHash: sql.NullString{Valid: false},
		})
		if err != nil {
			logging.Error("DB: Failed to create user:", err)
			utils.JSONIndent(c, http.StatusInternalServerError, "Something went wrong", nil)
			return
		}

		// Fetch the user again to get the correct ID
		user, err = h.Queries.GetUserByEmail(c, userInfo.Email)
		if err != nil {
			logging.Error("DB: Failed to fetch user after creation:", err)
			utils.JSONIndent(c, http.StatusInternalServerError, "Something went wrong", nil)
			return
		}
	}

	// Now user.ID is guaranteed to exist in users table!
	_, err = h.Queries.UpsertOauthLogin(c, db.UpsertOauthLoginParams{
		UserID:         sql.NullInt32{Int32: user.ID, Valid: true},
		Provider:       "google",
		ProviderUserID: userInfo.ID,
		AccessToken:    token.AccessToken,
		RefreshToken:   sql.NullString{String: token.RefreshToken, Valid: token.RefreshToken != ""},
	})
	if err != nil {
		logging.Error("DB: Failed to upsert OAuth login:", err)
		utils.JSONIndent(c, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}

	// Generate JWT token
	tokenString, err := tokens.GenerateToken(user.ID, user.Fullname.String, user.Email, "student")
	if err != nil {
		logging.Error("JWT: Failed to generate token:", err)
		utils.JSONIndent(c, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}

	// // Set JWT in HttpOnly cookie for React frontend
	// c.SetCookie("jwt", tokenString, 3600, "/", "localhost", false, true)

	// // Redirect to frontend
	// c.Redirect(http.StatusTemporaryRedirect, os.Getenv("FRONTEND_URL"))

	frontendURL := os.Getenv("FRONTEND_URL") + "?token=" + url.QueryEscape(tokenString)
	c.Redirect(http.StatusTemporaryRedirect, frontendURL)
}
