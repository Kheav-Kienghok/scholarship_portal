package models

import "github.com/golang-jwt/jwt/v5"

// Claims defines the JWT claims structure
type Claims struct {
	ID       int    `json:"id"`
	Fullname string `json:"fullname"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}
