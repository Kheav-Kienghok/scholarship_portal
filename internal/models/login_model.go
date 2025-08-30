package models

type LoginModel struct {
	Fullname string `json:"fullname"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role"`
}
