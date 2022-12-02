package login

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// Token represents a token used for login or booking
type Token struct {

	// Scopes controlling access booking system
	Scopes []string `json:"scopes"`

	jwt.RegisteredClaims
}

// TokenInBody represents a token marshalled into a string
type TokenInBody struct {
	Token string `json:"token"`
}

// String converts a token into a string, returning the string
func (t *Token) String() string {

	pretty, err := json.MarshalIndent(*t, "", "\t")

	if err != nil {
		return fmt.Sprintf("%+v", *t)
	}

	return string(pretty)
}

// New creates a new token (but does not sign it)
func New(audience, subject string, scopes []string, iat, nbf, exp int64) Token {
	return Token{
		Scopes: scopes,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Unix(iat, 0)),
			NotBefore: jwt.NewNumericDate(time.Unix(nbf, 0)),
			ExpiresAt: jwt.NewNumericDate(time.Unix(exp, 0)),
			Audience:  jwt.ClaimStrings{audience},
			Subject:   subject,
		},
	}
}

// Signed signs a token and returns the signed token as a string
func Sign(token Token, secret string) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, token).SignedString([]byte(secret))
}

// HasRequiredClaims returns true if there is at least one scope
func HasRequiredClaims(token Token) bool {
	return len(token.Scopes) != 0
}
