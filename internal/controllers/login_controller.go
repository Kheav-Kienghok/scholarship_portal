package controllers

import (
	"net/http"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/database/db"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/logging"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/models"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/tokens"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// LoginController handles login requests
type LoginController struct {
	Queries *db.Queries
}

func LoginControllerHandler(queries *db.Queries) *LoginController {
	return &LoginController{
		Queries: queries,
	}
}

// Login godoc
// @Summary Login into user account
// @Tags Authentication
// @Accept json
// @Produce json
// @Param body body models.LoginRequest true "Login user"
// @Success 200 {object} utils.Response{data=models.LoginResponse} "Login successful"
// @Router /login [post]
func (ctrl *LoginController) Login(c *gin.Context) {
	var input models.LoginRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.JSONIndent(c, http.StatusBadRequest, "Invalid input", err.Error())
		return
	}

	user, err := ctrl.Queries.GetUserByEmail(c, input.Email)
	if err != nil {
		utils.JSONIndent(c, http.StatusUnauthorized, "Incorrect Credential", nil)
		return
	}

	if !user.EmailVerified.Bool {
		utils.JSONIndent(c, http.StatusUnauthorized, "Please verify your email before logging in", nil)
		return
	}

	// Compare hashed password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String), []byte(input.Password)); err != nil {
		go logging.Error("Password mismatch:", err)
		utils.JSONIndent(c, http.StatusUnauthorized, "Incorrect Credential", nil)
		return
	}

	token, err := tokens.GenerateToken(user.ID, user.Fullname.String, user.Email, "student")
	if err != nil {
		utils.JSONIndent(c, http.StatusInternalServerError, "Could not generate token", nil)
		return
	}	

	loginResponse := models.LoginResponse{
		Token: token,
	}

	utils.JSONIndent(c, http.StatusOK, "Login successful", loginResponse)
}
