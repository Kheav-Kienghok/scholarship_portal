package tokens

import "github.com/golang-jwt/jwt/v5"

// AdminClaims defines the JWT claims structure for admin tokens (including 2FA setup)
type AdminClaims struct {
	ID       int32  `json:"id,omitempty"`
	Fullname string `json:"fullname,omitempty"` // optional, can be empty
	Email    string `json:"email"`
	Role     string `json:"role,omitempty"`    // "admin" for normal JWT
	Purpose  string `json:"purpose,omitempty"` // "setup" for 2FA setup
	jwt.RegisteredClaims
}

func (c *AdminClaims) GetEmail() string    { return c.Email }
func (c *AdminClaims) GetRole() string     { return c.Role }
func (c *AdminClaims) GetPurpose() string  { return c.Purpose }
func (c *AdminClaims) GetFullname() string { return c.Fullname } // optional, can be empty
