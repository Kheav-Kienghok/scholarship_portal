package models

type RegisterInput struct {
	Fullname    string `json:"fullname" example:"John Doe"`
	Email       string `json:"email" example:"john.doe@example.com"`
	Password    string `json:"password" example:"SuperSecret123"`
	PhoneNumber string `json:"phone_number" example:"+855-17-345-6790"`
	HighSchool  string `json:"high_school" example:"Springfield High"`
	GradeLevel  int    `json:"grade_level" example:"12"`
	DiplomaYear int    `json:"diploma_year" example:"2025"`
}
