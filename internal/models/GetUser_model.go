package models

import "time"

type StudentProfile struct {
	DiplomaGrade string   `json:"diploma_grade" example:"A+"`
	SelectMajors []string `json:"select_majors" example:"[\"Computer Science\", \"Law\"]"`
}

type UserProfileResponse struct {
	ID               int32          `json:"id" example:"1"`
	Fullname         string         `json:"fullname" example:"John Doe"`
	Email            string         `json:"email" example:"john.doe@example.com"`
	Role             string         `json:"role" example:"student"`
	PhoneNumber      string         `json:"phone_number" example:"123-456-7890"`
	HighSchool       string         `json:"high_school" example:"ABC High School"`
	GradeLevel       int            `json:"grade_level" example:"12"`
	DiplomaYear      int            `json:"diploma_year" example:"2023"`
	StudentProfile   StudentProfile `json:"student_profile"`
	ProfileCreatedAt time.Time      `json:"profile_created_at" example:"2023-01-01T00:00:00Z"`
	ProfileUpdatedAt time.Time      `json:"profile_updated_at" example:"2023-01-01T00:00:00Z"`
}
