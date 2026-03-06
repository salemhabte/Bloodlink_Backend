package Domain

import "bloodlink/Domain"

const (
	AccessToken  = "access_token"
	RefreshToken = "refresh_token"
)

type IAuthentication interface {
	ParseTokenToClaim(token string) (*Domain.UserClaims, error)
	GenerateToken(claims *Domain.UserClaims, tokenType string) (string, error)
}