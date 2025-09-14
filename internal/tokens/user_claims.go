package tokens

import "github.com/golang-jwt/jwt/v5"

type UserClaims struct {
	ID       int32  `json:"id"`
	Fullname string `json:"fullname,omitempty"` // optional, can be empty
	Email    string `json:"email"`
	Role     string `json:"role"`    // "student"
	Purpose  string `json:"purpose"` // e.g., "setup" for 2FA
	jwt.RegisteredClaims
}

func (c *UserClaims) GetID() int64        { return int64(c.ID) }
func (c *UserClaims) GetEmail() string    { return c.Email }
func (c *UserClaims) GetRole() string     { return c.Role }
func (c *UserClaims) GetPurpose() string  { return c.Purpose }
func (c *UserClaims) GetFullname() string { return c.Fullname } // optional, can be empty
