package utils

import (
	"regexp"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
)

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	vTrans     ut.Translator
)

// InitValidator registers custom validators and translations.
func InitValidator(v *validator.Validate) error {
	// Default EN translations
	enLoc := en.New()
	uni := ut.New(enLoc, enLoc)
	t, _ := uni.GetTranslator("en")
	if err := enTranslations.RegisterDefaultTranslations(v, t); err != nil {
		return err
	}

	// Register strong password validator
	_ = v.RegisterValidation("strongpassword", validateStrongPassword)

	// Custom translation: build message based on actual value
	_ = v.RegisterTranslation("strongpassword", t,
		func(ut ut.Translator) error {
			// Key kept generic; message is computed dynamically in the translate func
			return ut.Add("strongpassword", "{0}", false)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			p, _ := fe.Value().(string)
			return passwordStrengthMessage(p)
		},
	)

	vTrans = t
	return nil
}

func validateStrongPassword(fl validator.FieldLevel) bool {
	p := fl.Field().String()
	if len(p) < 8 {
		return false
	}
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(p)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(p)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(p)
	return hasUpper && hasLower && hasDigit
}

func passwordStrengthMessage(p string) string {
	var missing []string
	if !regexp.MustCompile(`[a-z]`).MatchString(p) {
		missing = append(missing, "one lowercase letter")
	}
	if !regexp.MustCompile(`[A-Z]`).MatchString(p) {
		missing = append(missing, "one uppercase letter")
	}
	if !regexp.MustCompile(`[0-9]`).MatchString(p) {
		missing = append(missing, "one digit")
	}
	if len(p) < 8 {
		missing = append(missing, "be at least 8 characters long")
	}

	if len(missing) == 0 {
		return "Password does not meet strength requirements."
	}

	// Join with commas and 'and' for readability
	if len(missing) == 1 {
		return "Password must contain " + missing[0] + "."
	}
	if len(missing) == 2 {
		return "Password must contain " + missing[0] + " and " + missing[1] + "."
	}
	return "Password must contain " + strings.Join(missing[:len(missing)-1], ", ") + ", and " + missing[len(missing)-1] + "."
}

// TranslateValidationError returns a user-friendly message from validator errors.
func TranslateValidationError(err error) string {
	if vTrans == nil {
		return err.Error()
	}
	if verrs, ok := err.(validator.ValidationErrors); ok && len(verrs) > 0 {
		fe := verrs[0]
		if fe.Tag() == "strongpassword" {
			if p, ok := fe.Value().(string); ok {
				return passwordStrengthMessage(p)
			}
			return "Password must include uppercase, lowercase, a digit, and be at least 8 characters."
		}
		return fe.Translate(vTrans)
	}
	return err.Error()
}

// ValidateEmail checks if the provided email is valid.
func ValidateEmail(email string) bool {
	if len(email) < 3 || len(email) > 254 {
		return false
	}
	return emailRegex.MatchString(email)
}
