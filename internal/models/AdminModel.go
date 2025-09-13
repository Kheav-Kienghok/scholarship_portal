package models

type AdminRequestInput struct {
	Fullname *string `json:"fullname,omitempty" example:"Admin User"`
	Email    string  `json:"email" binding:"required,email" example:"admin@example.com"`
	Password string  `json:"password" binding:"required" example:"yourpassword"`
	OTP      string  `json:"otp,omitempty" example:"123456"`
}

type AdminLoginInput struct {
	Email    string `json:"email" binding:"required,email" example:"admin@example.com"`
	Password string `json:"password" binding:"required" example:"yourpassword"`
}

type AdminOTPInput struct {
	OTP string `json:"otp" binding:"required" example:"123456"`
}

type Verify2FAInput struct {
	Email string `json:"email"`
	OTP   string `json:"otp" binding:"required"`
}
