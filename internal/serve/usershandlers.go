package serve

import (
	"strconv"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/golang-jwt/jwt/v4"
	"github.com/timdrysdale/interval/internal/config"
	lit "github.com/timdrysdale/interval/internal/login"
	"github.com/timdrysdale/interval/internal/serve/models"
	"github.com/timdrysdale/interval/internal/serve/restapi/operations/users"
)

// getAccessTokenHandler
func getAccessTokenHandler(config config.ServerConfig) func(users.GetAccessTokenParams) middleware.Responder {
	return func(params users.GetAccessTokenParams) middleware.Responder {

		if len(params.UserName) < config.MinUserNameLength {
			c := "404"
			m := "user name must be " + strconv.Itoa(config.MinUserNameLength) + " or more alphanumeric characters"
			return users.NewGetAccessTokenNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		now := jwt.NewNumericDate(config.Store.Now().Add(-1 * time.Second))
		later := jwt.NewNumericDate(config.Store.Now().Add(config.AccessTokenLifetime))

		claims := lit.Token{
			Scopes: []string{"booking:user"},
			RegisteredClaims: jwt.RegisteredClaims{
				IssuedAt:  now,
				NotBefore: now,
				ExpiresAt: later,
				Subject:   params.UserName,
				Audience:  jwt.ClaimStrings{config.Host},
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		// Sign and get the complete encoded token as a string using the secret
		tokenString, err := token.SignedString(config.StoreSecret)

		if err != nil {
			c := "500"
			m := "could not generate booking token"
			return users.NewGetAccessTokenInternalServerError().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		// If I recall correctly, using float64 here is a limitation of swagger
		exp := float64(claims.ExpiresAt.Unix())
		iat := float64(claims.IssuedAt.Unix())
		nbf := float64(claims.NotBefore.Unix())

		// The login token may have multiple audiences, but the booking token
		// we issue is only valid for us, so we pass our host as the only audience.
		return users.NewGetAccessTokenOK().WithPayload(
			&models.AccessToken{
				Aud:    &config.Host,
				Exp:    &exp,
				Iat:    iat,
				Nbf:    &nbf,
				Scopes: claims.Scopes,
				Sub:    &claims.Subject,
				Token:  &tokenString,
			})
	}
}
