package controllers

import (
	"database/sql"
	"net/http"

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
// @Tags Users
// @Accept json
// @Produce json
// @Param body body models.RegisterInput true "Register user"
// @Success 201 {object} utils.APIResponse "Registration successful"
// @Router /register [post]
func (r *RegisterController) Register(c *gin.Context) {

	var input models.RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.JSONIndent(c, http.StatusBadRequest, "Invalid input", err.Error())
		return
	}

	// Check if user already exists
	_, err := r.Queries.GetUserByIDOrEmail(c, db.GetUserByIDOrEmailParams{
		Email: input.Email,
	})
	if err != nil && err != sql.ErrNoRows {
		utils.JSONIndent(c, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}

	if err == nil {
		utils.JSONIndent(c, http.StatusBadRequest, "User already exists", nil)
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
		DiplomaYear:  sql.NullInt32{Int32: int32(input.DiplomaYear), Valid: input.DiplomaYear != 0},
		Email:        input.Email,
		GradeLevel:   sql.NullInt32{Int32: int32(input.GradeLevel), Valid: input.GradeLevel != 0},
		HighSchool:   sql.NullString{String: input.HighSchool, Valid: input.HighSchool != ""},
		PasswordHash: sql.NullString{String: string(hashedPassword), Valid: true},
		PhoneNumber:  sql.NullString{String: input.PhoneNumber, Valid: input.PhoneNumber != ""},
	}

	user, err := r.Queries.CreateUser(c, params)
	if err != nil {
		logging.Error("DB: Failed to create user:", err)
		utils.JSONIndent(c, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}

	// Create empty student profile for this user
	err = r.Queries.CreateStudentProfile(c, user.ID)
	if err != nil {
		logging.Error("DB: Failed to create student profile:", err)
		utils.JSONIndent(c, http.StatusInternalServerError, "Something went wrong", nil)
		return
	}

	utils.JSONIndent(c, http.StatusCreated, "Registration successful", nil)
}
