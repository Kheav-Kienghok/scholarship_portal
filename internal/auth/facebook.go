package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/database/db"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/errors"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/logging"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/tokens"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

// FBAuthHandler holds the OAuth config and db queries for Facebook OAuth
type FBAuthHandler struct {
	OAuthConfig *oauth2.Config
	Queries     *db.Queries
}

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

// NewFBAuthHandler creates a new FBAuthHandler with config from env
func NewFBAuthHandler(queries *db.Queries) *FBAuthHandler {
	clientID := os.Getenv("FB_APP_ID")
	clientSecret := os.Getenv("FB_APP_SECRET")
	redirectURI := os.Getenv("FB_REDIRECT_URL")

	if clientID == "" || clientSecret == "" || redirectURI == "" {
		logging.Error("FB_APP_ID, FB_APP_SECRET, and FB_REDIRECT_URL must be set")
		return nil
		// panic("Facebook OAuth credentials not configured")
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
		Queries:     queries,
	}
}

// LoginHandler returns the Facebook OAuth login URL
func (h *FBAuthHandler) LoginHandler(c *gin.Context) {
	state, err := GenerateState()
	if err != nil {
		logging.Error("Failed to generate OAuth state:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Could not generate state parameter",
		})
		return
	}

	// Save state in secure cookie (5 minutes expiry)
	secure := c.Request.TLS != nil // true if HTTPS
	c.SetCookie("oauth_state", state, 300, "/", "", secure, true)
	// c.SetCookie("oauth_state", state, 300, "/", "", false, true)

	loginURL := h.OAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)

	// Check if it's an API request or browser request
	if c.GetHeader("Accept") == "application/json" {
		c.JSON(http.StatusOK, gin.H{"login_url": loginURL})
	} else {
		c.Redirect(http.StatusTemporaryRedirect, loginURL)
	}
}

// CallbackHandler exchanges code for token and handles user creation/login
func (h *FBAuthHandler) CallbackHandler(c *gin.Context) {
	// Verify state parameter
	state := c.Query("state")
	storedState, err := c.Cookie("oauth_state")
	if err != nil || state == "" || state != storedState {
		logging.Error("Invalid OAuth state - possible CSRF attack")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_state",
			"message": "Invalid state parameter (possible CSRF)",
		})
		return
	}

	// Clear the state cookie
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)

	// Check for authorization error
	errMsg := c.Query("error")
	if errMsg != "" {
		logging.Error("Facebook OAuth error:", errMsg)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "oauth_error",
			"message": fmt.Sprintf("Facebook authorization failed: %s", errMsg),
		})
		return
	}

	// Get authorization code
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "missing_code",
			"message": "Missing authorization code in callback",
		})
		return
	}

	// Exchange code for token
	ctx := context.Background()
	token, err := h.OAuthConfig.Exchange(ctx, code)
	if err != nil {
		logging.Error("Token exchange failed:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "token_exchange_failed",
			"message": "Failed to exchange authorization code for token",
		})
		return
	}

	// Fetch user profile from Facebook
	client := h.OAuthConfig.Client(ctx, token)
	resp, err := client.Get("https://graph.facebook.com/v17.0/me?fields=id,name,email,picture.width(200).height(200)")
	if err != nil {
		logging.Error("Failed to fetch Facebook profile:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "profile_fetch_failed",
			"message": "Failed to fetch user profile from Facebook",
		})
		return
	}
	defer resp.Body.Close()

	// Parse Facebook profile
	var profile FBProfile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		logging.Error("Failed to decode Facebook profile:", err)
		utils.JSONIndent(c, http.StatusInternalServerError, "Failed to parse user profile", nil)
		return
	}

	// Validate required fields
	if profile.ID == "" {
		utils.JSONIndent(c, http.StatusInternalServerError, "Facebook profile missing required fields", nil)
		return
	}

	// Check if OAuth login exists
	existingUser, err := h.Queries.GetUserByOAuthProvider(ctx, db.GetUserByOAuthProviderParams{
		Provider:       "facebook",
		ProviderUserID: profile.ID,
	})

	var user db.User
	if err != nil {
		if err == sql.ErrNoRows {
			// Create new user
			createUserParams := db.CreateUserParams{
				Fullname: sql.NullString{
					String: profile.FullName,
					Valid:  profile.FullName != "",
				},
				Email: profile.Email,
				PasswordHash: sql.NullString{ // empty because OAuth login
					String: "",
					Valid:  false,
				},
			}

			if profile.Email != "" {
				createUserParams.Email = profile.Email
			} else {
				createUserParams.Email = ""
			}

			// // Add profile picture if available
			// if profile.Picture.Data.URL != "" {
			// 	createUserParams.ProfilePicture = sql.NullString{String: profile.Picture.Data.URL, Valid: true}
			// }

			userRow, err := h.Queries.CreateUser(ctx, createUserParams)
			if err != nil {
				logging.Error("Failed to create user:", err)
				errors.SanitizedErrorResponse(c, err, http.StatusInternalServerError, "Failed to create user account")
				return
			}

			// Create OAuth login record
			refreshToken := ""
			if token.RefreshToken != "" {
				refreshToken = token.RefreshToken
			}

			_, err = h.Queries.CreateOAuthLogin(ctx, db.CreateOAuthLoginParams{
				UserID:         sql.NullInt32{Int32: userRow.ID, Valid: true},
				Provider:       "facebook",
				ProviderUserID: profile.ID,
				AccessToken:    token.AccessToken,
				RefreshToken:   sql.NullString{String: refreshToken, Valid: refreshToken != ""},
			})
			if err != nil {
				logging.Error("Failed to create OAuth login:", err)
				utils.JSONIndent(c, http.StatusInternalServerError, "Failed to create OAuth record", nil)
				return
			}

			logging.Info("New Facebook user created:", user.ID)
		} else {
			logging.Error("Database error:", err)
			utils.JSONIndent(c, http.StatusInternalServerError, "Database error", nil)
			return
		}
	} else {
		// User exists, update OAuth login
		refreshToken := ""
		if token.RefreshToken != "" {
			refreshToken = token.RefreshToken
		}

		_, err = h.Queries.UpdateOAuthLogin(ctx, db.UpdateOAuthLoginParams{
			UserID:       sql.NullInt32{Int32: existingUser.ID, Valid: true},
			Provider:     "facebook",
			AccessToken:  token.AccessToken,
			RefreshToken: sql.NullString{String: refreshToken, Valid: refreshToken != ""},
		})
		if err != nil {
			logging.Error("Failed to update OAuth login:", err)
			utils.JSONIndent(c, http.StatusInternalServerError, "Failed to update OAuth record", nil)
			return
		}

		user = existingUser
		logging.Info("Existing Facebook user logged in:", user.ID)
	}

	// Generate JWT token
	jwtToken, err := tokens.GenerateToken(user.ID, user.Fullname.String, user.Email, "user")
	if err != nil {
		logging.Error("Failed to generate JWT:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "token_generation_failed",
			"message": "Failed to generate authentication token",
		})
		return
	}

	// Prepare user response
	userResponse := gin.H{
		"id":         user.ID,
		"full_name":  user.Fullname,
		"provider":   "facebook",
		"created_at": user.CreatedAt.Time.UTC().Format("2006-01-02"),
	}

	if user.Email != "" {
		userResponse["email"] = user.Email
	}

	// if user.ProfilePicture.Valid {
	// 	userResponse["profile_picture"] = user.ProfilePicture.String
	// }

	// Set JWT as HTTP-only cookie
	maxAge := 7 * 24 * 60 * 60 // 7 days in seconds
	secure := c.Request.TLS != nil
	c.SetCookie("auth_token", jwtToken, maxAge, "/", "", secure, true)

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"user":         userResponse,
		"access_token": token.AccessToken,
		"message":      "Facebook authentication successful",
	})
}
