package tokens

import (
	"time"

	"github.com/Kheav-Kienghok/scholarship_portal/internal/models"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("Scholarship") // Replace with your secure secret

// GenerateToken creates a JWT token for a user
func GenerateToken(id int32, fullname, email, role string) (string, error) {

	claims := models.Claims{
		ID:       int(id),
		Fullname: fullname,
		Email:    email,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ParseToken validates and parses a JWT token
func ParseToken(tokenStr string) (*models.Claims, error) {

	token, err := jwt.ParseWithClaims(tokenStr, &models.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*models.Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrTokenInvalidClaims
}
