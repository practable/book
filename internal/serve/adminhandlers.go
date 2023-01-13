package serve

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/icza/gog"
	log "github.com/sirupsen/logrus"
	"github.com/timdrysdale/interval/internal/config"
	dt "github.com/timdrysdale/interval/internal/datetime"
	"github.com/timdrysdale/interval/internal/interval"
	"github.com/timdrysdale/interval/internal/serve/models"
	"github.com/timdrysdale/interval/internal/serve/restapi/operations/admin"
	"github.com/timdrysdale/interval/internal/store"
	"gopkg.in/yaml.v2"
)

// checkManifestHandler
func checkManifestHandler(config config.ServerConfig) func(admin.CheckManifestParams, interface{}) middleware.Responder {
	return func(params admin.CheckManifestParams, principal interface{}) middleware.Responder {

		_, err := isAdmin(principal)

		if err != nil {
			c := "401"
			m := "no scope booking:admin"
			return admin.NewCheckManifestUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		sm, err := convertModelsManifestToStore(*params.Manifest)
		if err != nil {
			c := "500"
			m := err.Error()
			return admin.NewCheckManifestInternalServerError().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		err, msgs := store.CheckManifest(sm)

		if err != nil {
			c := "500"
			m := strings.Join(msgs, ",")
			return admin.NewCheckManifestInternalServerError().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		return admin.NewCheckManifestNoContent()
	}
}

// convertStoreStatusAdminToModel converts from internal to API type
func convertStoreStatusAdminToModel(s store.StoreStatusAdmin) (models.StoreStatusAdmin, error) {
	var m models.StoreStatusAdmin

	y, err := json.Marshal(s)

	if err != nil {
		return m, err
	}

	err = json.Unmarshal(y, &m)

	return m, err

}

/*
// convertBookingsToStore converts from YAML string to internal type
func convertBookingsToStore(m string) (map[string]store.Booking, error) {

	var s map[string]store.Booking

	err := yaml.Unmarshal([]byte(m), &s)

	return s, err
}*/

func convertBookingsToStore(m models.Bookings) (map[string]store.Booking, error) {

	sm := make(map[string]store.Booking)

	for _, v := range m {
		start, err := dt.Parse(v.When.Start.String())
		if err != nil {
			return sm, err
		}
		end, err := dt.Parse(v.When.End.String())
		if err != nil {
			return sm, err
		}
		b := store.Booking{
			Name:        *v.Name,
			Policy:      *v.Policy,
			Slot:        *v.Slot,
			User:        *v.User,
			Cancelled:   v.Cancelled,
			Started:     v.Started,
			Unfulfilled: v.Unfulfilled,
			When: interval.Interval{
				Start: start,
				End:   end,
			},
		}
		sm[b.Name] = b
	}

	return sm, nil
}

// convertManifestToStore converts from YAML string to internal type
func convertManifestToStore(m string) (store.Manifest, error) {

	var s store.Manifest

	err := yaml.Unmarshal([]byte(m), &s)

	return s, err
}

func convertModelsManifestToStore(mm models.Manifest) (store.Manifest, error) {

	dm := make(map[string]store.Description)

	for k, v := range mm.Descriptions {
		m := v
		dm[k] = store.Description{
			Name:    *(m.Name),
			Short:   m.Short,
			Type:    *(m.Type),
			Long:    m.Long,
			Further: m.Further,
			Thumb:   m.Thumb,
			Image:   m.Image,
		}
	}

	dgm := make(map[string]store.DisplayGuide)

	for k, v := range mm.DisplayGuides {
		m := v
		ba, err := time.ParseDuration(*m.BookAhead)
		if err != nil {
			return store.Manifest{}, err
		}
		dd, err := time.ParseDuration(*m.Duration)
		if err != nil {
			return store.Manifest{}, err
		}
		dgm[k] = store.DisplayGuide{
			BookAhead: ba,
			Duration:  dd,
			MaxSlots:  int(*(m.MaxSlots)),
			Label:     *(m.Label),
		}
	}

	pm := make(map[string]store.Policy)

	for k, v := range mm.Policies {
		m := v

		var ba, gpd, gpy, nd, xd, mu, na, sp, sw time.Duration
		var err error

		if m.EnforceBookAhead { //&& m.BookAhead != "" {
			ba, err = time.ParseDuration(m.BookAhead)
			if err != nil {
				return store.Manifest{}, errors.New("error parsing duration book_ahead in policy " + k + " is " + err.Error())
			}
		}

		if m.EnforceMinDuration {
			nd, err = time.ParseDuration(m.MinDuration)
			if err != nil {
				return store.Manifest{}, errors.New("error parsing duration min_duration in policy " + k + " is " + err.Error())
			}
		}

		if m.EnforceMaxDuration { //&& m.MaxDuration != "" {
			xd, err = time.ParseDuration(m.MaxDuration)
			if err != nil {
				return store.Manifest{}, errors.New("error parsing duration max_duration in policy " + k + " is " + err.Error())
			}
		}
		if m.EnforceMaxUsage { // && m.MaxUsage != "" {
			mu, err = time.ParseDuration(m.MaxUsage)
			if err != nil {
				return store.Manifest{}, errors.New("error parsing duration max_usage in policy " + k + " is " + err.Error())
			}
		}

		if m.EnforceNextAvailable { // && m.NextAvailable != "" {
			na, err = time.ParseDuration(m.NextAvailable)
			if err != nil {
				return store.Manifest{}, errors.New("error parsing duration next_available in policy " + k + " is " + err.Error())
			}
		}

		if m.EnforceAllowStartInPast { //&& m.AllowStartInPastWithin != "" {
			sp, err = time.ParseDuration(m.AllowStartInPastWithin)
			if err != nil {
				return store.Manifest{}, errors.New("error parsing duration allow_start_in_past_within in policy " + k + " is " + err.Error())
			}
		}

		if m.EnforceStartsWithin { //&& m.StartsWithin != "" {
			sw, err = time.ParseDuration(m.StartsWithin)
			if err != nil {
				return store.Manifest{}, errors.New("error parsing duration starts_within in policy " + k + " is " + err.Error())
			}
		}

		if m.EnforceGracePeriod {

			//if m.GracePeriod != "" {
			gpd, err = time.ParseDuration(m.GracePeriod)
			if err != nil {
				return store.Manifest{}, errors.New("error parsing duration grace_period in policy " + k + " is " + err.Error())
			}
			//}

			//if m.GracePenalty != "" {
			gpy, err = time.ParseDuration(m.GracePenalty)
			if err != nil {
				return store.Manifest{}, errors.New("error parsing duration grace_penalty in policy " + k + " is " + err.Error())
			}
			//}
		}
		pm[k] = store.Policy{
			AllowStartInPastWithin:  sp,
			BookAhead:               ba,
			Description:             *(m.Description),
			DisplayGuides:           m.DisplayGuides,
			EnforceAllowStartInPast: m.EnforceAllowStartInPast,
			EnforceBookAhead:        m.EnforceBookAhead,
			EnforceGracePeriod:      m.EnforceGracePeriod,
			EnforceMaxBookings:      m.EnforceMaxBookings,
			EnforceMaxDuration:      m.EnforceMaxDuration,
			EnforceMinDuration:      m.EnforceMinDuration,
			EnforceMaxUsage:         m.EnforceMaxUsage,
			EnforceNextAvailable:    m.EnforceNextAvailable,
			EnforceStartsWithin:     m.EnforceStartsWithin,
			EnforceUnlimitedUsers:   m.EnforceUnlimitedUsers,
			GracePenalty:            gpy,
			GracePeriod:             gpd,
			MaxBookings:             m.MaxBookings,
			MaxDuration:             xd,
			MinDuration:             nd,
			MaxUsage:                mu,
			NextAvailable:           na,
			Slots:                   m.Slots,
			StartsWithin:            sw,
		}
	}

	rm := make(map[string]store.Resource)

	for k, v := range mm.Resources {
		m := v
		rm[k] = store.Resource{
			ConfigURL:   m.ConfigURL,
			Description: *(m.Description),
			Streams:     m.Streams,
			TopicStub:   *(m.TopicStub),
		}
	}

	slm := make(map[string]store.Slot)

	for k, v := range mm.Slots {
		m := v
		slm[k] = store.Slot{
			Description: *(m.Description),
			Policy:      *(m.Policy),
			Resource:    *(m.Resource),
			UISet:       *(m.UISet),
			Window:      *(m.Window),
		}
	}

	stm := make(map[string]store.Stream)

	for k, v := range mm.Streams {
		m := v
		stm[k] = store.Stream{
			ConnectionType: *(m.ConnectionType),
			For:            *(m.For),
			Scopes:         m.Scopes,
			Topic:          *(m.Topic),
			URL:            *(m.URL),
		}
	}

	uim := make(map[string]store.UI)

	for k, v := range mm.Uis {
		m := v
		uim[k] = store.UI{
			Description:     *(m.Description),
			StreamsRequired: m.StreamsRequired,
			URL:             *(m.URL),
		}
	}

	usm := make(map[string]store.UISet)

	for k, v := range mm.UISets {
		m := v
		usm[k] = store.UISet{
			UIs: m.UIs,
		}
	}

	wm := make(map[string]store.Window)

	for k, v := range mm.Windows {
		m := v

		aa := []interval.Interval{}
		dd := []interval.Interval{}

		for _, mi := range m.Allowed {

			st, err := dt.Parse(mi.Start.String())
			if err != nil {
				return store.Manifest{}, err
			}
			et, err := dt.Parse(mi.End.String())
			if err != nil {
				return store.Manifest{}, err
			}
			mi := interval.Interval{
				Start: st,
				End:   et,
			}
			aa = append(aa, mi)
		}
		for _, mi := range m.Denied {

			st, err := dt.Parse(mi.Start.String())
			if err != nil {
				return store.Manifest{}, err
			}
			et, err := dt.Parse(mi.End.String())
			if err != nil {
				return store.Manifest{}, err
			}
			mi := interval.Interval{
				Start: st,
				End:   et,
			}

			dd = append(dd, mi)
		}

		wm[k] = store.Window{
			Allowed: aa,
			Denied:  dd,
		}
	}

	sm := store.Manifest{
		Descriptions:  dm,
		DisplayGuides: dgm,
		Policies:      pm,
		Resources:     rm,
		Slots:         slm,
		Streams:       stm,
		UIs:           uim,
		UISets:        usm,
		Windows:       wm,
	}

	return sm, nil

}

// exportBookingsHandler
// https://github.com/go-swagger/go-swagger/issues/2275
func exportBookingsHandler(config config.ServerConfig) func(admin.ExportBookingsParams, interface{}) middleware.Responder {
	return func(params admin.ExportBookingsParams, principal interface{}) middleware.Responder {

		_, err := isAdmin(principal)

		if err != nil {
			c := "401"
			m := "no scope booking:admin"
			return admin.NewExportBookingsUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		bs := config.Store.ExportBookings()

		bm := []*models.Booking{}

		for _, v := range bs {

			_, err := json.Marshal(v)

			if err != nil {
				c := "500"
				m := err.Error()
				return admin.NewExportBookingsInternalServerError().WithPayload(&models.Error{Code: &c, Message: &m})
			}

			b := models.Booking{

				Name:      gog.Ptr(v.Name),
				Policy:    gog.Ptr(v.Policy),
				Slot:      gog.Ptr(v.Slot),
				User:      gog.Ptr(v.User),
				Cancelled: v.Cancelled,

				Started:     v.Started,
				Unfulfilled: v.Unfulfilled,

				When: gog.Ptr(models.Interval{
					Start: strfmt.DateTime(v.When.Start),
					End:   strfmt.DateTime(v.When.End),
				}),
			}

			bm = append(bm, &b)

		}

		log.Debugf("exported " + strconv.Itoa(len(bm)) + " bookings")
		return admin.NewExportBookingsOK().WithPayload(bm)
	}
}

// exportManifestHandler
func exportManifestHandler(config config.ServerConfig) func(admin.ExportManifestParams, interface{}) middleware.Responder {
	return func(params admin.ExportManifestParams, principal interface{}) middleware.Responder {

		_, err := isAdmin(principal)

		if err != nil {
			c := "401"
			m := "no scope booking:admin"
			return admin.NewExportManifestUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		sm := config.Store.ExportManifest()

		dm := make(map[string]models.Description)

		for k, v := range sm.Descriptions {
			s := v
			dm[k] = models.Description{
				Name:    gog.Ptr(s.Name),
				Short:   s.Short,
				Type:    gog.Ptr(s.Type),
				Long:    s.Long,
				Further: s.Further,
				Thumb:   s.Thumb,
				Image:   s.Image,
			}
		}

		dgm := make(map[string]models.DisplayGuide)

		for k, v := range sm.DisplayGuides {
			s := v
			dgm[k] = models.DisplayGuide{
				BookAhead: gog.Ptr(s.BookAhead.String()),
				Duration:  gog.Ptr(s.Duration.String()),
				MaxSlots:  gog.Ptr(int64(s.MaxSlots)),
				Label:     gog.Ptr(s.Label),
			}
		}

		pm := make(map[string]models.Policy)

		for k, v := range sm.Policies {
			s := v

			pm[k] = models.Policy{
				AllowStartInPastWithin:  s.AllowStartInPastWithin.String(),
				BookAhead:               s.BookAhead.String(),
				Description:             gog.Ptr(s.Description),
				DisplayGuides:           s.DisplayGuides,
				EnforceAllowStartInPast: s.EnforceAllowStartInPast,
				EnforceBookAhead:        s.EnforceBookAhead,
				EnforceGracePeriod:      s.EnforceGracePeriod,
				EnforceMaxBookings:      s.EnforceMaxBookings,
				EnforceMaxDuration:      s.EnforceMaxDuration,
				EnforceMinDuration:      s.EnforceMinDuration,
				EnforceMaxUsage:         s.EnforceMaxUsage,
				EnforceNextAvailable:    s.EnforceNextAvailable,
				EnforceStartsWithin:     s.EnforceStartsWithin,
				EnforceUnlimitedUsers:   s.EnforceUnlimitedUsers,
				GracePenalty:            s.GracePenalty.String(),
				GracePeriod:             s.GracePeriod.String(),
				MaxBookings:             s.MaxBookings,
				MaxDuration:             s.MaxDuration.String(),
				MinDuration:             s.MinDuration.String(),
				MaxUsage:                s.MaxUsage.String(),
				NextAvailable:           s.NextAvailable.String(),
				Slots:                   s.Slots,
				StartsWithin:            s.StartsWithin.String(),
			}
		}

		rm := make(map[string]models.Resource)

		for k, v := range sm.Resources {
			s := v
			rm[k] = models.Resource{
				ConfigURL:   s.ConfigURL,
				Description: gog.Ptr(s.Description),
				Streams:     s.Streams,
				TopicStub:   gog.Ptr(s.TopicStub),
			}
		}

		slm := make(map[string]models.Slot)

		for k, v := range sm.Slots {
			s := v
			slm[k] = models.Slot{
				Description: gog.Ptr(s.Description),
				Policy:      gog.Ptr(s.Policy),
				Resource:    gog.Ptr(s.Resource),
				UISet:       gog.Ptr(s.UISet),
				Window:      gog.Ptr(s.Window),
			}
		}

		stm := make(map[string]models.ManifestStream)

		for k, v := range sm.Streams {
			s := v
			stm[k] = models.ManifestStream{
				ConnectionType: gog.Ptr(s.ConnectionType),
				For:            gog.Ptr(s.For),
				Scopes:         s.Scopes,
				Topic:          gog.Ptr(s.Topic),
				URL:            gog.Ptr(s.URL),
			}
		}

		uim := make(map[string]models.UI)

		for k, v := range sm.UIs {
			s := v
			uim[k] = models.UI{
				Description:     gog.Ptr(s.Description),
				StreamsRequired: s.StreamsRequired,
				URL:             gog.Ptr(s.URL),
			}
		}

		usm := make(map[string]models.UISet)

		for k, v := range sm.UISets {
			s := v
			usm[k] = models.UISet{
				UIs: s.UIs,
			}
		}

		wm := make(map[string]models.Window)

		for k, v := range sm.Windows {
			s := v

			aa := []*models.Interval{}
			dd := []*models.Interval{}

			for _, si := range s.Allowed {
				mi := models.Interval{
					Start: strfmt.DateTime(si.Start),
					End:   strfmt.DateTime(si.End),
				}
				aa = append(aa, &mi)
			}
			for _, si := range s.Denied {
				mi := models.Interval{
					Start: strfmt.DateTime(si.Start),
					End:   strfmt.DateTime(si.End),
				}
				dd = append(dd, &mi)
			}

			wm[k] = models.Window{
				Allowed: aa,
				Denied:  dd,
			}
		}

		mm := models.Manifest{
			Descriptions:  dm,
			DisplayGuides: dgm,
			Policies:      pm,
			Resources:     rm,
			Slots:         slm,
			Streams:       stm,
			Uis:           uim,
			UISets:        usm,
			Windows:       wm,
		}

		return admin.NewExportManifestOK().WithPayload(&mm)
	}
}

// exportOldBookingsHandler
func exportOldBookingsHandler(config config.ServerConfig) func(admin.ExportOldBookingsParams, interface{}) middleware.Responder {
	return func(params admin.ExportOldBookingsParams, principal interface{}) middleware.Responder {

		_, err := isAdmin(principal)

		if err != nil {
			c := "401"
			m := "no scope booking:admin"
			return admin.NewExportOldBookingsUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		bs := config.Store.ExportOldBookings()

		bm := []*models.Booking{}

		for _, v := range bs {

			_, err := json.Marshal(v)

			if err != nil {
				c := "500"
				m := err.Error()
				return admin.NewExportBookingsInternalServerError().WithPayload(&models.Error{Code: &c, Message: &m})
			}

			b := models.Booking{

				Name:      gog.Ptr(v.Name),
				Policy:    gog.Ptr(v.Policy),
				Slot:      gog.Ptr(v.Slot),
				User:      gog.Ptr(v.User),
				Cancelled: v.Cancelled,

				Started:     v.Started,
				Unfulfilled: v.Unfulfilled,

				When: gog.Ptr(models.Interval{
					Start: strfmt.DateTime(v.When.Start),
					End:   strfmt.DateTime(v.When.End),
				}),
			}

			bm = append(bm, &b)

		}

		log.Debugf("exported " + strconv.Itoa(len(bm)) + " old bookings")

		return admin.NewExportOldBookingsOK().WithPayload(bm)
	}
}

// exportUsersHandler
func exportUsersHandler(config config.ServerConfig) func(admin.ExportUsersParams, interface{}) middleware.Responder {
	return func(params admin.ExportUsersParams, principal interface{}) middleware.Responder {

		_, err := isAdmin(principal)

		if err != nil {
			c := "401"
			m := "no scope booking:admin"
			return admin.NewExportUsersUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		su := config.Store.ExportUsers()
		var mu models.Users

		mu = make(map[string]models.User)

		for k, v := range su {

			bs := []string{}
			obs := []string{}
			ps := []string{}
			um := make(map[string]string)

			for _, bv := range v.Bookings {
				bs = append(bs, bv)
			}

			for _, obv := range v.OldBookings {
				obs = append(obs, obv)
			}

			// ignore bool in map, has no meaning
			for _, pv := range v.Policies {
				ps = append(ps, pv)
			}

			// store format is map[string]*time.Duration
			for uk, uv := range v.Usage {
				um[uk] = uv
			}

			m := models.User{

				Bookings:    bs,
				OldBookings: obs,
				Policies:    ps,
				Usage:       um,
			}

			mu[k] = m
		}

		return admin.NewExportUsersOK().WithPayload(mu)
	}
}

// getStoreStatusAdminHandler
func getStoreStatusAdminHandler(config config.ServerConfig) func(admin.GetStoreStatusAdminParams, interface{}) middleware.Responder {
	return func(params admin.GetStoreStatusAdminParams, principal interface{}) middleware.Responder {

		_, err := isAdmin(principal)

		if err != nil {
			c := "401"
			m := "no scope booking:admin"
			return admin.NewGetStoreStatusAdminUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		s, err := convertStoreStatusAdminToModel(config.Store.GetStoreStatusAdmin())

		if err != nil {
			log.Error("could not convert StoreStatusAdmin to model format")
		}

		return admin.NewGetStoreStatusAdminOK().WithPayload(&s)
	}
}

// getSlotIsAvailableHandlerFunc
func getSlotIsAvailableHandler(config config.ServerConfig) func(admin.GetSlotIsAvailableParams, interface{}) middleware.Responder {
	return func(params admin.GetSlotIsAvailableParams, principal interface{}) middleware.Responder {

		_, err := isAdmin(principal)

		if err != nil {
			c := "401"
			m := "no scope booking:admin"
			return admin.NewGetSlotIsAvailableUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		avail, reason, err := config.Store.GetSlotIsAvailable(params.SlotName)

		if err != nil {
			c := "404"
			m := err.Error()
			return admin.NewGetSlotIsAvailableNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		s := models.SlotStatus{
			Available: &avail,
			Reason:    &reason,
		}
		return admin.NewGetSlotIsAvailableOK().WithPayload(&s)
	}
}

// replaceBookingsHandler
func replaceBookingsHandler(config config.ServerConfig) func(admin.ReplaceBookingsParams, interface{}) middleware.Responder {
	return func(params admin.ReplaceBookingsParams, principal interface{}) middleware.Responder {

		_, err := isAdmin(principal)

		if err != nil {
			c := "401"
			m := "no scope booking:admin"
			return admin.NewReplaceBookingsUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		//params.Bookings is array of bookings, need a map
		sm, err := convertBookingsToStore(params.Bookings)
		if err != nil {
			c := "500"
			m := "error parsing bookings: " + err.Error()
			return admin.NewReplaceBookingsInternalServerError().WithPayload(&models.Error{Code: &c, Message: &m})
		}
		log.Debugf("replaced " + strconv.Itoa(len(sm)) + " bookings")

		err, msgs := config.Store.ReplaceBookings(sm)
		if err != nil {
			c := "500"
			m := err.Error() + " : " + strings.Join(msgs, ",")
			return admin.NewReplaceBookingsInternalServerError().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		s, err := convertStoreStatusAdminToModel(config.Store.GetStoreStatusAdmin())

		if err != nil {
			log.Error("could not convert StoreStatusAdmin to model format")
		}

		return admin.NewReplaceBookingsOK().WithPayload(&s)
	}
}

// replaceManifestHandler
func replaceManifestHandler(config config.ServerConfig) func(admin.ReplaceManifestParams, interface{}) middleware.Responder {
	return func(params admin.ReplaceManifestParams, principal interface{}) middleware.Responder {

		_, err := isAdmin(principal)

		if err != nil {
			c := "401"
			m := "no scope booking:admin"
			return admin.NewReplaceManifestUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		sm, err := convertModelsManifestToStore(*params.Manifest)
		if err != nil {
			c := "500"
			m := err.Error()
			return admin.NewReplaceManifestInternalServerError().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		err = config.Store.ReplaceManifest(sm)
		if err != nil {
			c := "500"
			m := err.Error()
			return admin.NewReplaceManifestInternalServerError().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		s, err := convertStoreStatusAdminToModel(config.Store.GetStoreStatusAdmin())

		if err != nil {
			log.Error("could not convert StoreStatusAdmin to model format")
		}

		return admin.NewReplaceManifestOK().WithPayload(&s)
	}
}

// replaceOldBookingsHandler
func replaceOldBookingsHandler(config config.ServerConfig) func(admin.ReplaceOldBookingsParams, interface{}) middleware.Responder {
	return func(params admin.ReplaceOldBookingsParams, principal interface{}) middleware.Responder {

		_, err := isAdmin(principal)

		if err != nil {
			c := "401"
			m := "no scope booking:admin"
			return admin.NewReplaceOldBookingsUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		sm, err := convertBookingsToStore(params.Bookings)
		if err != nil {
			c := "500"
			m := err.Error()
			return admin.NewReplaceOldBookingsInternalServerError().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		err, msgs := config.Store.ReplaceOldBookings(sm)
		if err != nil {
			c := "500"
			m := err.Error() + " : " + strings.Join(msgs, ",")
			return admin.NewReplaceOldBookingsInternalServerError().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		s, err := convertStoreStatusAdminToModel(config.Store.GetStoreStatusAdmin())

		if err != nil {
			log.Error("could not convert StoreStatusAdmin to model format")
		}

		return admin.NewReplaceOldBookingsOK().WithPayload(&s)
	}
}

// setStoreStatusAdminHandler
func setLockHandler(config config.ServerConfig) func(admin.SetLockParams, interface{}) middleware.Responder {
	return func(params admin.SetLockParams, principal interface{}) middleware.Responder {

		_, err := isAdmin(principal)

		if err != nil {
			c := "401"
			m := "no scope booking:admin"
			return admin.NewSetLockUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		config.Store.Locked = params.Lock

		if params.Msg != nil {
			config.Store.Message = *(params.Msg)
		}

		s, err := convertStoreStatusAdminToModel(config.Store.GetStoreStatusAdmin())

		if err != nil {
			log.Error("could not convert StoreStatusAdmin to model format")
		}

		return admin.NewSetLockOK().WithPayload(&s)
	}
}

// setSlotIsAvailableHandlerFunc
func setSlotIsAvailableHandler(config config.ServerConfig) func(admin.SetSlotIsAvailableParams, interface{}) middleware.Responder {
	return func(params admin.SetSlotIsAvailableParams, principal interface{}) middleware.Responder {

		_, err := isAdmin(principal)

		if err != nil {
			c := "401"
			m := "no scope booking:admin"
			return admin.NewSetSlotIsAvailableUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		err = config.Store.SetSlotIsAvailable(params.SlotName, params.Available, params.Reason)

		if err != nil {
			c := "404"
			m := err.Error()
			return admin.NewSetSlotIsAvailableNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		return admin.NewSetSlotIsAvailableNoContent()
	}
}
