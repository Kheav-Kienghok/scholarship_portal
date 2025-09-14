package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

// // GoogleAuthHandler holds the db queries for Google OAuth
// type GoogleAuthHandler struct {
// 	Queries *db.Queries
// }

type FBProfile struct {
	ID       string `json:"id"`
	FullName string `json:"name"`
	Email    string `json:"email"`
	Picture  struct {
		Data struct {
			Height       int    `json:"height"`
			IsSilhouette bool   `json:"is_silhouette"`
			URL          string `json:"url"`
			Width        int    `json:"width"`
		} `json:"data"`
	} `json:"picture"`
}

// GenerateState generates a random string for OAuth state parameter
func GenerateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

type FBAuthHandler struct {
	OAuthConfig *oauth2.Config
}

// func NewGoogleAuthHandler(queries *db.Queries) *GoogleAuthHandler {
// 	return &GoogleAuthHandler{Queries: queries}
// }

// NewFBAuthHandler creates a new FBAuthHandler with config from env
func NewFBAuthHandler() *FBAuthHandler {
	clientID := os.Getenv("FB_APP_ID")
	clientSecret := os.Getenv("FB_APP_SECRET")
	redirectURI := os.Getenv("REDIRECT_URI")

	if clientID == "" || clientSecret == "" {
		panic("FB_APP_ID and FB_APP_SECRET must be set")
	}

	oauthConfig := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURI,
		Scopes:       []string{"public_profile", "email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.facebook.com/v17.0/dialog/oauth",
			TokenURL: "https://graph.facebook.com/v17.0/oauth/access_token",
		},
	}

	return &FBAuthHandler{
		OAuthConfig: oauthConfig,
	}
}

// LoginHandler returns the Facebook OAuth login URL
func (h *FBAuthHandler) LoginHandler(c *gin.Context) {

	state, err := GenerateState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate state"})
		return
	}
	// Save state in cookie
	c.SetCookie("oauth_state", state, 300, "/", "", false, true)

	loginURL := h.OAuthConfig.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, loginURL)
	// c.JSON(http.StatusOK, gin.H{"login_url": loginURL})
}

// CallbackHandler exchanges code for token and returns typed profile
func (h *FBAuthHandler) CallbackHandler(c *gin.Context) {
	state := c.Query("state")
	storedState, err := c.Cookie("oauth_state")
	if err != nil || state != storedState {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid state parameter (possible CSRF)"})
		return
	}

	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing code in callback"})
		return
	}

	ctx := context.Background()
	token, err := h.OAuthConfig.Exchange(ctx, code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("token exchange failed: %v", err)})
		return
	}

	client := h.OAuthConfig.Client(ctx, token)
	resp, err := client.Get("https://graph.facebook.com/v17.0/me?fields=id,name,email,picture.width(200).height(200)")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to fetch profile: %v", err)})
		return
	}
	defer resp.Body.Close()

	var profile FBProfile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("decode error: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":    token.AccessToken,
		"user_id":         profile.ID,
		"full_name":       profile.FullName,
		"email":           profile.Email,
		"profile_picture": profile.Picture.Data.URL,
	})
}
