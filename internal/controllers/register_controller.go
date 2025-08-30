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

// HomeController handles home page requests
type RegisterController struct {
	DB *database.Database
}

// NewRegisterController creates a new register controller instance
func RegisterControllerHandler(db *database.Database) *RegisterController {
	return &RegisterController{
		DB: db,
	}
}

func (r *RegisterController) Register(c *gin.Context) {

	logging.LogRequest(c.Request.Method, c.Request.URL.Path, c.ClientIP(), http.StatusOK)

	var input models.RegisterModel
	if err := c.ShouldBindJSON(&input); err != nil {
		logging.Error("Invalid input: " + err.Error())
		utils.JSONIndent(c, http.StatusBadRequest, "Invalid input", err.Error())
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.JSONIndent(c, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}
	input.Password = string(hashedPassword)

	if err := r.DB.CreateUser(&input); err != nil {
		logging.Error("Failed to create user: " + err.Error())
		utils.JSONIndent(c, http.StatusInternalServerError, "Something went wrong", err.Error())
		return
	}

	resp := models.RegisterResponse{
		Fullname: input.Fullname,
		Email:    input.Email,
		Timestamps: models.Timestamps{
			CreatedAt: input.CreatedAt,
			UpdatedAt: input.UpdatedAt,
		},
	}

	utils.JSONIndent(c, http.StatusOK, "Registration successful", resp)

}
