package tokens

type ClaimsInterface interface {
	GetID() int64
	GetEmail() string
	GetRole() string
	GetPurpose() string
}
