package controllers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/database/db"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/logging"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/models"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// RegisterController handles registration requests
type RegisterController struct {
	Queries *db.Queries
}

func RegisterControllerHandler(queries *db.Queries) *RegisterController {
	return &RegisterController{
		Queries: queries,
	}
}

// generateVerificationToken creates a random verification token
func generateVerificationToken() (string, error) {
	bytes := make([]byte, 128)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Register godoc
// @Summary Register a new user
// @Description Create a new student account with email, password, and profile info
// @Tags Authentication
// @Accept json
// @Produce json
// @Param body body models.RegisterInput true "User registration payload"
// @Success 201 {object} utils.APIResponse "Registration successful"
// @Router /register [post]
func (r *RegisterController) Register(c *gin.Context) {

	var input models.RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.JSONIndent(c, http.StatusBadRequest, "Invalid input", err.Error())
		return
	}

	validateEmail := utils.ValidateEmail(input.Email)
	if !validateEmail {
		utils.JSONIndent(c, http.StatusBadRequest, "Invalid email format", nil)
		return
	}

	validatePassword := utils.ValidatePassword(input.Password)
	if !validatePassword {
		utils.JSONIndent(c, http.StatusBadRequest, "Password must be at least 6 characters", nil)
		return
	}

	// Check if user already exists
	_, err := r.Queries.GetUserByEmail(c, input.Email)
	if err != nil && err != sql.ErrNoRows {
		utils.JSONIndent(c, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}

	if err == nil {
		utils.JSONIndent(c, http.StatusBadRequest, "Email is already registered", nil)
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.JSONIndent(c, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}

	// Create user (initially unverified)
	params := db.CreateUserParams{
		Fullname:      sql.NullString{String: input.Fullname, Valid: input.Fullname != ""},
		Email:         strings.ToLower(input.Email),
		PasswordHash:  sql.NullString{String: string(hashedPassword), Valid: true},
		EmailVerified: sql.NullBool{Bool: false, Valid: true}, // Set to false initially
	}

	user, err := r.Queries.CreateUser(c, params)
	if err != nil {
		logging.Error("DB: Failed to create user:", err)
		utils.JSONIndent(c, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}

	// Generate verification token
	token, err := generateVerificationToken()
	if err != nil {
		logging.Error("Failed to generate verification token:", err)
		utils.JSONIndent(c, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}

	// Save verification token to database
	expiresAt := time.Now().Add(24 * time.Hour)

	_, err = r.Queries.CreateEmailVerification(c, db.CreateEmailVerificationParams{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		logging.Error("Failed to save verification token:", err)
		utils.JSONIndent(c, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}

	// Generate verification link
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "https://eduvision.live"
	}
	verificationLink := fmt.Sprintf("%s/api/v1/auth/verify-email?token=%s", baseURL, token)

	// Return success with verification link
	utils.JSONIndent(c, http.StatusCreated, "Please verify your email to activate your account.", gin.H{
		"verification": gin.H{
			"required": true,
			"link":     verificationLink,
			"expires":  "24h",
		},
	})
}

// VerifyEmail godoc
// @Summary Verify user email
// @Description Verify user email using the token sent via email
// @Tags Authentication
// @Produce json
// @Param token query string true "Verification token"
// @Success 200 {object} utils.APIResponse "Email verified successfully"
// @Router /auth/verify-email [get]
func (r *RegisterController) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		utils.JSONIndent(c, http.StatusBadRequest, "Verification token is required", nil)
		return
	}

	// Get verification record
	verification, err := r.Queries.GetEmailVerificationByToken(c, token)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.JSONIndent(c, http.StatusBadRequest, "Invalid or expired verification token", nil)
			return
		}
		logging.Error("Failed to get verification token:", err)
		utils.JSONIndent(c, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}

	// Check if token is expired
	if time.Now().After(verification.ExpiresAt) {
		utils.JSONIndent(c, http.StatusBadRequest, "Verification token has expired", nil)
		return
	}

	// Check if already verified
	if verification.VerifiedAt.Valid {
		utils.JSONIndent(c, http.StatusBadRequest, "Email is already verified", nil)
		return
	}

	// Verify the user
	err = r.Queries.VerifyUserEmail(c, verification.UserID)
	if err != nil {
		logging.Error("Failed to verify user email:", err)
		utils.JSONIndent(c, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}

	// Mark verification as completed
	err = r.Queries.MarkEmailVerificationAsUsed(c, verification.ID)
	if err != nil {
		logging.Error("Failed to mark verification as used:", err)
		// Don't fail the verification, just log
	}

	utils.JSONIndent(c, http.StatusOK, "Email verified successfully! You can now log in.", nil)
}
