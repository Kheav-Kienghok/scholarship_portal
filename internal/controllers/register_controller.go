package controllers

import (
	"database/sql"
	"net/http"
	"strings"

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

	params := db.CreateUserParams{
		Fullname:     sql.NullString{String: input.Fullname, Valid: input.Fullname != ""},
		Email:        strings.ToLower(input.Email),
		PasswordHash: sql.NullString{String: string(hashedPassword), Valid: true},
	}

	_, err = r.Queries.CreateUser(c, params)
	if err != nil {
		logging.Error("DB: Failed to create user:", err)
		utils.JSONIndent(c, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}

	utils.JSONIndent(c, http.StatusCreated, "Registration successful", nil)
}
