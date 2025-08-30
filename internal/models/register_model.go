package models

import "time"

type Timestamps struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RegisterModel struct {
	ID           int     `json:"id,omitempty"`
	Fullname     string  `json:"fullname" binding:"required"`
	Email        string  `json:"email" binding:"required,email"`
	Password     string  `json:"password" binding:"required"`
	Role         *string `json:"role,omitempty"` // Optional field
	PhoneNumber  string  `json:"phone_number" binding:"required"`
	HighSchool   string  `json:"high_school" binding:"required"`
	GradeLevel   int     `json:"grade_level"`
	DiplomaYear  int     `json:"diploma_year"`
	DiplomaGrade string  `json:"diploma_grade"`
	Timestamps
}

type RegisterResponse struct {
	Fullname string `json:"fullname"`
	Email    string `json:"email"`
	Timestamps
}
