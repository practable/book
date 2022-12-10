package serve

import (
	"strconv"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/icza/gog"
	log "github.com/sirupsen/logrus"
	"github.com/timdrysdale/interval/internal/config"
	dt "github.com/timdrysdale/interval/internal/datetime"
	"github.com/timdrysdale/interval/internal/interval"
	lit "github.com/timdrysdale/interval/internal/login"
	"github.com/timdrysdale/interval/internal/serve/models"
	"github.com/timdrysdale/interval/internal/serve/restapi/operations/users"
	"github.com/timdrysdale/interval/internal/store"
)

// dt "github.com/timdrysdale/interval/internal/datetime
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
			m := "could not generate booking token because " + err.Error()
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

// getDescriptiontHandler
func getDescriptionHandler(config config.ServerConfig) func(users.GetDescriptionParams, interface{}) middleware.Responder {
	return func(params users.GetDescriptionParams, principal interface{}) middleware.Responder {

		_, _, err := isAdminOrUser(principal)

		if err != nil {
			c := "401"
			m := err.Error()
			return users.NewGetDescriptionUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		if params.DescriptionName == "" {
			c := "404"
			m := "no description_name in path"
			return users.NewGetDescriptionNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		d, err := config.Store.GetDescription(params.DescriptionName)

		if err != nil {
			c := "500"
			m := err.Error()
			return users.NewGetDescriptionInternalServerError().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		dm := models.Description{
			Name:    &d.Name,
			Type:    &d.Type,
			Short:   d.Short,
			Long:    d.Long,
			Further: d.Further,
			Thumb:   d.Thumb,
			Image:   d.Image,
		}

		return users.NewGetDescriptionOK().WithPayload(&dm)
	}
}

// getPolicytHandler
func getPolicyHandler(config config.ServerConfig) func(users.GetPolicyParams, interface{}) middleware.Responder {
	return func(params users.GetPolicyParams, principal interface{}) middleware.Responder {

		_, _, err := isAdminOrUser(principal)

		if err != nil {
			c := "401"
			m := err.Error()
			return users.NewGetPolicyUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		if params.PolicyName == "" {
			c := "404"
			m := "no policy_name in path"
			return users.NewGetPolicyNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		p, err := config.Store.GetPolicy(params.PolicyName)

		if err != nil {
			c := "500"
			m := err.Error()
			return users.NewGetPolicyInternalServerError().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		dgs := []*models.DisplayGuide{}

		for _, v := range p.DisplayGuides {
			dg := models.DisplayGuide{
				Duration:  gog.Ptr(store.HumaniseDuration(v.Duration)),
				MaxSlots:  gog.Ptr(int64(v.MaxSlots)),
				BookAhead: gog.Ptr(store.HumaniseDuration(v.BookAhead)),
			}
			dgs = append(dgs, &dg)
		}

		pm := models.Policy{
			BookAhead:          store.HumaniseDuration(p.BookAhead),
			Description:        &p.Description,
			DisplayGuides:      dgs,
			EnforceBookAhead:   p.EnforceBookAhead,
			EnforceMaxBookings: p.EnforceMaxBookings,
			EnforceMaxDuration: p.EnforceMaxDuration,
			EnforceMinDuration: p.EnforceMinDuration,
			EnforceMaxUsage:    p.EnforceMaxUsage,
			MaxBookings:        p.MaxBookings,
			MaxDuration:        store.HumaniseDuration(p.MaxDuration),
			MinDuration:        store.HumaniseDuration(p.MinDuration),
			MaxUsage:           store.HumaniseDuration(p.MaxUsage),
			Slots:              p.Slots,
		}

		return users.NewGetPolicyOK().WithPayload(&pm)

	}
}

// getAvailabilityHandler
func getAvailabilityHandler(config config.ServerConfig) func(users.GetAvailabilityParams, interface{}) middleware.Responder {
	return func(params users.GetAvailabilityParams, principal interface{}) middleware.Responder {

		_, _, err := isAdminOrUser(principal)

		if err != nil {
			c := "401"
			m := err.Error()
			return users.NewGetAvailabilityUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		if params.PolicyName == "" {
			c := "404"
			m := "no policy_name in path"
			return users.NewGetAvailabilityNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		if params.SlotName == "" {
			c := "404"
			m := "no slot_name in path"
			return users.NewGetAvailabilityNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		when, err := config.Store.GetAvailability(params.PolicyName, params.SlotName)

		if err != nil {
			c := "500"
			m := "error getting availability from store: " + err.Error()
			return users.NewGetAvailabilityInternalServerError().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		// handle pagination. The offset is equal to the zero-indexed value of the first item of the next page to be
		// returned (20 items are indexed from 0 to 19, so 20 is the first item to be returned in the second page).
		// Note that drift can occur if slots are booked during the sending of availability data, potentially
		// preventing a user from seeing some slots that move earlier in the index and cross a pagination boundary.
		// Users should refresh their results from 0 offset on a regular-ish basis if they wish to avoid this.
		// Or request more results in a single page.

		var limit, offset int

		if params.Limit != nil {
			limit = int(*(params.Limit))
		}
		if params.Offset != nil {
			offset = int(*(params.Offset))
		}

		page := when[offset:]

		if limit > 0 {
			page = page[:limit]
		}

		pm := []*models.Interval{}

		for _, v := range page {
			p := models.Interval{
				Start: strfmt.DateTime(v.Start),
				End:   strfmt.DateTime(v.End),
			}
			pm = append(pm, &p)
		}

		return users.NewGetAvailabilityOK().WithPayload(pm)

	}
}

// makeBookingHandler
func makeBookingHandler(config config.ServerConfig) func(users.MakeBookingParams, interface{}) middleware.Responder {
	return func(params users.MakeBookingParams, principal interface{}) middleware.Responder {

		_, claims, err := isAdminOrUser(principal)

		if err != nil {
			c := "401"
			m := err.Error()
			return users.NewMakeBookingUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		if params.UserName == "" {
			c := "404"
			m := "no user_name in query"
			return users.NewMakeBookingNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		// check username against token
		if claims.Subject != params.UserName {
			c := "401"
			m := "user_name in query does not match subject in token"
			return users.NewMakeBookingUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		if params.PolicyName == "" {
			c := "404"
			m := "no policy_name in path"
			return users.NewMakeBookingNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		if params.SlotName == "" {
			c := "404"
			m := "no slot_name in path"
			return users.NewMakeBookingNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		// Check that the from, to exist and that they parse as future dates
		var emptyDT strfmt.DateTime

		if params.From == emptyDT {
			c := "404"
			m := `no query parameter: from`
			return users.NewMakeBookingNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}
		if params.To == emptyDT {
			c := "404"
			m := `no query parameter: to`
			return users.NewMakeBookingNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		from, err := dt.Parse(params.From.String())

		if err != nil {
			c := "404"
			m := "could not parse ?from=" + params.From.String() + " as RFC3339 datetime"
			return users.NewMakeBookingNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		to, err := dt.Parse(params.To.String())

		if err != nil {
			c := "404"
			m := "could not parse ?to=" + params.To.String() + " as RFC3339 datetime"
			return users.NewMakeBookingNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		when := interval.Interval{
			Start: from,
			End:   to,
		}

		log.Debug(when)

		_, err = config.Store.MakeBooking(params.PolicyName, params.SlotName, params.UserName, when)

		if err != nil {
			c := "404"
			m := "could not make the booking because " + err.Error()
			return users.NewMakeBookingNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		// existing UI ignores any booking info in response to booking request
		// so save sending info we don't need (revisit if UI develops a need for info at this stage)
		return users.NewMakeBookingNoContent()

	}
}
