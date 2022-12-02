package login

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
)

func TestNewValidate(t *testing.T) {

	audience := "https://login.something.com"
	subject := "someuser"
	scopes := []string{"login", "user"}
	nbf := int64(1609721410)
	iat := nbf
	exp := nbf + 10

	token := New(audience, subject, scopes, iat, nbf, exp)
	assert.Equal(t, audience, token.Audience[0])
	assert.Equal(t, "someuser", token.Subject)
	assert.Equal(t, scopes, token.Scopes)
	assert.Equal(t, *jwt.NewNumericDate(time.Unix(iat, 0)), *token.IssuedAt)
	assert.Equal(t, *jwt.NewNumericDate(time.Unix(nbf, 0)), *token.NotBefore)
	assert.Equal(t, *jwt.NewNumericDate(time.Unix(exp, 0)), *token.ExpiresAt)
	assert.True(t, HasRequiredClaims(token))

}
