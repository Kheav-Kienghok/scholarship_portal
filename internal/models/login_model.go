package models

type LoginInput struct {
	Fullname string `json:"fullname"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role"`
}

type LoginResponse struct {
	Token string `json:"token"`
}
