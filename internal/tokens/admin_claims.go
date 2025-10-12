package tokens

import (
	"github.com/golang-jwt/jwt/v5"
)

// AdminClaims defines the JWT claims structure for admin tokens (including 2FA setup)
type AdminClaims struct {
	ID       int32  `json:"id,omitempty"`
	Username string `json:"username,omitempty"`
	Email    string `json:"email"`
	Role     string `json:"role,omitempty"`    // "admin" for normal JWT
	Purpose  string `json:"purpose,omitempty"` // "setup" for 2FA setup
	jwt.RegisteredClaims
}

func (c *AdminClaims) GetID() int64        { return int64(c.ID) }
func (c *AdminClaims) GetUsername() string { return c.Username }
func (c *AdminClaims) GetEmail() string    { return c.Email }
func (c *AdminClaims) GetRole() string     { return c.Role }
func (c *AdminClaims) GetPurpose() string  { return c.Purpose }
