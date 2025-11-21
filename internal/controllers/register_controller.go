package controllers

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/database/db"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/logging"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/models"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/tokens"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

// RegisterController handles registration requests
type RegisterController struct {
	Queries *db.Queries
	DB      *sql.DB
}

func RegisterControllerHandler(dbConn *sql.DB, queries *db.Queries) *RegisterController {
	return &RegisterController{
		Queries: queries,
		DB:      dbConn,
	}
}

// generateVerificationToken creates a random verification token
func generateVerificationToken() (string, error) {
	bytes := make([]byte, 128)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
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

	// Validate input
	if err := c.ShouldBindJSON(&input); err != nil {
		r.handleValidationError(c, err)
		return
	}

	// Check if user exists
	if err := r.checkUserExists(c, input.Email); err != nil {
		return
	}

	// Start transaction
	tx, err := r.DB.BeginTx(c, nil)
	if err != nil {
		utils.JSONIndent(c, http.StatusInternalServerError, "Failed to start transaction", nil)
		return
	}
	defer func() {
		// rollback if still open (i.e., not committed)
		_ = tx.Rollback()
	}()

	qtx := r.Queries.WithTx(tx)

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
		EmailVerified: sql.NullBool{Bool: false, Valid: true},
	}

	user, err := qtx.CreateUser(c, params)
	if err != nil {
		go logging.Error("DB: Failed to create user:", err)
		utils.JSONIndent(c, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}

	// Generate verification token
	token, err := generateVerificationToken()
	if err != nil {
		go logging.Error("Failed to generate verification token:", err)
		utils.JSONIndent(c, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}

	// Save verification token to database
	expiresAt := time.Now().Add(1 * time.Hour)

	_, err = qtx.CreateEmailVerification(c, db.CreateEmailVerificationParams{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		go logging.Error("Failed to save verification token:", err)
		utils.JSONIndent(c, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}

	// Commit transaction first (user + token must be consistent)
	if err := tx.Commit(); err != nil {
		go logging.Error("Failed to commit transaction:", err)
		utils.JSONIndent(c, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}

	// Generate verification link
	verificationLink := generateVerificationLink(token)

	// Uncomment this to send link
	err = utils.SendVerificationEmail(c, user.Email, user.Fullname.String, verificationLink)
	if err != nil {
		go logging.Error("Failed to send verification email:", err)
		utils.JSONIndent(c, http.StatusInternalServerError, "Could not send verification email", nil)
		return
	}

	// Return success with verification link
	utils.JSONIndent(c, http.StatusCreated, "Please verify your email to activate your account.", nil)
}

func (r *RegisterController) handleValidationError(c *gin.Context, err error) {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		utils.JSONIndent(c, http.StatusBadRequest, utils.TranslateValidationError(ve), nil)
		return
	}

	var ute *json.UnmarshalTypeError
	if errors.As(err, &ute) {
		utils.JSONIndent(c, http.StatusBadRequest,
			fmt.Sprintf("Invalid type for field %s", ute.Field), nil)
		return
	}

	utils.JSONIndent(c, http.StatusBadRequest, "Invalid input", err.Error())
}

func (r *RegisterController) checkUserExists(c *gin.Context, email string) error {
	existingUser, err := r.Queries.CheckUserExistByEmail(c, email)
	if err != nil && err != sql.ErrNoRows {
		utils.JSONIndent(c, http.StatusInternalServerError, "Something went wrong", nil)
		return err
	}

	if err == nil {
		msg := "Email is already registered"
		if !existingUser.EmailVerified.Bool {
			msg = "Email is not verified yet. Please check your inbox."
		}
		utils.JSONIndent(c, http.StatusBadRequest, msg, nil)
		return errors.New("user exists")
	}

	return nil
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
		go logging.Error("Failed to get verification token:", err)
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
		go logging.Error("Failed to verify user email:", err)
		utils.JSONIndent(c, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}

	// Mark verification as completed
	err = r.Queries.MarkEmailVerificationAsUsed(c, verification.ID)
	if err != nil {
		go logging.Error("Failed to mark verification as used:", err)
	}

	// Fetch the user to generate JWT
	user, err := r.Queries.GetUserByID(c, verification.UserID)
	if err != nil {
		go logging.Error("Failed to fetch user:", err)
		utils.JSONIndent(c, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}

	// Generate JWT token
	jwtToken, err := tokens.GenerateToken(user.ID, user.Fullname.String, user.Email, "student")
	if err != nil {
		utils.JSONIndent(c, http.StatusInternalServerError, "Could not generate token", nil)
		return
	}

	// Set JWT as HttpOnly cookie
	c.SetCookie(
		"jwt",             // cookie name
		jwtToken,          // cookie value
		3600*12,           // max age in seconds (e.g., 1 day)
		"/",               // path
		".eduvision.live", // domain (empty = current domain)
		true,              // secure (true if using HTTPS)
		false,             // HttpOnly (JS cannot access)
	)

	// c.Redirect(http.StatusSeeOther, "https://www.eduvision.live")
	// utils.JSONIndent(c, http.StatusOK, "Email verified successfully! You can now log in.", nil)

	c.Redirect(http.StatusSeeOther, "https://www.eduvision.live")
}

// ResendVerification godoc
// @Summary Resend verification email
// @Description Resend verification email for unverified users
// @Tags Authentication
// @Accept json
// @Produce json
// @Param body body models.ResendVerificationInput true "Email address"
// @Success 200 {object} utils.APIResponse "Verification email sent"
// @Router /auth/resend-verification [post]
func (r *RegisterController) ResendVerification(c *gin.Context) {
	var input models.ResendVerificationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.JSONIndent(c, http.StatusBadRequest, "Invalid input", err.Error())
		return
	}

	// Validate email format
	if !utils.ValidateEmail(input.Email) {
		utils.JSONIndent(c, http.StatusBadRequest, "Invalid email format", nil)
		return
	}

	// Normalize email
	email := strings.ToLower(input.Email)

	// Check if user exists and is unverified
	user, err := r.Queries.GetUnverifiedUserByEmail(c, email)
	if err != nil {
		if err == sql.ErrNoRows {
			// Don't reveal if email exists or not for security
			utils.JSONIndent(c, http.StatusOK, "If this email is registered and unverified, a verification link will be sent.", nil)
			return
		}
		go logging.Error("Failed to get user:", err)
		utils.JSONIndent(c, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}

	// Check if already verified
	if user.EmailVerified.Valid && user.EmailVerified.Bool {
		utils.JSONIndent(c, http.StatusBadRequest, "Email is already verified", nil)
		return
	}

	// Check for rate limiting (prevent spam)
	latestVerification, err := r.Queries.GetLatestVerificationByEmail(c, email)
	if err != nil && err != sql.ErrNoRows {
		go logging.Error("Failed to get latest verification:", err)
		utils.JSONIndent(c, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}

	// If a verification was sent within the last 5 minutes, reject
	if err == nil && latestVerification.CreatedAt.Valid && time.Since(latestVerification.CreatedAt.Time) < 5*time.Minute {
		remainingTime := 5*time.Minute - time.Since(latestVerification.CreatedAt.Time)
		utils.JSONIndent(c, http.StatusTooManyRequests,
			fmt.Sprintf("Please wait %d seconds before requesting another verification email", int(remainingTime.Seconds())),
			nil)
		return
	}

	// Delete old unverified tokens for this user
	err = r.Queries.DeleteExpiredVerifications(c, user.ID)
	if err != nil {
		go logging.Error("Failed to delete old verifications:", err)
		// Continue anyway
	}

	// Generate new verification token
	token, err := generateVerificationToken()
	if err != nil {
		go logging.Error("Failed to generate verification token:", err)
		utils.JSONIndent(c, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}

	// Save new verification token
	expiresAt := time.Now().Add(24 * time.Hour)
	_, err = r.Queries.CreateEmailVerification(c, db.CreateEmailVerificationParams{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		go logging.Error("Failed to save verification token:", err)
		utils.JSONIndent(c, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}

	// Generate verification link
	// verificationLink := generateVerificationLink(token)

	utils.JSONIndent(c, http.StatusOK, "Verification email has been resent. Please check your inbox.", nil)
}

// Helper function to generate verification link
func generateVerificationLink(token string) string {
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "https://eduvision.live"
	}
	return fmt.Sprintf("%s/verify-email?token=%s", baseURL, token)
}
