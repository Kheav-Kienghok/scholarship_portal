package controllers

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/database/db"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/logging"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/models"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/storage"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/tokens"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/utils"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"
	"golang.org/x/crypto/bcrypt"
)

type AdminController struct {
	Queries *db.Queries
	DB      *sql.DB
}

func AdminControllerHandler(dbConn *sql.DB, queries *db.Queries) *AdminController {
	return &AdminController{Queries: queries, DB: dbConn}
}

// Admin login with 2FA check
func (ctrl *AdminController) AdminLogin(c *gin.Context) {

	var input models.AdminRequestInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.JSONIndent(c, http.StatusBadRequest, "Invalid input", err.Error())
		return
	}

	tx, err := ctrl.DB.BeginTx(c, &sql.TxOptions{})
	if err != nil {
		utils.JSONIndent(c, http.StatusInternalServerError, "Failed to start transaction", nil)
		return
	}
	defer tx.Rollback() // rollback if not committed

	qtx := ctrl.Queries.WithTx(tx)

	admin, err := qtx.GetAdminByIDOrEmail(c, db.GetAdminByIDOrEmailParams{
		Email: input.Email,
	})
	if err != nil {
		utils.JSONIndent(c, http.StatusUnauthorized, "Incorrect Credential", nil)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(input.Password)); err != nil {
		utils.JSONIndent(c, http.StatusUnauthorized, "Incorrect Credential", nil)
		return
	}

	// Check if 2FA is enabled
	if !admin.IsTwoFactor {
		logging.Info("2FA not enabled, issuing setup token")
		setupToken, _ := tokens.GenerateSetupToken(admin.Email)
		utils.JSONIndent(c, http.StatusOK, "2FA setup required", gin.H{
			"setup_token": setupToken,
			"next":        "/admin/enable-2fa",
		})
		return
	}

	// 2FA enabled → require OTP
	if input.OTP == "" {
		utils.JSONIndent(c, http.StatusUnauthorized, "OTP required", nil)
		return
	}

	// Validate OTP
	valid, err := totp.ValidateCustom(
		input.OTP,
		admin.TotpSecret.String,
		time.Now().UTC(),
		totp.ValidateOpts{
			Period:    30,
			Skew:      1,
			Digits:    otp.DigitsSix,
			Algorithm: otp.AlgorithmSHA1,
		},
	)

	if err != nil {
		utils.JSONIndent(c, http.StatusInternalServerError, "Failed to validate OTP", nil)
		return
	}

	if !valid {
		utils.JSONIndent(c, http.StatusUnauthorized, "Invalid OTP", nil)
		return
	}

	// Generate JWT token (implement your own token logic)
	token, err := tokens.GenerateToken(admin.ID, admin.Fullname.String, admin.Email, "admin")
	if err != nil {
		utils.JSONIndent(c, http.StatusInternalServerError, "Could not generate token", nil)
		return
	}

	if err := tx.Commit(); err != nil {
		utils.JSONIndent(c, http.StatusInternalServerError, "Failed to commit transaction", nil)
		return
	}

	utils.JSONIndent(c, http.StatusOK, "Login successful", gin.H{
		"token": token,
	})
}


// Enable 2FA for admin
func (ctrl *AdminController) Enable2FAForAdmin(c *gin.Context) {

	email, err := getEmailFromJWT(c)
	if err != nil {
		logging.Error("Failed to get email from JWT:", err)
		utils.JSONIndent(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	admin, err := ctrl.Queries.GetAdminByIDOrEmail(c, db.GetAdminByIDOrEmailParams{
		Email: email,
	})
	if err != nil {
		utils.JSONIndent(c, http.StatusUnauthorized, "Something went wrong!", nil)
		return
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "ScholarshipPortal",
		AccountName: admin.Email,
	})
	if err != nil {
		utils.JSONIndent(c, http.StatusInternalServerError, "Failed to generate TOTP secret", nil)
		return
	}

	// Save secret to DB
	err = ctrl.Queries.AdminUpdateUserTOTPSecret(c, db.AdminUpdateUserTOTPSecretParams{
		ID:         admin.ID,
		TotpSecret: sql.NullString{String: key.Secret(), Valid: true},
	})
	if err != nil {
		utils.JSONIndent(c, http.StatusInternalServerError, "Failed to save TOTP secret", nil)
		return
	}

	// Generate QR code as base64
	png, err := qrcode.Encode(key.URL(), qrcode.Medium, 256)
	if err != nil {
		utils.JSONIndent(c, http.StatusInternalServerError, "Failed to generate QR code", nil)
		return
	}

	// upload QR code to S3
	qrKey := fmt.Sprintf("qr_code/2fa_qr_codes/%d.png", admin.ID)
	_, err = storage.S3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: &storage.BucketName,
		Key:    &qrKey,
		Body:   bytes.NewReader(png),
	})
	if err != nil {
		utils.JSONIndent(c, http.StatusInternalServerError, "Failed to upload QR code to S3", nil)
		return
	}

	// generate presigned URL (15 min expiry)
	qrURL, err := utils.GenerateQRCodeURL(storage.BucketName, qrKey, storage.S3Client)
	if err != nil {
		utils.JSONIndent(c, http.StatusInternalServerError, "Failed to generate presigned QR code URL", nil)
		return
	}

	utils.JSONIndent(c, http.StatusOK, "2FA enabled", gin.H{
		"qr_code_url": qrURL,
	})
}

// Verify 2FA for admin
func (ctrl *AdminController) Verify2FAForAdmin(c *gin.Context) {

	email, err := getEmailFromJWT(c)
	if err != nil {
		utils.JSONIndent(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var input models.Verify2FAInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.JSONIndent(c, http.StatusBadRequest, "Invalid input", err.Error())
		return
	}

	admin, err := ctrl.Queries.GetAdminByIDOrEmail(c, db.GetAdminByIDOrEmailParams{
		Email: email,
	})
	if err != nil {
		utils.JSONIndent(c, http.StatusUnauthorized, "Admin not found", nil)
		return
	}

	if !admin.TotpSecret.Valid || admin.TotpSecret.String == "" {
		utils.JSONIndent(c, http.StatusBadRequest, "2FA not enabled for this admin", nil)
		return
	}

	if !totp.Validate(input.OTP, admin.TotpSecret.String) {
		utils.JSONIndent(c, http.StatusUnauthorized, "Invalid OTP", nil)
		return
	}

	// Validate OTP
	valid, err := totp.ValidateCustom(
		input.OTP,
		admin.TotpSecret.String,
		time.Now().UTC(),
		totp.ValidateOpts{
			Period:    30,
			Skew:      1,
			Digits:    otp.DigitsSix,
			Algorithm: otp.AlgorithmSHA1,
		},
	)

	if err != nil {
		utils.JSONIndent(c, http.StatusInternalServerError, "Failed to validate OTP", nil)
		return
	}

	if !valid {
		utils.JSONIndent(c, http.StatusUnauthorized, "Invalid OTP", nil)
		return
	}

	// Mark 2FA as enabled
	err = ctrl.Queries.EnableAdmin2FA(c, admin.ID)
	if err != nil {
		utils.JSONIndent(c, http.StatusInternalServerError, "Failed to enable 2FA", nil)
		return
	}

	utils.JSONIndent(c, http.StatusOK, "2FA verified successfully", nil)
}

func getEmailFromJWT(c *gin.Context) (string, error) {

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", errors.New("missing Authorization header")
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := tokens.ParseToken(tokenStr)
	if err != nil {
		return "", err
	}
	return claims.GetEmail(), nil
}
