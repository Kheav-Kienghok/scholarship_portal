package models

type RegisterInput struct {
	Fullname    string `json:"fullname" example:"John Doe"`
	Email       string `json:"email" example:"john.doe@example.com"`
	Password    string `json:"password" example:"SuperSecret123"`
	PhoneNumber string `json:"phone_number" example:"+855-17-345-6790"`
}

// ResendVerificationInput represents the request body for resending verification email
type ResendVerificationInput struct {
    Email string `json:"email" binding:"required,email"`
}