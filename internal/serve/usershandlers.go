package serve

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/icza/gog"
	"github.com/practable/book/internal/config"
	dt "github.com/practable/book/internal/datetime"
	"github.com/practable/book/internal/interval"
	lit "github.com/practable/book/internal/login"
	"github.com/practable/book/internal/serve/models"
	"github.com/practable/book/internal/serve/restapi/operations/users"
	"github.com/practable/book/internal/store"
	log "github.com/sirupsen/logrus"
)

func calendarError(code, message string) *models.Error {
	return &models.Error{Code: &code, Message: &message}
}

func calendarSelectorFromModel(value *models.CalendarSelector) store.CalendarSelector {
	if value == nil || value.Policy == nil {
		return store.CalendarSelector{}
	}
	return store.CalendarSelector{Policy: *value.Policy, Resource: value.Resource, Properties: value.Properties}
}

func calendarSelectorToModel(value store.CalendarSelector) *models.CalendarSelector {
	return &models.CalendarSelector{Policy: gog.Ptr(value.Policy), Resource: value.Resource, Properties: value.Properties}
}

func authorizeCalendarUser(principal interface{}, requestedUser string) (bool, *lit.Token, error) {
	isAdmin, claims, err := isAdminOrUser(principal)
	if err != nil {
		return false, nil, err
	}
	if !isAdmin && claims.Subject != requestedUser {
		return false, claims, errors.New("user_name does not match subject in token")
	}
	return isAdmin, claims, nil
}

func calendarPolicyAllowed(s *store.Store, user, policy string) bool {
	groups, err := s.GetGroupsFor(user)
	if err != nil {
		return false
	}
	for _, groupName := range groups {
		group, err := s.GetGroup(groupName)
		if err != nil {
			continue
		}
		for _, candidate := range group.Policies {
			if candidate == policy {
				return true
			}
		}
	}
	return false
}

type Permission struct {
	BookingID string   `json:"booking_id"`
	Topic     string   `json:"topic"`
	Prefix    string   `json:"prefix"`
	Scopes    []string `json:"scopes"`
	jwt.RegisteredClaims
}

// convertStoreStatusUserToModel converts from internal to API type
func convertStoreStatusUserToModel(s store.StoreStatusUser) (models.StoreStatusUser, error) {
	var m models.StoreStatusUser

	y, err := json.Marshal(s)

	if err != nil {
		return m, err
	}

	err = json.Unmarshal(y, &m)

	return m, err

}

// dt "github.com/practable/book/internal/datetime
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

// getGroupHandler
func getGroupHandler(config config.ServerConfig) func(users.GetGroupParams, interface{}) middleware.Responder {
	return func(params users.GetGroupParams, principal interface{}) middleware.Responder {

		_, _, err := isAdminOrUser(principal)

		if err != nil {
			c := "401"
			m := err.Error()
			return users.NewGetGroupUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		if params.GroupName == "" {
			c := "404"
			m := "no group_name in path"
			return users.NewGetGroupNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		g, err := config.Store.GetGroup(params.GroupName)

		if err != nil {
			c := "500"
			m := err.Error()
			return users.NewGetGroupInternalServerError().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		d := g.Description

		gd := models.GroupDescribedWithPolicies{
			Description: gog.Ptr(models.Description{
				Name:    &d.Name,
				Type:    &d.Type,
				Short:   d.Short,
				Long:    d.Long,
				Further: d.Further,
				Thumb:   d.Thumb,
				Image:   d.Image,
			}),
		}

		pms := models.PoliciesDescribed{}

		for _, pn := range g.Policies {

			p, err := config.Store.GetPolicy(pn)

			if err != nil {
				c := "500"
				m := err.Error()
				return users.NewGetGroupInternalServerError().WithPayload(&models.Error{Code: &c, Message: &m})
			}

			dgm := make(map[string]models.DisplayGuide)

			for k, v := range p.DisplayGuidesMap {

				dg := v // avoid pointers to last element problem

				dgm[k] = models.DisplayGuide{
					Duration:  gog.Ptr(store.HumaniseDuration(dg.Duration)),
					MaxSlots:  gog.Ptr(int64(dg.MaxSlots)),
					BookAhead: gog.Ptr(store.HumaniseDuration(dg.BookAhead)),
					Label:     gog.Ptr(dg.Label),
				}
			}

			descr, err := config.Store.GetDescription(p.Description)
			if err != nil {
				c := "500"
				m := err.Error()
				return users.NewGetGroupInternalServerError().WithPayload(&models.Error{Code: &c, Message: &m})
			}

			slm := make(map[string]models.SlotDescribed)

			for _, sln := range p.Slots {
				sl, err := config.Store.GetSlot(sln)
				if err != nil {
					c := "500"
					m := err.Error()
					return users.NewGetGroupInternalServerError().WithPayload(&models.Error{Code: &c, Message: &m})
				}
				d, err := config.Store.GetDescription(sl.Description)
				if err != nil {
					c := "500"
					m := err.Error()
					return users.NewGetGroupInternalServerError().WithPayload(&models.Error{Code: &c, Message: &m})
				}
				sld := models.SlotDescribed{
					Description: gog.Ptr(models.Description{
						Name:    &d.Name,
						Type:    &d.Type,
						Short:   d.Short,
						Long:    d.Long,
						Further: d.Further,
						Thumb:   d.Thumb,
						Image:   d.Image,
					}),
					Policy: gog.Ptr(sl.Policy),
				}
				slm[sln] = sld
			}

			pm := models.PolicyDescribed{
				AllowStartInPastWithin: store.HumaniseDuration(p.AllowStartInPastWithin),
				BookAhead:              store.HumaniseDuration(p.BookAhead),
				Description: gog.Ptr(models.Description{
					Name:    &descr.Name,
					Type:    &descr.Type,
					Short:   descr.Short,
					Long:    descr.Long,
					Further: descr.Further,
					Thumb:   descr.Thumb,
					Image:   descr.Image,
				}),
				DisplayGuides:           dgm,
				EnforceAllowStartInPast: p.EnforceAllowStartInPast,
				EnforceBookAhead:        p.EnforceBookAhead,
				EnforceMaxBookings:      p.EnforceMaxBookings,
				EnforceMaxDuration:      p.EnforceMaxDuration,
				EnforceMinDuration:      p.EnforceMinDuration,
				EnforceMaxUsage:         p.EnforceMaxUsage,
				EnforceNextAvailable:    p.EnforceNextAvailable,
				EnforceStartsWithin:     p.EnforceStartsWithin,
				EnforceUnlimitedUsers:   p.EnforceUnlimitedUsers,
				MaxBookings:             p.MaxBookings,
				MaxDuration:             store.HumaniseDuration(p.MaxDuration),
				MinDuration:             store.HumaniseDuration(p.MinDuration),
				MaxUsage:                store.HumaniseDuration(p.MaxUsage),
				NextAvailable:           store.HumaniseDuration(p.NextAvailable),
				Slots:                   slm,
				StartsWithin:            store.HumaniseDuration(p.StartsWithin),
			}

			pms[pn] = pm
		}

		gd.Policies = pms

		return users.NewGetGroupOK().WithPayload(&gd)

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

		dgm := make(map[string]models.DisplayGuide)

		for k, v := range p.DisplayGuidesMap {

			dg := v // avoid pointers to last element problem

			dgm[k] = models.DisplayGuide{
				Duration:  gog.Ptr(store.HumaniseDuration(dg.Duration)),
				MaxSlots:  gog.Ptr(int64(dg.MaxSlots)),
				BookAhead: gog.Ptr(store.HumaniseDuration(dg.BookAhead)),
				Label:     gog.Ptr(dg.Label),
			}
		}

		descr, err := config.Store.GetDescription(p.Description)
		if err != nil {
			c := "500"
			m := err.Error()
			return users.NewGetPolicyInternalServerError().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		slm := make(map[string]models.SlotDescribed)

		for _, sln := range p.Slots {
			sl, err := config.Store.GetSlot(sln)
			if err != nil {
				c := "500"
				m := err.Error()
				return users.NewGetGroupInternalServerError().WithPayload(&models.Error{Code: &c, Message: &m})
			}
			d, err := config.Store.GetDescription(sl.Description)
			if err != nil {
				c := "500"
				m := err.Error()
				return users.NewGetGroupInternalServerError().WithPayload(&models.Error{Code: &c, Message: &m})
			}
			sld := models.SlotDescribed{
				Description: gog.Ptr(models.Description{
					Name:    &d.Name,
					Type:    &d.Type,
					Short:   d.Short,
					Long:    d.Long,
					Further: d.Further,
					Thumb:   d.Thumb,
					Image:   d.Image,
				}),
				Policy: gog.Ptr(sl.Policy),
			}
			slm[sln] = sld
		}

		pm := models.PolicyDescribed{
			AllowStartInPastWithin: store.HumaniseDuration(p.AllowStartInPastWithin),
			BookAhead:              store.HumaniseDuration(p.BookAhead),
			Description: gog.Ptr(models.Description{
				Name:    &descr.Name,
				Type:    &descr.Type,
				Short:   descr.Short,
				Long:    descr.Long,
				Further: descr.Further,
				Thumb:   descr.Thumb,
				Image:   descr.Image,
			}),
			DisplayGuides:           dgm,
			EnforceAllowStartInPast: p.EnforceAllowStartInPast,
			EnforceBookAhead:        p.EnforceBookAhead,
			EnforceMaxBookings:      p.EnforceMaxBookings,
			EnforceMaxDuration:      p.EnforceMaxDuration,
			EnforceMinDuration:      p.EnforceMinDuration,
			EnforceMaxUsage:         p.EnforceMaxUsage,
			EnforceNextAvailable:    p.EnforceNextAvailable,
			EnforceStartsWithin:     p.EnforceStartsWithin,
			EnforceUnlimitedUsers:   p.EnforceUnlimitedUsers,
			MaxBookings:             p.MaxBookings,
			MaxDuration:             store.HumaniseDuration(p.MaxDuration),
			MinDuration:             store.HumaniseDuration(p.MinDuration),
			MaxUsage:                store.HumaniseDuration(p.MaxUsage),
			NextAvailable:           store.HumaniseDuration(p.NextAvailable),
			Slots:                   slm,
			StartsWithin:            store.HumaniseDuration(p.StartsWithin),
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

		if params.SlotName == "" {
			c := "404"
			m := "no slot_name in path"
			return users.NewGetAvailabilityNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		when, err := config.Store.GetAvailability(params.SlotName)

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

func uniqueNameHandler(config config.ServerConfig) func(users.UniqueNameParams) middleware.Responder {
	return func(params users.UniqueNameParams) middleware.Responder {

		u := models.UserName{
			UserName: config.Store.GenerateUniqueUser(),
		}
		return users.NewUniqueNameOK().WithPayload(&u)
	}
}

// makeBookingHandler
func makeBookingHandler(config config.ServerConfig) func(users.MakeBookingParams, interface{}) middleware.Responder {
	return func(params users.MakeBookingParams, principal interface{}) middleware.Responder {

		isAdmin, claims, err := isAdminOrUser(principal)

		if err != nil {
			c := "401"
			m := err.Error()
			log.WithFields(log.Fields{"token": principal, "error": err.Error()}).Debug("make booking unauthorized")
			return users.NewMakeBookingUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		if params.UserName == "" {
			c := "404"
			m := "no user_name in query"
			log.Debug("make booking no user_name in query")
			return users.NewMakeBookingNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		// check username against token (admins can book on behalf of users, so ignore if
		if (!isAdmin) && (claims.Subject != params.UserName) {
			c := "401"
			m := "user_name in query does not match subject in token"
			return users.NewMakeBookingUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
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

		_, err = config.Store.MakeBooking(params.SlotName, params.UserName, when)

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

func getCalendarCatalogueHandler(config config.ServerConfig) func(users.GetCalendarCatalogueParams, interface{}) middleware.Responder {
	return func(params users.GetCalendarCatalogueParams, principal interface{}) middleware.Responder {
		isAdmin, claims, err := isAdminOrUser(principal)
		if err != nil {
			return users.NewGetCalendarCatalogueUnauthorized().WithPayload(calendarError("401", err.Error()))
		}
		if !isAdmin {
			groups, err := config.Store.GetGroupsFor(claims.Subject)
			if err != nil {
				return users.NewGetCalendarCatalogueNotFound().WithPayload(calendarError("404", err.Error()))
			}
			allowed := false
			for _, group := range groups {
				allowed = allowed || group == params.GroupName
			}
			if !allowed {
				return users.NewGetCalendarCatalogueUnauthorized().WithPayload(calendarError("401", "group is not assigned to this user"))
			}
		}
		items, err := config.Store.GetCalendarCatalogue(params.GroupName)
		if err != nil {
			return users.NewGetCalendarCatalogueNotFound().WithPayload(calendarError("404", err.Error()))
		}
		payload := make(models.CalendarCatalogue, 0, len(items))
		for _, item := range items {
			description := item.Description
			model := &models.CalendarCatalogueItem{
				Policy: gog.Ptr(item.Policy), BookingIncrement: gog.Ptr(store.HumaniseDuration(item.BookingIncrement)),
				Description: &models.Description{Name: gog.Ptr(description.Name), Type: gog.Ptr(description.Type), Short: description.Short, Long: description.Long, Further: description.Further, Thumb: description.Thumb, Image: description.Image},
				Resources:   make([]*models.CalendarResource, 0, len(item.Resources)),
			}
			if item.RecommendedDuration > 0 {
				model.RecommendedDuration = store.HumaniseDuration(item.RecommendedDuration)
			}
			for _, resource := range item.Resources {
				model.Resources = append(model.Resources, calendarResourceToModel(resource))
			}
			payload = append(payload, model)
		}
		return users.NewGetCalendarCatalogueOK().WithPayload(payload)
	}
}

func queryCalendarAvailabilityHandler(config config.ServerConfig) func(users.QueryCalendarAvailabilityParams, interface{}) middleware.Responder {
	return func(params users.QueryCalendarAvailabilityParams, principal interface{}) middleware.Responder {
		request := params.Request
		if request == nil || request.UserName == nil || request.Selector == nil || request.From == nil || request.To == nil || request.Duration == nil || request.Resolution == nil {
			return users.NewQueryCalendarAvailabilityBadRequest().WithPayload(calendarError("400", "incomplete calendar availability request"))
		}
		isAdmin, _, err := authorizeCalendarUser(principal, *request.UserName)
		if err != nil {
			return users.NewQueryCalendarAvailabilityUnauthorized().WithPayload(calendarError("401", err.Error()))
		}
		selector := calendarSelectorFromModel(request.Selector)
		if !isAdmin && !calendarPolicyAllowed(config.Store, *request.UserName, selector.Policy) {
			return users.NewQueryCalendarAvailabilityUnauthorized().WithPayload(calendarError("401", "policy is not assigned to this user"))
		}
		duration, err := time.ParseDuration(*request.Duration)
		if err != nil {
			return users.NewQueryCalendarAvailabilityBadRequest().WithPayload(calendarError("400", "invalid duration"))
		}
		resolution, err := time.ParseDuration(*request.Resolution)
		if err != nil {
			return users.NewQueryCalendarAvailabilityBadRequest().WithPayload(calendarError("400", "invalid resolution"))
		}
		bands, err := config.Store.GetCalendarAvailability(selector, time.Time(*request.From), time.Time(*request.To), duration, resolution)
		if err != nil {
			return users.NewQueryCalendarAvailabilityBadRequest().WithPayload(calendarError("400", err.Error()))
		}
		payload := make(models.CalendarAvailability, 0, len(bands))
		for _, band := range bands {
			start, end, bookable, count, mode := strfmt.DateTime(band.Start), strfmt.DateTime(band.End), band.Bookable, band.MatchingResources, band.OperatingMode
			payload = append(payload, &models.CalendarAvailabilityBand{Start: &start, End: &end, Bookable: &bookable, MatchingResources: &count, DegradedResources: band.DegradedResources, OperatingMode: &mode, Reason: band.Reason})
		}
		return users.NewQueryCalendarAvailabilityOK().WithPayload(payload)
	}
}

func calendarInterval(value *models.Interval) (interval.Interval, error) {
	if value == nil || time.Time(value.Start).IsZero() || time.Time(value.End).IsZero() {
		return interval.Interval{}, errors.New("booking interval requires start and end")
	}
	return interval.Interval{Start: time.Time(value.Start), End: time.Time(value.End)}, nil
}

func previewCalendarBookingHandler(config config.ServerConfig) func(users.PreviewCalendarBookingParams, interface{}) middleware.Responder {
	return func(params users.PreviewCalendarBookingParams, principal interface{}) middleware.Responder {
		request := params.Request
		if request == nil || request.UserName == nil || request.Selector == nil {
			return users.NewPreviewCalendarBookingBadRequest().WithPayload(calendarError("400", "incomplete calendar booking request"))
		}
		isAdmin, _, err := authorizeCalendarUser(principal, *request.UserName)
		if err != nil {
			return users.NewPreviewCalendarBookingUnauthorized().WithPayload(calendarError("401", err.Error()))
		}
		selector := calendarSelectorFromModel(request.Selector)
		if !isAdmin && !calendarPolicyAllowed(config.Store, *request.UserName, selector.Policy) {
			return users.NewPreviewCalendarBookingUnauthorized().WithPayload(calendarError("401", "policy is not assigned to this user"))
		}
		when, err := calendarInterval(request.When)
		if err != nil {
			return users.NewPreviewCalendarBookingBadRequest().WithPayload(calendarError("400", err.Error()))
		}
		preview, err := config.Store.PreviewCalendarBooking(*request.UserName, selector, when)
		if err != nil {
			return users.NewPreviewCalendarBookingNotFound().WithPayload(calendarError("404", err.Error()))
		}
		bookable, version, usageAfter := preview.Bookable, preview.ManifestVersion, store.HumaniseDuration(preview.UsageAfter)
		payload := &models.CalendarBookingPreview{
			When: request.When, Selector: calendarSelectorToModel(preview.Selector), MatchingResources: preview.MatchingResources,
			Bookable: &bookable, Reasons: preview.Reasons, UsageAfter: &usageAfter, ManifestVersion: &version,
			OperationalEffects: preview.OperationalEffects,
		}
		for _, resource := range preview.DegradedResources {
			payload.DegradedResources = append(payload.DegradedResources, calendarResourceToModel(resource))
		}
		if preview.UsageLimit > 0 {
			payload.UsageLimit = store.HumaniseDuration(preview.UsageLimit)
		}
		return users.NewPreviewCalendarBookingOK().WithPayload(payload)
	}
}

func calendarResourceToModel(resource store.CalendarResource) *models.CalendarResource {
	return &models.CalendarResource{Name: gog.Ptr(resource.Name), Class: resource.Class, Properties: resource.Properties,
		ActivationStreams: resource.ActivationStreams, Degraded: resource.Degraded, DegradedReason: resource.DegradedReason, UnavailableStreams: resource.UnavailableStreams}
}

func makeCalendarBookingHandler(config config.ServerConfig) func(users.MakeCalendarBookingParams, interface{}) middleware.Responder {
	return func(params users.MakeCalendarBookingParams, principal interface{}) middleware.Responder {
		request := params.Request
		if request == nil || request.UserName == nil || request.Selector == nil {
			return users.NewMakeCalendarBookingBadRequest().WithPayload(calendarError("400", "incomplete calendar booking request"))
		}
		isAdmin, _, err := authorizeCalendarUser(principal, *request.UserName)
		if err != nil {
			return users.NewMakeCalendarBookingUnauthorized().WithPayload(calendarError("401", err.Error()))
		}
		selector := calendarSelectorFromModel(request.Selector)
		if !isAdmin && !calendarPolicyAllowed(config.Store, *request.UserName, selector.Policy) {
			return users.NewMakeCalendarBookingUnauthorized().WithPayload(calendarError("401", "policy is not assigned to this user"))
		}
		when, err := calendarInterval(request.When)
		if err != nil {
			return users.NewMakeCalendarBookingBadRequest().WithPayload(calendarError("400", err.Error()))
		}
		booking, err := config.Store.MakeCalendarBooking(*request.UserName, selector, when, params.IdempotencyKey)
		if err != nil {
			if errors.Is(err, store.ErrBookingConflict) || errors.Is(err, store.ErrBookingIDConflict) || errors.Is(err, store.ErrMaxBookings) || errors.Is(err, store.ErrMaxUsage) {
				return users.NewMakeCalendarBookingConflict().WithPayload(calendarError("409", err.Error()))
			}
			return users.NewMakeCalendarBookingNotFound().WithPayload(calendarError("404", err.Error()))
		}
		return users.NewMakeCalendarBookingOK().WithPayload(bookingToModel(booking))
	}
}

func rescheduleCalendarBookingHandler(config config.ServerConfig) func(users.RescheduleCalendarBookingParams, interface{}) middleware.Responder {
	return func(params users.RescheduleCalendarBookingParams, principal interface{}) middleware.Responder {
		request := params.Request
		if request == nil || request.UserName == nil || request.Selector == nil || request.Revision == nil {
			return users.NewRescheduleCalendarBookingBadRequest().WithPayload(calendarError("400", "incomplete calendar booking edit"))
		}
		isAdmin, _, err := authorizeCalendarUser(principal, *request.UserName)
		if err != nil {
			return users.NewRescheduleCalendarBookingUnauthorized().WithPayload(calendarError("401", err.Error()))
		}
		selector := calendarSelectorFromModel(request.Selector)
		if !isAdmin && !calendarPolicyAllowed(config.Store, *request.UserName, selector.Policy) {
			return users.NewRescheduleCalendarBookingUnauthorized().WithPayload(calendarError("401", "policy is not assigned to this user"))
		}
		when, err := calendarInterval(request.When)
		if err != nil {
			return users.NewRescheduleCalendarBookingBadRequest().WithPayload(calendarError("400", err.Error()))
		}
		updated, err := config.Store.RescheduleCalendarBooking(*request.UserName, params.BookingName, *request.Revision, selector, when)
		if err != nil {
			if errors.Is(err, store.ErrPersistentNotFound) {
				return users.NewRescheduleCalendarBookingNotFound().WithPayload(calendarError("404", "booking not found"))
			}
			return users.NewRescheduleCalendarBookingConflict().WithPayload(calendarError("409", err.Error()))
		}
		return users.NewRescheduleCalendarBookingOK().WithPayload(editableBookingToModel(updated))
	}
}

// getStoreStatusUserHandler
func getStoreStatusUserHandler(config config.ServerConfig) func(users.GetStoreStatusUserParams, interface{}) middleware.Responder {
	return func(params users.GetStoreStatusUserParams, principal interface{}) middleware.Responder {

		_, _, err := isAdminOrUser(principal)

		if err != nil {
			c := "401"
			m := err.Error()
			return users.NewGetStoreStatusUserUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		s, err := convertStoreStatusUserToModel(config.Store.GetStoreStatusUser())

		if err != nil {
			log.Error("could not convert StoreStatusAdmin to model format")
		}

		return users.NewGetStoreStatusUserOK().WithPayload(&s)
	}
}

// getBookingsForHandler
func getBookingsForUserHandler(config config.ServerConfig) func(users.GetBookingsForUserParams, interface{}) middleware.Responder {
	return func(params users.GetBookingsForUserParams, principal interface{}) middleware.Responder {

		isAdmin, claims, err := isAdminOrUser(principal)

		if err != nil {
			c := "401"
			m := err.Error()
			return users.NewGetBookingsForUserUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		if params.UserName == "" {
			c := "404"
			m := "no user_name in query"
			return users.NewGetBookingsForUserNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		// check username against token, unless admin (admin can check on behalf of users)
		if (!isAdmin) && (claims.Subject != params.UserName) {
			c := "401"
			m := "user_name in path " + params.UserName + " does not match subject " + claims.Subject + " in token"
			return users.NewGetBookingsForUserUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		bs, err := config.Store.GetBookingsFor(params.UserName)

		if err != nil {
			c := "404"
			m := "error retrieving bookings for user " + params.UserName + ": " + err.Error()
			return users.NewGetBookingsForUserNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		var bm models.Bookings

		for _, v := range bs {

			b := models.Booking{
				Cancelled:   v.Cancelled,
				Name:        gog.Ptr(v.Name),
				Policy:      gog.Ptr(v.Policy),
				Slot:        gog.Ptr(v.Slot),
				Started:     v.Started,
				Unfulfilled: v.Unfulfilled,
				User:        gog.Ptr(v.User),
				When: gog.Ptr(models.Interval{
					Start: strfmt.DateTime(v.When.Start),
					End:   strfmt.DateTime(v.When.End),
				}),
			}
			bm = append(bm, &b)
		}

		return users.NewGetBookingsForUserOK().WithPayload(bm)
	}
}

// cancelBookingHandler
func cancelBookingHandler(config config.ServerConfig) func(users.CancelBookingParams, interface{}) middleware.Responder {
	return func(params users.CancelBookingParams, principal interface{}) middleware.Responder {

		isAdmin, claims, err := isActivityCaller(principal)

		//fmt.Printf("CancelBooking %s by %s isAdmin?%v isLocked?%v\n", params.BookingName, params.UserName, isAdmin, config.Store.Locked)
		if err != nil {
			c := "401"
			m := err.Error()
			return users.NewCancelBookingUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		if params.UserName == "" {
			c := "404"
			m := "no user_name in path"
			return users.NewCancelBookingNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		// check username against token, if not admin (admin can cancel on behalf of users)
		if (!isAdmin) && (claims.Subject != params.UserName) {
			c := "401"
			m := "user_name in query does not match subject in token"
			return users.NewCancelBookingUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		if params.BookingName == "" {
			c := "404"
			m := "no booking_name in path"
			return users.NewCancelBookingNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		b, err := config.Store.GetBooking(params.BookingName)
		if err != nil {
			c := "404"
			m := "not found"
			return users.NewCancelBookingNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}
		if !cancelScopeAllowsBooking(isAdmin, claims, b) {
			c := "401"
			m := "maintenance scope may cancel only an owned maintenance booking"
			return users.NewCancelBookingUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		cancelledBy := claims.Subject
		if isAdmin && claims.Subject == "" {
			cancelledBy = "admin"
		}
		err = config.Store.CancelBooking(b, cancelledBy)

		if err != nil {
			c := "500"
			m := err.Error()
			return users.NewCancelBookingInternalServerError().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		// Use NotFound to indicate successful deletion. Repeat calls will return NotFound
		log.WithFields(log.Fields{"user": params.UserName, "booking": params.BookingName}).Info("booking cancelled successfully")
		c := "404"
		m := "cancelled"
		return users.NewCancelBookingNotFound().WithPayload(&models.Error{Code: &c, Message: &m})

	}
}

func cancelScopeAllowsBooking(isAdmin bool, claims *lit.Token, booking store.Booking) bool {
	if isAdmin {
		return true
	}
	for _, scope := range claims.Scopes {
		if scope == "booking:user" {
			return true
		}
	}
	return booking.Maintenance
}

// getActivityHandler
func getActivityHandler(config config.ServerConfig) func(users.GetActivityParams, interface{}) middleware.Responder {
	return func(params users.GetActivityParams, principal interface{}) middleware.Responder {

		isAdmin, claims, err := isActivityCaller(principal)

		if err != nil {
			c := "404"
			m := err.Error()
			return users.NewGetActivityNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		if params.UserName == "" {
			c := "404"
			m := "no user_name in path"
			return users.NewGetActivityNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		// check username against token, if not admin (admin can cancel on behalf of users)
		if (!isAdmin) && (claims.Subject != params.UserName) {
			c := "404"
			m := "user_name in query does not match subject in token"
			return users.NewGetActivityNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		if params.BookingName == "" {
			c := "404"
			m := "no booking_name in path"
			return users.NewGetActivityNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		b, err := config.Store.GetBooking(params.BookingName)

		if err != nil {
			c := "404"
			m := "booking not found " + err.Error()
			return users.NewGetActivityNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		a, err := config.Store.GetActivity(b)

		if err != nil {
			c := "404"
			m := err.Error()
			return users.NewGetActivityNotFound().WithPayload(&models.Error{Code: &c, Message: &m})

		}
		// convert stream format
		streams := []*models.ActivityStream{}

		/* Stream token format:
		   {
		     "topic": "pend13-data",
		     "prefix": "session",
		     "scopes": [
		       "read",
		       "write"
		     ],
		     "aud": [
		       "https://relay-access.practable.io"
		     ],
		     "exp": 1670703344,
		     "nbf": 1670703044,
		     "iat": 1670703044
		   }*/
		for k, v := range a.Streams {

			st := v //avoid all pointers pointing to last in map
			now := jwt.NewNumericDate(config.Store.Now().Add(-1 * time.Second))
			later := jwt.NewNumericDate(b.When.End)

			permission := Permission{
				BookingID: b.Name,
				Topic:     st.Topic,
				Prefix:    st.ConnectionType,
				Scopes:    st.Scopes,
				RegisteredClaims: jwt.RegisteredClaims{
					IssuedAt:  now,
					NotBefore: now,
					ExpiresAt: later,
					Subject:   params.UserName, //adding for future usage
					Audience:  jwt.ClaimStrings{st.URL},
				},
			}
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, permission)
			// Sign and get the complete encoded token as a string using the relay secret
			stoken, err := token.SignedString(config.RelaySecret)

			if err != nil {
				c := "500"
				m := "error making token for stream " + k + " : " + err.Error()
				return users.NewGetActivityInternalServerError().WithPayload(&models.Error{Code: &c, Message: &m})
			}

			stm := gog.Ptr(models.ActivityStream{
				Audience:       gog.Ptr(st.URL),
				ConnectionType: gog.Ptr(st.ConnectionType),
				For:            gog.Ptr(st.For),
				Scopes:         v.Scopes,
				Topic:          gog.Ptr(st.Topic),
				URL:            gog.Ptr(st.URL + "/" + st.ConnectionType + "/" + st.Topic), //this is the URL the UI attempts to connect to, so must be complete
				Token:          stoken,
			})
			streams = append(streams, stm)
		}

		// convert UIDescribed format
		uids := []*models.UIDescribed{}

		for _, v := range a.UIs {
			u := v //avoid all pointers pointing to last in map
			uid := gog.Ptr(models.UIDescribed{
				Description: gog.Ptr(models.Description{
					Name:    &u.Description.Name,
					Type:    &u.Description.Type,
					Short:   u.Description.Short,
					Long:    u.Description.Long,
					Further: u.Description.Further,
					Thumb:   u.Description.Thumb,
					Image:   u.Description.Image,
				}),
				URL:             gog.Ptr(u.URL),
				StreamsRequired: u.StreamsRequired,
			})
			uids = append(uids, uid)

		}

		am := models.Activity{
			Description: gog.Ptr(models.Description{
				Name:    &a.Description.Name,
				Type:    &a.Description.Type,
				Short:   a.Description.Short,
				Long:    a.Description.Long,
				Further: a.Description.Further,
				Thumb:   a.Description.Thumb,
				Image:   a.Description.Image,
			}),
			Config:  a.ConfigURL,
			Nbf:     gog.Ptr(float64(a.NotBefore.Unix())),
			Exp:     gog.Ptr(float64(a.ExpiresAt.Unix())),
			Streams: streams,
			Uis:     uids,
		}

		log.WithFields(log.Fields{"user": params.UserName, "booking": params.BookingName}).Info("get activity from booking successful")
		return users.NewGetActivityOK().WithPayload(&am)

	}
}

// getOldBookingsForHandler
func getOldBookingsForUserHandler(config config.ServerConfig) func(users.GetOldBookingsForUserParams, interface{}) middleware.Responder {
	return func(params users.GetOldBookingsForUserParams, principal interface{}) middleware.Responder {

		isAdmin, claims, err := isAdminOrUser(principal)

		if err != nil {
			c := "401"
			m := err.Error()
			return users.NewGetOldBookingsForUserUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		if params.UserName == "" {
			c := "404"
			m := "no user_name in query"
			return users.NewGetOldBookingsForUserNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		// check username against token, unless admin (admin can check on behalf of users)
		if (!isAdmin) && (claims.Subject != params.UserName) {
			c := "401"
			m := "user_name in path does not match subject in token"
			return users.NewGetOldBookingsForUserUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		bs, err := config.Store.GetOldBookingsFor(params.UserName)

		if err != nil {
			c := "404"
			m := err.Error()
			return users.NewGetOldBookingsForUserNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		var bm models.Bookings

		for _, v := range bs {

			b := models.Booking{
				Cancelled:   v.Cancelled,
				Name:        gog.Ptr(v.Name),
				Policy:      gog.Ptr(v.Policy),
				Slot:        gog.Ptr(v.Slot),
				Started:     v.Started,
				Unfulfilled: v.Unfulfilled,
				User:        gog.Ptr(v.User),
				When: gog.Ptr(models.Interval{
					Start: strfmt.DateTime(v.When.Start),
					End:   strfmt.DateTime(v.When.End),
				}),
			}
			bm = append(bm, &b)
		}

		return users.NewGetOldBookingsForUserOK().WithPayload(bm)
	}
}

// getGroupsForUserHandler - includes describedGroups, but only policy names.
func getGroupsForUserHandler(config config.ServerConfig) func(users.GetGroupsForUserParams, interface{}) middleware.Responder {
	return func(params users.GetGroupsForUserParams, principal interface{}) middleware.Responder {

		isAdmin, claims, err := isAdminOrUser(principal)

		if err != nil {
			c := "401"
			m := err.Error()
			return users.NewGetGroupsForUserUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		if params.UserName == "" {
			c := "404"
			m := "no user_name in path"
			return users.NewGetGroupsForUserNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		// check username against token, unless admin (admin can check on behalf of users)
		if (!isAdmin) && (claims.Subject != params.UserName) {
			c := "401"
			m := "user_name in path does not match subject in token"
			return users.NewGetGroupsForUserUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		gs, err := config.Store.GetGroupsFor(params.UserName)

		if err != nil {
			c := "404"
			m := err.Error()
			return users.NewGetGroupsForUserNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		gm := make(map[string]models.GroupDescribed)

		for _, v := range gs {

			g, err := config.Store.GetGroup(v)

			if err != nil {
				c := "500"
				m := "policy " + v + ": " + err.Error()
				return users.NewGetGroupsForUserInternalServerError().WithPayload(&models.Error{Code: &c, Message: &m})
			}

			d := g.Description

			gd := models.GroupDescribed{
				Description: gog.Ptr(models.Description{
					Name:    &d.Name,
					Type:    &d.Type,
					Short:   d.Short,
					Long:    d.Long,
					Further: d.Further,
					Thumb:   d.Thumb,
					Image:   d.Image,
				}),
			}
			gm[v] = gd
		}

		return users.NewGetGroupsForUserOK().WithPayload(gm)
	}
}

// getPolicyStatusForHandler
func getPolicyStatusForUserHandler(config config.ServerConfig) func(users.GetPolicyStatusForUserParams, interface{}) middleware.Responder {
	return func(params users.GetPolicyStatusForUserParams, principal interface{}) middleware.Responder {

		isAdmin, claims, err := isAdminOrUser(principal)
		if err != nil {
			c := "401"
			m := err.Error()
			return users.NewGetPolicyStatusForUserUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		if params.UserName == "" {
			c := "404"
			m := "no user_name in path"
			return users.NewGetPolicyStatusForUserNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		if params.PolicyName == "" {
			c := "404"
			m := "no policy_name in path"
			return users.NewGetPolicyStatusForUserNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		// check username against token, unless admin (admin can check on behalf of users)
		if (!isAdmin) && (claims.Subject != params.UserName) {
			c := "401"
			m := "user_name in path does not match subject in token"
			return users.NewGetPolicyStatusForUserUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		ps, err := config.Store.GetPolicyStatusFor(params.UserName, params.PolicyName)

		if err != nil {
			c := "404"
			m := err.Error()
			return users.NewGetPolicyStatusForUserNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		pm := models.PolicyStatus{
			CurrentBookings: gog.Ptr(int64(ps.CurrentBookings)),
			OldBookings:     gog.Ptr(int64(ps.OldBookings)),
			Usage:           gog.Ptr(ps.Usage.String()),
		}

		return users.NewGetPolicyStatusForUserOK().WithPayload(&pm)
	}
}

// addGroupForHandler
func addGroupForUserHandler(config config.ServerConfig) func(users.AddGroupForUserParams, interface{}) middleware.Responder {
	return func(params users.AddGroupForUserParams, principal interface{}) middleware.Responder {

		isAdmin, claims, err := isAdminOrUser(principal)
		if err != nil {
			c := "401"
			m := err.Error()
			return users.NewAddGroupForUserUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		if params.UserName == "" {
			c := "404"
			m := "no user_name in path"
			return users.NewAddGroupForUserNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		if params.GroupName == "" {
			c := "404"
			m := "no group_name in path"
			return users.NewAddGroupForUserNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		// check username against token, unless admin (admin can check on behalf of users)
		if (!isAdmin) && (claims.Subject != params.UserName) {
			c := "401"
			m := "user_name in path does not match subject in token"
			return users.NewAddGroupForUserUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		err = config.Store.AddGroupForUser(params.UserName, params.GroupName)

		if err != nil {
			c := "404"
			m := err.Error()
			return users.NewAddGroupForUserNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		return users.NewAddGroupForUserNoContent()
	}
}
