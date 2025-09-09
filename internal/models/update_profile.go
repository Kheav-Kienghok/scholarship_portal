package models

type UpdateUserRequest struct {
	Fullname     *string  `json:"fullname"`
	PhoneNumber  *string  `json:"phone_number"`
	HighSchool   *string  `json:"high_school"`
	GradeLevel   *int32   `json:"grade_level"`
	DiplomaYear  *int32   `json:"diploma_year"`
	DiplomaGrade *string  `json:"diploma_grade"`
	SelectMajors []string `json:"select_majors"`
}
