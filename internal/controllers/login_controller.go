package controllers

import (
	"net/http"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/database"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/logging"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/models"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// LoginController handles login requests
type LoginController struct {
	DB *database.Database
}

// NewLoginController creates a new login controller instance
func LoginControllerHandler(db *database.Database) *LoginController {
	return &LoginController{
		DB: db,
	}
}

func (h *LoginController) Login(c *gin.Context) {

	logging.Info("Login attempt", "ip", c.ClientIP(), "path", c.Request.URL.Path)

	var input models.LoginModel
	if err := c.ShouldBindJSON(&input); err != nil {
		logging.Warn("Invalid login input", "ip", c.ClientIP(), "error", err.Error())
		utils.JSONIndent(c, http.StatusBadRequest, "Invalid request payload", err.Error())
		return
	}

	// Find user by email
	user, err := h.DB.FindUserByEmail(input.Email)
	if err != nil {
		logging.Warn("Login failed: user not found", "email", input.Email, "ip", c.ClientIP())
		utils.JSONIndent(c, http.StatusUnauthorized, "Invalid email or password", nil)
		return
	}

	// Compare password hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {

		logging.Warn("Login failed: wrong password", "email", input.Email, "ip", c.ClientIP())
		utils.JSONIndent(c, http.StatusUnauthorized, "Invalid email or password", nil)
		return
	}

	// Successful login
	logging.Info("User logged in", "email", user.Email, "role", user.Role, "ip", c.ClientIP())

	// TODO: Generate JWT or session here and return it instead of just message
	utils.JSONIndent(c, http.StatusOK, "Login successful", gin.H{
		"fullname": user.Fullname,
		"email":    user.Email,
		"role":     user.Role,
	})
}
