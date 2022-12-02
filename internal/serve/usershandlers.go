package serve

import (
	"strconv"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/golang-jwt/jwt/v4"
	lit "github.com/timdrysdale/interval/internal/login"
	"github.com/timdrysdale/interval/internal/serve/models"
	"github.com/timdrysdale/interval/internal/serve/restapi/operations/users"
	"github.com/timdrysdale/interval/internal/store"
)

// getAccessTokenHandler
// minUserNameLength is the minimum length of the usernames that will be accepted
// bookingDuration  is the time in seconds for which to issue the booking token
func getAccessTokenHandler(host, secret string, minUserNameLength int, bookingDuration int64, s *store.Store) func(users.GetAccessTokenParams, interface{}) middleware.Responder {
	return func(params users.GetAccessTokenParams, principal interface{}) middleware.Responder {

		token, ok := principal.(*jwt.Token)
		if !ok {
			c := "401"
			m := "unauthorized: token not JWT"
			return users.NewGetAccessTokenUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		// save checking for key existence individually by checking all at once
		claims, ok := token.Claims.(*lit.Token)

		if !ok {
			c := "401"
			m := "unauthorized: token claims incorrect type"
			return users.NewGetAccessTokenUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		if !lit.HasRequiredClaims(*claims) {
			c := "401"
			m := "unauthorized: token missing required claim"
			return users.NewGetAccessTokenUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		hasLoginUserScope := false
		hasLoginAdminScope := false

		scopes := []string{}

		for _, scope := range claims.Scopes {
			if scope == "login:user" {
				hasLoginUserScope = true
			} else if scope == "login:admin" {
				hasLoginAdminScope = true
			} else {
				scopes = append(scopes, scope)
			}
		}

		if !(hasLoginUserScope || hasLoginAdminScope) {
			c := "401"
			m := "unauthorized: missing login:user or login:admin scope"
			return users.NewGetAccessTokenUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		if len(params.UserName) < minUserNameLength {
			c := "404"
			m := "user name must be " + strconv.Itoa(minUserNameLength) + " or more alphanumeric characters"
			return users.NewGetAccessTokenNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		if hasLoginAdminScope {
			scopes = append(scopes, "booking:admin")
		}
		if hasLoginUserScope {
			scopes = append(scopes, "booking:user")
		}

		bookingClaims := claims
		//keep groups and any other fields added
		bookingClaims.Scopes = scopes //update scopes
		now := jwt.NewNumericDate(s.Now().Add(-1 * time.Second))
		later := jwt.NewNumericDate(s.Now().Add(time.Duration(bookingDuration) * time.Second))
		bookingClaims.IssuedAt = now
		bookingClaims.NotBefore = now
		bookingClaims.ExpiresAt = later
		bookingClaims.Subject = params.UserName

		// sign user token
		// Create a new token object, specifying signing method and the claims
		// you would like it to contain.

		bookingToken := jwt.NewWithClaims(jwt.SigningMethodHS256, bookingClaims)

		// Sign and get the complete encoded token as a string using the secret
		tokenString, err := bookingToken.SignedString(secret)

		if err != nil {
			c := "500"
			m := "could not generate booking token"
			return users.NewGetAccessTokenInternalServerError().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		// If I recall correctly, using float64 here is a limitation of swagger
		exp := float64(bookingClaims.ExpiresAt.Unix())
		iat := float64(bookingClaims.ExpiresAt.Unix())
		nbf := float64(bookingClaims.ExpiresAt.Unix())

		// The login token may have multiple audiences, but the booking token
		// we issue is only valid for us, so we pass our host as the only audience.
		return users.NewGetAccessTokenOK().WithPayload(
			&models.AccessToken{
				Aud:    &host,
				Exp:    &exp,
				Iat:    iat,
				Nbf:    &nbf,
				Scopes: bookingClaims.Scopes,
				Sub:    &bookingClaims.Subject,
				Token:  &tokenString,
			})
	}
}
