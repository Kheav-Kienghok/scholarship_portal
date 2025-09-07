package models

// type LoginInput struct {
// 	Fullname string `json:"fullname"`
// 	Email    string `json:"email" binding:"required,email"`
// 	Password string `json:"password" binding:"required"`
// 	Role     string `json:"role"`
// }

// LoginRequest represents what the client sends for login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"johndoe@example.com"`
	Password string `json:"password" binding:"required" example:"supersecret"`
}


// LoginResponse represents the JWT token returned after successful login
type LoginResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}