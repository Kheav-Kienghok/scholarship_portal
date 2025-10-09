package models

import "time"

type StudentProfileRequest struct {
    HighSchool   *string  `json:"high_school,omitempty"`
    GradeLevel   *string  `json:"grade_level,omitempty"`
    DiplomaYear  *int32   `json:"diploma_year,omitempty"`
    DiplomaGrade *string  `json:"diploma_grade,omitempty"`
    SelectMajors []string `json:"select_majors,omitempty"`
}

type StudentProfileResponse struct {
    HighSchool   string    `json:"high_school,omitempty"`
    GradeLevel   string    `json:"grade_level,omitempty"`
    DiplomaYear  int32     `json:"diploma_year,omitempty"`
    DiplomaGrade string    `json:"diploma_grade,omitempty"`
    SelectMajors []string  `json:"select_majors,omitempty"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}

type UserWithStudentProfileResponse struct {
    ID          int32                   `json:"id"`
    Fullname    string                  `json:"fullname"`
    Email       string                  `json:"email"`
    PhoneNumber string                  `json:"phone_number"`
    CreatedAt   time.Time               `json:"created_at"`
    UpdatedAt   time.Time               `json:"updated_at"`
    Profile     *StudentProfileResponse `json:"student_profile,omitempty"`
}

type UpdateUserProfileRequest struct {
    // User basic info
    Fullname    *string `json:"fullname,omitempty"`
    PhoneNumber *string `json:"phone_number,omitempty"`
    
    // Student profile info
    HighSchool   *string  `json:"high_school,omitempty"`
    GradeLevel   *string  `json:"grade_level,omitempty"`
    DiplomaYear  *int32   `json:"diploma_year,omitempty"`
    DiplomaGrade *string  `json:"diploma_grade,omitempty"`
    SelectMajors []string `json:"select_majors,omitempty"`
}

type UnifiedUserProfileResponse struct {
    // User basic info
    ID          int32  `json:"id"`
    Fullname    string `json:"fullname"`
    Email       string `json:"email"`
    PhoneNumber string `json:"phone_number"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
    
    // Student profile (optional)
    StudentProfile *StudentProfileResponse `json:"student_profile,omitempty"`
}