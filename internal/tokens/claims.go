package tokens

type ClaimsInterface interface {
	GetEmail() string
	GetRole() string
	GetPurpose() string
}
