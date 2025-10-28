package models

type AdminLoginInput struct {
	Email    string `json:"email" binding:"required,email" example:"admin@example.com"`
	Password string `json:"password" binding:"required" example:"yourpassword"`
}

type AdminOTPInput struct {
	OTP string `json:"otp" binding:"required,len=6,numeric" example:"123456"`
}

type Verify2FAInput struct {
	Email string `json:"email"`
	OTP   string `json:"otp" binding:"required,len=6,numeric" example:"123456"`
}
