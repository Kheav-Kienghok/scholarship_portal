package models

type RegisterInput struct {
	Fullname    string `json:"fullname" binding:"required" example:"John Doe"`
	Email       string `json:"email" binding:"required,email" example:"john.doe@example.com"`
	Password    string `json:"password" binding:"required,strongpassword" example:"SuperSecret123"`
}

// ResendVerificationInput represents the request body for resending verification email
type ResendVerificationInput struct {
    Email string `json:"email" binding:"required,email"`
}