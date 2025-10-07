package utils

import (
	"fmt"
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func ValidateEmail(email string) bool {
	if len(email) < 3 || len(email) > 254 {
		return false
	}
	return emailRegex.MatchString(email)
}

func ValidatePassword(password string) bool {
	return len(password) >= 6
}

func NormalizeCambodianPhone(phone string) string {
	// Clean up whitespace and symbols
	phone = strings.TrimSpace(phone)
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")

	// Handle different possible prefixes
	if strings.HasPrefix(phone, "0") {
		phone = "+855" + phone[1:]
	} else if strings.HasPrefix(phone, "855") {
		phone = "+855" + phone[3:]
	} else if !strings.HasPrefix(phone, "+855") {
		phone = "+855" + phone
	}

	// Extract only digits after +855
	re := regexp.MustCompile(`^\+855(\d+)$`)
	matches := re.FindStringSubmatch(phone)
	if len(matches) != 2 {
		return phone // fallback, invalid format
	}

	digits := matches[1]

	// Enforce grouping: +855-XXX-XXX-XXX (truncate or leave as-is if shorter)
	if len(digits) >= 9 {
		return fmt.Sprintf("+855-%s-%s-%s", digits[:3], digits[3:6], digits[6:9])
	} else if len(digits) >= 6 {
		return fmt.Sprintf("+855-%s-%s-%s", digits[:3], digits[3:6], digits[6:])
	} else {
		return "+855-" + digits
	}
}

func ValidatePhoneNumber(phone string) bool {
	// Accepts +855-XX-XXX-XXXX or +855-XXX-XXX-XXX
	phoneRegex := regexp.MustCompile(`^\+855-[0-9]{2,3}-[0-9]{3}-[0-9]{3,4}$`)
	return phoneRegex.MatchString(phone)
}