package tokens

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AdminClaims struct {
	ID      int64  `json:"id,omitempty"`
	Email   string `json:"email"`
	Role    string `json:"role,omitempty"`    // "admin" for normal JWT
	Purpose string `json:"purpose,omitempty"` // "setup" for 2FA setup
	jwt.RegisteredClaims
}

var jwtKey = []byte("Scholarship") // keep secure

func GenerateSetupToken(email string) (string, error) {
	claims := AdminClaims{
		Email:   email,
		Purpose: "setup",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func ParseSetupToken(tokenStr string) (*AdminClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &AdminClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*AdminClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid setup token")
	}

	// Ensure the purpose is "setup"
	if claims.Purpose != "setup" {
		return nil, errors.New("token purpose is not setup")
	}

	return claims, nil
}
