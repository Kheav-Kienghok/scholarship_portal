package models

type RegisterInput struct {
	Fullname    string `json:"fullname"`
	DiplomaYear int    `json:"diploma_year"`
	Email       string `json:"email"`
	GradeLevel  int    `json:"grade_level"`
	HighSchool  string `json:"high_school"`
	Password    string `json:"password"`
	PhoneNumber string `json:"phone_number"`
}
