package models

type LoginModel struct {
	Fullname string `json:"fullname"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role"`
}

type UpdatePasswordInput struct {
    Email       string `json:"email" binding:"required,email"`
    OldPassword string `json:"old_password" binding:"required"`
    NewPassword string `json:"new_password" binding:"required"`
}