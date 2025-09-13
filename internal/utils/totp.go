package utils

import (
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// ValidateTOTP validates a TOTP code with standard options.
func ValidateTOTP(otpCode, secret string) (bool, error) {
	return totp.ValidateCustom(
		otpCode,
		secret,
		time.Now().UTC(),
		totp.ValidateOpts{
			Period:    30,
			Skew:      1,
			Digits:    otp.DigitsSix,
			Algorithm: otp.AlgorithmSHA1,
		},
	)
}
