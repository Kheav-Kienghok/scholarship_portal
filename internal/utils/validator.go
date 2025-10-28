package utils

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func ValidateEmail(email string) bool {
	if len(email) < 3 || len(email) > 254 {
		return false
	}
	return emailRegex.MatchString(email)
}

func RegisterCustomValidators(v *validator.Validate) {
	_ = v.RegisterValidation("password", passwordValidator)
}

func passwordValidator(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if len(password) < 8 {
		return false
	}
	hasUppercase, _ := regexp.MatchString(`[A-Z]`, password)
	hasLowercase, _ := regexp.MatchString(`[a-z]`, password)

	return hasUppercase && hasLowercase
}

func ValidatePassword(password string) string {
	if len(password) < 8 {
		return "Password must be at least 8 characters long."
	}
	if ok, _ := regexp.MatchString(`[A-Z]`, password); !ok {
		return "Password must contain at least one uppercase letter."
	}
	if ok, _ := regexp.MatchString(`[a-z]`, password); !ok {
		return "Password must contain at least one lowercase letter."
	}
	return ""
}