package utils

import (
	"errors"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/database/db"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/models"
)

// HasUserFields checks if any user basic info fields are provided for update
func HasUserFields(input models.UpdateUserProfileRequest) bool {
	return input.Fullname != nil || input.PhoneNumber != nil
}

// HasStudentProfileFields checks if any student profile fields are provided for update
func HasStudentProfileFields(input models.UpdateUserProfileRequest) bool {
	return input.HighSchool != nil ||
		input.GradeLevel != nil ||
		input.DiplomaYear != nil ||
		input.DiplomaGrade != nil ||
		input.SelectMajors != nil
}

// HasExistingStudentProfile checks if a student profile already exists in the database
func HasExistingStudentProfile(profile db.GetUserWithStudentProfileRow) bool {
	return profile.HighSchool.Valid ||
		profile.GradeLevel.Valid ||
		profile.DiplomaYear.Valid ||
		profile.DiplomaGrade.Valid ||
		profile.SelectMajors.Valid
}

// ValidateUserProfileInput validates the user profile update request
func ValidateUserProfileInput(input models.UpdateUserProfileRequest) error {
	// Add any cross-field validation logic here
	// For example: check if diploma year is reasonable, etc.

	if input.DiplomaYear != nil && *input.DiplomaYear < 1900 {
		return errors.New("diploma year must be after 1900")
	}

	if input.DiplomaYear != nil && *input.DiplomaYear > 2030 {
		return errors.New("diploma year cannot be in the future")
	}

	return nil
}
