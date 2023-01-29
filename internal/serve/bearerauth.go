package serve

import (
	"errors"
	"fmt"

	"github.com/go-openapi/runtime/security"
	"github.com/golang-jwt/jwt/v4"
	lit "github.com/practable/book/internal/login"
	log "github.com/sirupsen/logrus"
)

func claimsCheck(principal interface{}) (*lit.Token, error) {

	token, ok := principal.(*jwt.Token)
	if !ok {
		return nil, errors.New("token Not JWT")
	}

	// save checking for key existence individually by checking all at once
	claims, ok := token.Claims.(*lit.Token)

	if !ok {
		return nil, errors.New("token claims incorrect type")
	}

	if !lit.HasRequiredClaims(*claims) {
		return nil, errors.New("token missing required claims")
	}

	return claims, nil
}

// isAdmin returns nil if token has booking:admin scope, otherwise error
func isAdmin(principal interface{}) (*lit.Token, error) {

	claims, err := claimsCheck(principal)

	if err != nil {
		log.WithFields(log.Fields{"token": principal, "error": err.Error()}).Info("token failed claimsCheck")
		return nil, err
	}

	hasAdminScope := false

	for _, scope := range claims.Scopes {
		if scope == "booking:admin" {
			hasAdminScope = true
		}
	}

	if !hasAdminScope {
		return nil, errors.New("Missing booking:admin Scope")
	}

	return claims, nil
}

// isUser returns nil if token has booking:user scope, otherwise error
func isUser(principal interface{}) (*lit.Token, error) {

	claims, err := claimsCheck(principal)

	if err != nil {
		log.WithFields(log.Fields{"token": principal, "error": err.Error()}).Info("token failed claimsCheck")
		return nil, err
	}

	hasUserScope := false

	for _, scope := range claims.Scopes {
		if scope == "booking:user" {
			hasUserScope = true
		}
	}

	if !hasUserScope {
		return nil, errors.New("Missing booking:user Scope")
	}

	return claims, nil
}

func isAdminOrUser(principal interface{}) (bool, *lit.Token, error) {

	claims, err := claimsCheck(principal)

	if err != nil {
		log.WithFields(log.Fields{"token": principal, "error": err.Error()}).Info("token failed claimsCheck")
		return false, nil, err
	}

	hasAdminScope := false
	hasUserScope := false

	for _, scope := range claims.Scopes {
		if scope == "booking:admin" {
			hasAdminScope = true
		}
		if scope == "booking:user" {
			hasUserScope = true
		}
	}

	if !hasAdminScope && !hasUserScope {
		return false, nil, errors.New("Missing booking:admin or booking:user Scope")
	}

	return hasAdminScope, claims, nil
}

// ValidateHeader checks the bearer token.
// wrap the secret so we can get it at runtime without using global
func validateHeader(secret []byte, host string) security.TokenAuthentication {

	return func(bearerToken string) (interface{}, error) {
		// For apiKey security syntax see https://swagger.io/docs/specification/2-0/authentication/

		claims := &lit.Token{}

		token, err := jwt.ParseWithClaims(bearerToken, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				log.WithFields(log.Fields{"alg": token.Header["alg"], "token": token}).Info("wrong signing method")
				return nil, fmt.Errorf("unexpected signing method was %v", token.Header["alg"])
			}
			return secret, nil
		})

		if err != nil {
			log.WithFields(log.Fields{"error": err.Error(), "token": token}).Info("error parsing claims")
			return nil, errors.New("error parsing claims was " + err.Error())
		}

		if token == nil {
			log.Info("nil token")
			return nil, fmt.Errorf("nil token")
		}

		if !token.Valid { //checks iat, nbf, exp
			log.WithFields(log.Fields{"token": token}).Info("token invalid")
			return nil, fmt.Errorf("token invalid")
		}

		if cc, ok := token.Claims.(*lit.Token); ok {

			if !cc.RegisteredClaims.VerifyAudience(host, true) {
				log.WithFields(log.Fields{"aud": cc.RegisteredClaims.Audience, "host": host}).Info("token aud does not match this host")
				return nil, fmt.Errorf("aud %s does not match this host %s", cc.RegisteredClaims.Audience, host)
			}

		} else {
			log.WithFields(log.Fields{"token": bearerToken, "host": host}).Info("error parsing token")
			return nil, err
		}

		return token, nil
	}
}
