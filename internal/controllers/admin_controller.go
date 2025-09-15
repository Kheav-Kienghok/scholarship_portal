package controllers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/database/db"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/logging"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/models"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/storage"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/tokens"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/utils"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
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

func getEmailFromJWT(c *gin.Context) (string, error) {

	claimsVal, exists := c.Get("claims")
	if !exists {
		return "", errors.New("claims not found in context")
	}

	adminClaims, ok := claimsVal.(tokens.ClaimsInterface)
	if !ok {
		return "", errors.New("invalid claims type")
	}

	return adminClaims.GetEmail(), nil
}

func (ctrl *AdminController) getAdminFromJWT(c *gin.Context) (*db.Admin, error) {
	email, err := getEmailFromJWT(c)
	if err != nil {
		return nil, err
	}

	admin, err := ctrl.Queries.GetAdminByIDOrEmail(c, db.GetAdminByIDOrEmailParams{
		Email: email,
	})
	if err != nil {
		return nil, err
	}

	return &admin, nil
}

func validateAdminOTP(admin *db.Admin, otp string) error {
	if !admin.TotpSecret.Valid || admin.TotpSecret.String == "" {
		return errors.New("2FA not enabled")
	}
	valid, err := utils.ValidateTOTP(otp, admin.TotpSecret.String)
	if err != nil {
		return fmt.Errorf("failed to validate OTP: %w", err)
	}
	if !valid {
		return errors.New("invalid OTP")
	}
	return nil
}

func (ctrl *AdminController) AdminLogin(c *gin.Context) {

	data, err := c.GetRawData()
	if err != nil || len(data) == 0 {
		utils.RespondBadRequest(c, "Empty request body", nil)
		return
	}

	// Try OTP first
	var otpInput models.AdminOTPInput
	if err := json.Unmarshal(data, &otpInput); err == nil && otpInput.OTP != "" {
		ctrl.handleStep2(c, otpInput)
		return
	}

	// Otherwise try login
	var loginInput models.AdminLoginInput
	if err := json.Unmarshal(data, &loginInput); err != nil {
		logging.Error("Failed to parse login input: ", err)
		utils.RespondBadRequest(c, "Invalid input", err.Error())
		return
	}

	ctrl.handleStep1(c, loginInput)
}

func (ctrl *AdminController) handleStep1(c *gin.Context, loginInput models.AdminLoginInput) {

	tx, err := ctrl.DB.BeginTx(c, &sql.TxOptions{})
	if err != nil {
		utils.RespondInternalError(c, "Failed to start transaction")
		return
	}
	defer tx.Rollback()
	qtx := ctrl.Queries.WithTx(tx)

	admin, err := qtx.GetAdminByIDOrEmail(c, db.GetAdminByIDOrEmailParams{Email: loginInput.Email})
	if err != nil || bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(loginInput.Password)) != nil {
		utils.RespondUnauthorized(c, "Incorrect Credential")
		return
	}

	if !admin.IsTwoFactor {
		logging.Info("2FA not enabled, issuing setup token")
		setupToken, _ := tokens.GenerateSetupToken(admin.Email)
		utils.RespondOK(c, "Required to setup 2 Factor Auth", gin.H{
			"setup_token": setupToken,
			"next":        "/admin/enable-2fa",
		})
		return
	}

	tempToken, _ := tokens.GenerateTempToken(admin.Email)
	utils.RespondOK(c, "OTP required", gin.H{
		"temp_token": tempToken,
	})
}

func (ctrl *AdminController) handleStep2(c *gin.Context, otpInput models.AdminOTPInput) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		utils.RespondUnauthorized(c, "Missing temp token")
		return
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := tokens.ParseTempToken(tokenStr)
	if err != nil || claims.GetPurpose() != "admin_2fa" {
		utils.RespondUnauthorized(c, "Invalid temp token")
		return
	}

	admin, err := ctrl.Queries.GetAdminByIDOrEmail(c, db.GetAdminByIDOrEmailParams{Email: claims.GetEmail()})
	if err != nil {
		utils.RespondUnauthorized(c, "Incorrect Credential")
		return
	}

	if err := validateAdminOTP(&admin, otpInput.OTP); err != nil {
		utils.RespondUnauthorized(c, err.Error())
		return
	}

	token, err := tokens.GenerateToken(admin.ID, admin.Fullname.String, admin.Email, "admin")
	if err != nil {
		utils.RespondInternalError(c, "Could not generate token")
		return
	}

	utils.RespondOK(c, "Login successful", gin.H{"token": token})
}

func (ctrl *AdminController) Enable2FAForAdmin(c *gin.Context) {

	admin, err := ctrl.getAdminFromJWT(c)
	if err != nil {
		utils.RespondUnauthorized(c, "Something went wrong!")
		return
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "ScholarshipPortal",
		AccountName: admin.Email,
	})
	if err != nil {
		utils.RespondInternalError(c, "Something went wrong!")
		return
	}

	if err := ctrl.Queries.AdminUpdateUserTOTPSecret(c, db.AdminUpdateUserTOTPSecretParams{
		ID:         admin.ID,
		TotpSecret: sql.NullString{String: key.Secret(), Valid: true},
	}); err != nil {
		utils.RespondInternalError(c, "Something went wrong!")
		return
	}

	png, err := qrcode.Encode(key.URL(), qrcode.Medium, 256)
	if err != nil {
		utils.RespondInternalError(c, "Failed to generate QR code")
		return
	}

	qrKey := fmt.Sprintf("qr_code/2fa_qr_codes/%d.png", admin.ID)
	_, err = storage.S3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: &storage.BucketName,
		Key:    &qrKey,
		Body:   bytes.NewReader(png),
	})
	if err != nil {
		utils.RespondInternalError(c, "Failed to upload images")
		return
	}

	qrURL, err := utils.GenerateQRCodeURL(storage.BucketName, qrKey, storage.S3Client)
	if err != nil {
		utils.RespondInternalError(c, "Failed to generate presigned QR code URL")
		return
	}

	utils.RespondOK(c, "2FA enabled", gin.H{"qr_code_url": qrURL})
}

func (ctrl *AdminController) Verify2FAForAdmin(c *gin.Context) {
	admin, err := ctrl.getAdminFromJWT(c)
	if err != nil {
		utils.RespondUnauthorized(c, "Unauthorized")
		return
	}

	var input models.Verify2FAInput
	if !utils.BindJSONOrFail(c, &input) {
		return
	}

	if err := validateAdminOTP(admin, input.OTP); err != nil {
		utils.RespondUnauthorized(c, err.Error())
		return
	}

	if err := ctrl.Queries.EnableAdmin2FA(c, admin.ID); err != nil {
		utils.RespondInternalError(c, "Failed to enable 2FA")
		return
	}

	utils.RespondOK(c, "2FA verified successfully", nil)
}
