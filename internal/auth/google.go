package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET")) // make sure you set this in .env

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

func GoogleLogin(c *gin.Context) {
	cfg := getGoogleOAuthConfig()

	// TODO: generate a random state and store it in session for security
	url := cfg.AuthCodeURL("state-token", oauth2.AccessTypeOnline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func GoogleCallback(c *gin.Context) {
	cfg := getGoogleOAuthConfig()

	code := c.Query("code")
	token, err := cfg.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to exchange token", "details": err.Error()})
		return
	}

	client := cfg.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get user info", "details": err.Error()})
		return
	}
	defer resp.Body.Close()

	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode user info"})
		return
	}

	// Create JWT token
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    userInfo["id"],
		"email": userInfo["email"],
		"name":  userInfo["name"],
		"exp":   time.Now().Add(time.Hour * 72).Unix(), // token expires in 72h
	})

	tokenString, err := jwtToken.SignedString(jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate JWT"})
		return
	}

	// Redirect to frontend with JWT token as query param
	frontendURL := os.Getenv("FRONTEND_URL")
	redirectURL := frontendURL + "?token=" + url.QueryEscape(tokenString)
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

func GetLoginURL(c *gin.Context) {
	cfg := getGoogleOAuthConfig()
	loginURL := cfg.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	c.JSON(http.StatusOK, gin.H{
		"login_url": loginURL,
	})
}
