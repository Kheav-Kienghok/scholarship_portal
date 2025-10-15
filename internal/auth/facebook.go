package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/database/db"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/logging"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/tokens"
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
	b := make([]byte, 32) // Increased from 16 to 32 bytes for better security
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
		panic("Facebook OAuth credentials not configured")
	}

	// Log the redirect URI for debugging
	logging.Info("Facebook OAuth Redirect URI:", redirectURI)

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

	logging.Info("Generated state:", state)

	// Store state in session/cookie for verification in callback
	// Use more permissive cookie settings for development

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "fb_oauth_state",
		Value:    state,
		Path:     "/",
		Domain:   "localhost", // make sure backend runs on localhost
		MaxAge:   600,
		HttpOnly: true,
		Secure:   false,                 // false for local HTTP
		SameSite: http.SameSiteNoneMode, // important! allows cross-site cookies
	})

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "fb_state_backup",
		Value:    state,
		Path:     "/",
		Domain:   "localhost",
		MaxAge:   600,
		HttpOnly: false,
		Secure:   false,
		SameSite: http.SameSiteNoneMode,
	})

	loginURL := h.OAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)

	// Log the generated URL for debugging
	logging.Info("Generated Facebook OAuth URL:", loginURL)

	// Check if it's an API request or browser request
	if c.GetHeader("Accept") == "application/json" {
		c.JSON(http.StatusOK, gin.H{
			"success":   true,
			"message":   "Facebook login URL generated",
			"login_url": loginURL,
			"state":     state, // Include state in response for debugging
		})
	} else {
		c.Redirect(http.StatusTemporaryRedirect, loginURL)
	}
}

// CallbackHandler exchanges code for token and handles user creation/login
func (h *FBAuthHandler) CallbackHandler(c *gin.Context) {
	// Log the callback request for debugging
	logging.Info("Facebook callback received with params:", c.Request.URL.RawQuery)

	// Get state from query parameter
	state := c.Query("state")
	logging.Info("Received state from Facebook:", state)

	// Try to get stored state from cookies (try both)
	storedState, err := c.Cookie("fb_oauth_state")
	if err != nil {
		logging.Info("Primary state cookie not found, trying backup...")
		storedState, err = c.Cookie("fb_state_backup")
		if err != nil {
			logging.Error("No state cookies found:", err)
		}
	}

	cookies := c.Request.Cookies()
	for _, cookie := range cookies {
		logging.Info("Cookie received:", cookie.Name, "=", cookie.Value)
	}
	logging.Info("Facebook returned state:", state)

	logging.Info("Stored state from cookie:", storedState)

	// Verify state parameter
	if err != nil || state == "" || storedState == "" || state != storedState {
		logging.Error("OAuth state mismatch or missing. Expected:", storedState, "Got:", state, "Error:", err)

		// For development, let's be more lenient and provide detailed error info
		errorMsg := "Invalid state parameter"
		if state == "" {
			errorMsg = "Missing state parameter from Facebook"
		} else if storedState == "" {
			errorMsg = "No stored state found in cookies"
		} else if state != storedState {
			errorMsg = fmt.Sprintf("State mismatch: expected %s, got %s", storedState, state)
		}

		// Redirect to frontend with detailed error for debugging
		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			frontendURL = "http://localhost:8080" // Updated default
		}

		c.Redirect(http.StatusTemporaryRedirect,
			fmt.Sprintf("%s?error=%s", frontendURL, url.QueryEscape(errorMsg)))
		return
	}

	// Clear the state cookies
	c.SetCookie("fb_oauth_state", "", -1, "/", "", false, true)
	c.SetCookie("fb_state_backup", "", -1, "/", "", false, false)

	// Check for authorization error
	errMsg := c.Query("error")
	errDescription := c.Query("error_description")

	if errMsg != "" {
		logging.Error("Facebook OAuth error:", errMsg, "Description:", errDescription)

		// Handle specific Facebook errors
		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			frontendURL = "http://localhost:8080" // Updated default
		}

		errorMessage := errMsg
		if errDescription != "" {
			errorMessage = fmt.Sprintf("%s: %s", errMsg, errDescription)
		}

		c.Redirect(http.StatusTemporaryRedirect,
			fmt.Sprintf("%s?error=%s", frontendURL, url.QueryEscape(errorMessage)))
		return
	}

	// Get authorization code
	code := c.Query("code")
	if code == "" {
		logging.Error("Missing authorization code in Facebook callback")

		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			frontendURL = "http://localhost:8080" // Updated default
		}

		c.Redirect(http.StatusTemporaryRedirect,
			fmt.Sprintf("%s?error=%s", frontendURL, url.QueryEscape("Missing authorization code")))
		return
	}

	// Exchange code for token
	ctx := context.Background()
	token, err := h.OAuthConfig.Exchange(ctx, code)
	if err != nil {
		logging.Error("Token exchange failed:", err)

		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			frontendURL = "http://localhost:8080" // Updated default
		}

		c.Redirect(http.StatusTemporaryRedirect,
			fmt.Sprintf("%s?error=%s", frontendURL, url.QueryEscape("Failed to exchange authorization code")))
		return
	}

	// Fetch user profile from Facebook
	client := h.OAuthConfig.Client(ctx, token)
	resp, err := client.Get("https://graph.facebook.com/v17.0/me?fields=id,name,email,picture.width(200).height(200)")
	if err != nil {
		logging.Error("Failed to fetch Facebook profile:", err)

		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			frontendURL = "http://localhost:8080" // Updated default
		}

		c.Redirect(http.StatusTemporaryRedirect,
			fmt.Sprintf("%s?error=%s", frontendURL, url.QueryEscape("Failed to fetch user profile")))
		return
	}
	defer resp.Body.Close()

	// Parse Facebook profile
	var profile FBProfile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		logging.Error("Failed to decode Facebook profile:", err)

		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			frontendURL = "http://localhost:8080" // Updated default
		}

		c.Redirect(http.StatusTemporaryRedirect,
			fmt.Sprintf("%s?error=%s", frontendURL, url.QueryEscape("Failed to parse user profile")))
		return
	}

	// Validate required fields
	if profile.ID == "" {
		logging.Error("Facebook profile missing required fields")

		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			frontendURL = "http://localhost:8080" // Updated default
		}

		c.Redirect(http.StatusTemporaryRedirect,
			fmt.Sprintf("%s?error=%s", frontendURL, url.QueryEscape("Invalid profile data")))
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

			if profile.Email == "" {
				createUserParams.Email = ""
			}

			userRow, err := h.Queries.CreateUser(ctx, createUserParams)
			if err != nil {
				logging.Error("Failed to create user:", err)

				frontendURL := os.Getenv("FRONTEND_URL")
				if frontendURL == "" {
					frontendURL = "http://localhost:8080" // Updated default
				}

				c.Redirect(http.StatusTemporaryRedirect,
					fmt.Sprintf("%s?error=%s", frontendURL, url.QueryEscape("Failed to create user account")))
				return
			}

			// Fetch the created user as db.User type
			user, err = h.Queries.GetUserByID(ctx, userRow.ID)
			if err != nil {
				logging.Error("Failed to fetch created user:", err)

				frontendURL := os.Getenv("FRONTEND_URL")
				if frontendURL == "" {
					frontendURL = "http://localhost:8080" // Updated default
				}

				c.Redirect(http.StatusTemporaryRedirect,
					fmt.Sprintf("%s?error=%s", frontendURL, url.QueryEscape("Failed to fetch created user")))
				return
			}

			// Create OAuth login record
			refreshToken := ""
			if token.RefreshToken != "" {
				refreshToken = token.RefreshToken
			}

			_, err = h.Queries.CreateOAuthLogin(ctx, db.CreateOAuthLoginParams{
				UserID:         sql.NullInt32{Int32: user.ID, Valid: true},
				Provider:       "facebook",
				ProviderUserID: profile.ID,
				AccessToken:    token.AccessToken,
				RefreshToken:   sql.NullString{String: refreshToken, Valid: refreshToken != ""},
			})
			if err != nil {
				logging.Error("Failed to create OAuth login:", err)

				frontendURL := os.Getenv("FRONTEND_URL")
				if frontendURL == "" {
					frontendURL = "http://localhost:8080" // Updated default
				}

				c.Redirect(http.StatusTemporaryRedirect,
					fmt.Sprintf("%s?error=%s", frontendURL, url.QueryEscape("Failed to create OAuth record")))
				return
			}

			logging.Info("New Facebook user created:", user.ID)
		} else {
			logging.Error("Database error:", err)

			frontendURL := os.Getenv("FRONTEND_URL")
			if frontendURL == "" {
				frontendURL = "http://localhost:8080" // Updated default
			}

			c.Redirect(http.StatusTemporaryRedirect,
				fmt.Sprintf("%s?error=%s", frontendURL, url.QueryEscape("Database error occurred")))
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

			frontendURL := os.Getenv("FRONTEND_URL")
			if frontendURL == "" {
				frontendURL = "http://localhost:8080" // Updated default
			}

			c.Redirect(http.StatusTemporaryRedirect,
				fmt.Sprintf("%s?error=%s", frontendURL, url.QueryEscape("Failed to update OAuth record")))
			return
		}

		user = existingUser
		logging.Info("Existing Facebook user logged in:", user.ID)
	}

	// Generate JWT token
	jwtToken, err := tokens.GenerateToken(user.ID, user.Fullname.String, user.Email, "user")
	if err != nil {
		logging.Error("Failed to generate JWT:", err)

		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			frontendURL = "http://localhost:8080" // Updated default
		}

		c.Redirect(http.StatusTemporaryRedirect,
			fmt.Sprintf("%s?error=%s", frontendURL, url.QueryEscape("Failed to generate authentication token")))
		return
	}

	// Set JWT as HTTP-only cookie
	maxAge := 7 * 24 * 60 * 60 // 7 days in seconds
	secure := c.Request.TLS != nil
	c.SetCookie("auth_token", jwtToken, maxAge, "/", "", secure, true)

	// Redirect to frontend with success
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:8080" // Updated default
	}

	c.Redirect(http.StatusTemporaryRedirect,
		fmt.Sprintf("%s?token=%s&provider=facebook", frontendURL, url.QueryEscape(jwtToken)))
}
