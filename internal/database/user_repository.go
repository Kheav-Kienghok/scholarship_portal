package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/logging"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/models"
)

// CreateUser inserts a new user into the database
func (d *Database) CreateUser(user *models.RegisterModel) error {

	// Set default role if not provided
	var role string
	if user.Role == nil || *user.Role == "" {
		role = "student"
	} else {
		role = *user.Role
	}

	query := `
		INSERT INTO users (fullname, email, password, role, phone_number, high_school, grade_level, diploma_year) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`

	err := d.DB.QueryRow(
		query,
		user.Fullname,
		user.Email,
		user.Password,
		role,
		user.PhoneNumber,
		user.HighSchool,
		user.GradeLevel,
		user.DiplomaYear,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		logging.Error(fmt.Sprintf("Failed to create user: %v", err))
		return err
	}

	logging.Info(fmt.Sprintf("User (%s) created successfully", user.Fullname))
	return nil
}

func (d *Database) FindUserByEmail(email string) (*models.LoginModel, error) {

	var user models.LoginModel
	query := `
		SELECT fullname, email, password, role 
		FROM users 
		WHERE email = $1`
	err := d.DB.QueryRow(query, email).Scan(&user.Fullname, &user.Email, &user.Password, &user.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("no user found with email: %s", email)
		}

		logging.Error(fmt.Sprintf("Failed to find user by email: %s, error: %v", email, err))
		return nil, fmt.Errorf("db query failed: %w", err)
	}
	return &user, nil
}
