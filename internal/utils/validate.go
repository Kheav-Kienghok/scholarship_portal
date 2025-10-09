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
    // Clean up whitespace and common symbols
    phone = strings.TrimSpace(phone)
    phone = strings.ReplaceAll(phone, " ", "")
    phone = strings.ReplaceAll(phone, "-", "")
    phone = strings.ReplaceAll(phone, "(", "")
    phone = strings.ReplaceAll(phone, ")", "")
    phone = strings.ReplaceAll(phone, ".", "")

    // Remove any existing country code formats and normalize
    if strings.HasPrefix(phone, "+855") {
        phone = phone[4:] // Remove +855, keep the rest
    } else if strings.HasPrefix(phone, "855") {
        phone = phone[3:] // Remove 855
    } else if strings.HasPrefix(phone, "0") {
        phone = phone[1:] // Remove leading 0
    }

    // Remove any non-digit characters
    re := regexp.MustCompile(`\D`)
    phone = re.ReplaceAllString(phone, "")

    // For your specific case: +855171467914 becomes 171467914 (9 digits)
    // Cambodia mobile numbers can be 8-9 digits after country code
    if len(phone) < 8 || len(phone) > 9 {
        return "" // Invalid length
    }

    // Format based on length
    if len(phone) == 8 {
        // Format: +855-XX-XXX-XXX (like +855-12-345-678)
        return fmt.Sprintf("+855-%s-%s-%s", phone[:2], phone[2:5], phone[5:8])
    } else if len(phone) == 9 {
        // Format: +855-XXX-XXX-XXX (like +855-171-467-914)
        return fmt.Sprintf("+855-%s-%s-%s", phone[:3], phone[3:6], phone[6:9])
    }

    return ""
}

func ValidatePhoneNumber(phone string) bool {
    // Accepts both formats:
    // +855-XX-XXX-XXX (8 digits: 2-3-3 pattern)
    // +855-XXX-XXX-XXX (9 digits: 3-3-3 pattern)
    phoneRegex := regexp.MustCompile(`^\+855-([0-9]{2}-[0-9]{3}-[0-9]{3}|[0-9]{3}-[0-9]{3}-[0-9]{3})$`)
    return phoneRegex.MatchString(phone)
}