package serve

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/icza/gog"
	"github.com/practable/book/internal/config"
	dt "github.com/practable/book/internal/datetime"
	"github.com/practable/book/internal/interval"
	operational "github.com/practable/book/internal/operations"
	"github.com/practable/book/internal/serve/models"
	"github.com/practable/book/internal/serve/restapi/operations/admin"
	"github.com/practable/book/internal/store"
	log "github.com/sirupsen/logrus"
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
	jobTemplates, pipelineTemplates, err := modelOperationalTemplatesToStore(mm)
	if err != nil {
		return store.Manifest{}, err
	}
	workflows := make(map[string]store.OperationalWorkflow, len(mm.OperationalWorkflows))
	for name, value := range mm.OperationalWorkflows {
		expected, err := time.ParseDuration(*value.ExpectedDuration)
		if err != nil {
			return store.Manifest{}, errors.New("error parsing expected_duration in operational workflow " + name + ": " + err.Error())
		}
		maximum, err := time.ParseDuration(*value.MaximumDuration)
		if err != nil {
			return store.Manifest{}, errors.New("error parsing maximum_duration in operational workflow " + name + ": " + err.Error())
		}
		workflows[name] = store.OperationalWorkflow{Description: *value.Description, Kind: value.Kind, ExpectedDuration: expected, MaximumDuration: maximum}
	}
	schedules := make(map[string]store.OperationalSchedule, len(mm.OperationalSchedules))
	for name, value := range mm.OperationalSchedules {
		duration, err := time.ParseDuration(*value.Duration)
		if err != nil {
			return store.Manifest{}, errors.New("error parsing duration in operational schedule " + name + ": " + err.Error())
		}
		exceptions := make([]string, 0, len(value.Recurrence.Exceptions))
		for _, exception := range value.Recurrence.Exceptions {
			exceptions = append(exceptions, exception.String())
		}
		schedules[name] = store.OperationalSchedule{Slot: *value.Slot, Workflow: *value.Workflow, Duration: duration, Conflict: *value.Conflict,
			Recurrence: store.OperationalRecurrence{Timezone: *value.Recurrence.Timezone, StartDate: value.Recurrence.StartDate.String(),
				EndDate: value.Recurrence.EndDate.String(), Weekdays: value.Recurrence.Weekdays, Time: *value.Recurrence.Time, Exceptions: exceptions}}
	}

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

	gm := make(map[string]store.Group)

	for k, v := range mm.Groups {
		m := v
		gm[k] = store.Group{
			Description: *(m.Description),
			Policies:    m.Policies,
		}
	}

	//debug - groups info is getting into gm ok:
	//fmt.Printf("\n\ngroups: %+v\n\n", gm)
	// groups: map[g-a:{Description:d-g-a Policies:[p-a]} g-b:{Description:d-g-b Policies:[p-b]}]

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
		operations, err := modelOperationalProfileToStore(m.Operations)
		if err != nil {
			return store.Manifest{}, errors.New("resource " + k + ": " + err.Error())
		}
		rm[k] = store.Resource{
			ConfigURL:        m.ConfigURL,
			Class:            m.Class,
			Description:      *(m.Description),
			Properties:       m.Properties,
			Operations:       operations,
			StreamOperations: modelStreamOperationsToStore(m.StreamOperations),
			Streams:          m.Streams,
			Tests:            m.Tests,
			TopicStub:        *(m.TopicStub),
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

		convertRecurrences := func(values []*models.WeeklyRecurrence) []store.WeeklyRecurrence {
			var result []store.WeeklyRecurrence
			for _, value := range values {
				if value == nil {
					continue
				}
				exceptions := make([]string, 0, len(value.Exceptions))
				for _, exception := range value.Exceptions {
					exceptions = append(exceptions, exception.String())
				}
				result = append(result, store.WeeklyRecurrence{Timezone: *value.Timezone, StartDate: value.StartDate.String(), EndDate: value.EndDate.String(), Weekdays: value.Weekdays, StartTime: *value.StartTime, EndTime: *value.EndTime, Exceptions: exceptions})
			}
			return result
		}
		wm[k] = store.Window{Allowed: aa, Denied: dd, RecurringAllowed: convertRecurrences(m.RecurringAllowed), RecurringDenied: convertRecurrences(m.RecurringDenied)}
	}

	sm := store.Manifest{
		Descriptions:                 dm,
		DisplayGuides:                dgm,
		Groups:                       gm,
		Policies:                     pm,
		OperationalWorkflows:         workflows,
		OperationalSchedules:         schedules,
		OperationalJobTemplates:      jobTemplates,
		OperationalPipelineTemplates: pipelineTemplates,
		Resources:                    rm,
		Slots:                        slm,
		Streams:                      stm,
		UIs:                          uim,
		UISets:                       usm,
		Windows:                      wm,
	}

	return sm, nil

}

func modelOperationalProfileToStore(profile *models.OperationalProfile) (store.OperationalProfile, error) {
	if profile == nil {
		return store.OperationalProfile{}, nil
	}
	convert := func(values []*models.OperationalGuard) ([]store.OperationalGuard, error) {
		result := make([]store.OperationalGuard, 0, len(values))
		for _, value := range values {
			if value == nil {
				continue
			}
			duration, err := time.ParseDuration(*value.Duration)
			if err != nil {
				return nil, errors.New("invalid operational guard duration: " + err.Error())
			}
			result = append(result, store.OperationalGuard{Workflow: *value.Workflow, Duration: duration, Applies: *value.Applies, Reclaimable: value.Reclaimable})
		}
		return result, nil
	}
	before, err := convert(profile.BeforeBooking)
	if err != nil {
		return store.OperationalProfile{}, err
	}
	after, err := convert(profile.AfterBooking)
	if err != nil {
		return store.OperationalProfile{}, err
	}
	return store.OperationalProfile{OperatingWindow: profile.OperatingWindow, CostOwner: profile.CostOwner, BeforeBooking: before, AfterBooking: after}, nil
}

func storeOperationalProfileToModel(profile store.OperationalProfile) *models.OperationalProfile {
	if profile.OperatingWindow == "" && len(profile.BeforeBooking) == 0 && len(profile.AfterBooking) == 0 {
		return nil
	}
	convert := func(values []store.OperationalGuard) []*models.OperationalGuard {
		result := make([]*models.OperationalGuard, 0, len(values))
		for _, value := range values {
			workflow, duration, applies := value.Workflow, value.Duration.String(), value.Applies
			result = append(result, &models.OperationalGuard{Workflow: &workflow, Duration: &duration, Applies: &applies, Reclaimable: value.Reclaimable})
		}
		return result
	}
	return &models.OperationalProfile{OperatingWindow: profile.OperatingWindow, CostOwner: profile.CostOwner, BeforeBooking: convert(profile.BeforeBooking), AfterBooking: convert(profile.AfterBooking)}
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

func bookingToModel(booking store.Booking) *models.Booking {
	model := &models.Booking{
		Name: gog.Ptr(booking.Name), Policy: gog.Ptr(booking.Policy), Slot: gog.Ptr(booking.Slot), User: gog.Ptr(booking.User),
		Cancelled: booking.Cancelled, Started: booking.Started, Unfulfilled: booking.Unfulfilled, Maintenance: booking.Maintenance,
		When: gog.Ptr(models.Interval{Start: strfmt.DateTime(booking.When.Start), End: strfmt.DateTime(booking.When.End)}),
	}
	if !booking.CancelledAt.IsZero() {
		model.CancelledAt = strfmt.DateTime(booking.CancelledAt)
		model.CancelledBy = booking.CancelledBy
	}
	if booking.StartedAt != "" {
		if started, err := dt.Parse(booking.StartedAt); err == nil {
			model.StartedAt = strfmt.DateTime(started)
		}
	}
	if booking.UsageCharged != 0 {
		model.UsageCharged = booking.UsageCharged.String()
	}
	return model
}

func modelToReplacement(model *models.Booking) (store.Booking, error) {
	if model == nil || model.Name == nil || model.Policy == nil || model.Slot == nil || model.User == nil || model.When == nil {
		return store.Booking{}, errors.New("booking name, policy, slot, user, and interval are required")
	}
	start, err := dt.Parse(model.When.Start.String())
	if err != nil {
		return store.Booking{}, err
	}
	end, err := dt.Parse(model.When.End.String())
	if err != nil {
		return store.Booking{}, err
	}
	booking := store.Booking{Name: *model.Name, Policy: *model.Policy, Slot: *model.Slot, User: *model.User,
		Cancelled: model.Cancelled, CancelledBy: model.CancelledBy, Started: model.Started, Unfulfilled: model.Unfulfilled, Maintenance: model.Maintenance,
		When: interval.Interval{Start: start, End: end}}
	if !time.Time(model.CancelledAt).IsZero() {
		booking.CancelledAt = time.Time(model.CancelledAt)
	}
	if !time.Time(model.StartedAt).IsZero() {
		booking.StartedAt = time.Time(model.StartedAt).UTC().Format(time.RFC3339Nano)
	}
	if model.UsageCharged != "" {
		usage, err := time.ParseDuration(model.UsageCharged)
		if err != nil {
			return store.Booking{}, err
		}
		booking.UsageCharged = usage
	}
	return booking, nil
}

func editableBookingToModel(edit store.EditableBooking) *models.BookingEdit {
	return &models.BookingEdit{OriginalName: gog.Ptr(edit.OriginalName), Revision: gog.Ptr(edit.Revision), Booking: bookingToModel(edit.Booking)}
}

func exportBookingForEditHandler(config config.ServerConfig) func(admin.ExportBookingForEditParams, interface{}) middleware.Responder {
	return func(params admin.ExportBookingForEditParams, principal interface{}) middleware.Responder {
		if _, err := isAdmin(principal); err != nil {
			code, message := "401", "no scope booking:admin"
			return admin.NewExportBookingForEditUnauthorized().WithPayload(&models.Error{Code: &code, Message: &message})
		}
		edit, err := config.Store.GetBookingForEdit(params.BookingName)
		if errors.Is(err, store.ErrPersistentNotFound) {
			code, message := "404", err.Error()
			return admin.NewExportBookingForEditNotFound().WithPayload(&models.Error{Code: &code, Message: &message})
		}
		if err != nil {
			code, message := "500", err.Error()
			return admin.NewExportBookingForEditInternalServerError().WithPayload(&models.Error{Code: &code, Message: &message})
		}
		return admin.NewExportBookingForEditOK().WithPayload(editableBookingToModel(edit))
	}
}

func replaceBookingHandler(config config.ServerConfig) func(admin.ReplaceBookingParams, interface{}) middleware.Responder {
	return func(params admin.ReplaceBookingParams, principal interface{}) middleware.Responder {
		if _, err := isAdmin(principal); err != nil {
			code, message := "401", "no scope booking:admin"
			return admin.NewReplaceBookingUnauthorized().WithPayload(&models.Error{Code: &code, Message: &message})
		}
		if params.BookingEdit == nil || params.BookingEdit.OriginalName == nil || params.BookingEdit.Revision == nil {
			code, message := "409", "booking edit original_name and revision are required"
			return admin.NewReplaceBookingConflict().WithPayload(&models.Error{Code: &code, Message: &message})
		}
		if *params.BookingEdit.OriginalName != params.BookingName {
			code, message := "409", "booking edit original_name does not match path"
			return admin.NewReplaceBookingConflict().WithPayload(&models.Error{Code: &code, Message: &message})
		}
		booking, err := modelToReplacement(params.BookingEdit.Booking)
		if err != nil {
			code, message := "409", err.Error()
			return admin.NewReplaceBookingConflict().WithPayload(&models.Error{Code: &code, Message: &message})
		}
		result, err := config.Store.ReplaceBooking(store.EditableBooking{
			OriginalName: *params.BookingEdit.OriginalName, Revision: *params.BookingEdit.Revision, Booking: booking,
		})
		if errors.Is(err, store.ErrPersistentNotFound) {
			code, message := "404", err.Error()
			return admin.NewReplaceBookingNotFound().WithPayload(&models.Error{Code: &code, Message: &message})
		}
		if errors.Is(err, store.ErrBookingRevision) || errors.Is(err, store.ErrBookingStarted) ||
			errors.Is(err, store.ErrInvalidReplacement) || errors.Is(err, store.ErrBookingConflict) ||
			errors.Is(err, store.ErrMaxBookings) || errors.Is(err, store.ErrMaxUsage) || errors.Is(err, store.ErrStaleManifest) {
			code, message := "409", err.Error()
			return admin.NewReplaceBookingConflict().WithPayload(&models.Error{Code: &code, Message: &message})
		}
		if err != nil {
			code, message := "500", err.Error()
			return admin.NewReplaceBookingInternalServerError().WithPayload(&models.Error{Code: &code, Message: &message})
		}
		return admin.NewReplaceBookingOK().WithPayload(editableBookingToModel(result))
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

		gm := make(map[string]models.Group)

		for k, v := range sm.Groups {
			s := v
			gm[k] = models.Group{
				Description: gog.Ptr(s.Description),
				Policies:    s.Policies,
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
				ConfigURL:        s.ConfigURL,
				Class:            s.Class,
				Description:      gog.Ptr(s.Description),
				Properties:       s.Properties,
				Operations:       storeOperationalProfileToModel(s.Operations),
				StreamOperations: storeStreamOperationsToModel(s.StreamOperations),
				Streams:          s.Streams,
				Tests:            s.Tests,
				TopicStub:        gog.Ptr(s.TopicStub),
			}
		}

		workflows := make(map[string]models.OperationalWorkflow, len(sm.OperationalWorkflows))
		for name, workflow := range sm.OperationalWorkflows {
			description, expected, maximum := workflow.Description, workflow.ExpectedDuration.String(), workflow.MaximumDuration.String()
			workflows[name] = models.OperationalWorkflow{Description: &description, Kind: workflow.Kind, ExpectedDuration: &expected, MaximumDuration: &maximum}
		}
		jobTemplates, pipelineTemplates := storeOperationalTemplatesToModel(sm)
		schedules := make(map[string]models.OperationalSchedule, len(sm.OperationalSchedules))
		for name, schedule := range sm.OperationalSchedules {
			duration, conflict, slot, workflow := schedule.Duration.String(), schedule.Conflict, schedule.Slot, schedule.Workflow
			timezone, at := schedule.Recurrence.Timezone, schedule.Recurrence.Time
			startValue, _ := time.Parse("2006-01-02", schedule.Recurrence.StartDate)
			endValue, _ := time.Parse("2006-01-02", schedule.Recurrence.EndDate)
			startDate, endDate := strfmt.Date(startValue), strfmt.Date(endValue)
			exceptions := make([]strfmt.Date, 0, len(schedule.Recurrence.Exceptions))
			for _, item := range schedule.Recurrence.Exceptions {
				parsed, _ := time.Parse("2006-01-02", item)
				exceptions = append(exceptions, strfmt.Date(parsed))
			}
			schedules[name] = models.OperationalSchedule{Slot: &slot, Workflow: &workflow, Duration: &duration, Conflict: &conflict,
				Recurrence: &models.OperationalRecurrence{Timezone: &timezone, StartDate: &startDate, EndDate: &endDate,
					Weekdays: schedule.Recurrence.Weekdays, Time: &at, Exceptions: exceptions}}
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

			convertRecurrences := func(values []store.WeeklyRecurrence) []*models.WeeklyRecurrence {
				var result []*models.WeeklyRecurrence
				for _, value := range values {
					startValue, _ := time.Parse("2006-01-02", value.StartDate)
					endValue, _ := time.Parse("2006-01-02", value.EndDate)
					startDate, endDate := strfmt.Date(startValue), strfmt.Date(endValue)
					exceptions := make([]strfmt.Date, 0, len(value.Exceptions))
					for _, item := range value.Exceptions {
						parsed, _ := time.Parse("2006-01-02", item)
						exceptions = append(exceptions, strfmt.Date(parsed))
					}
					timezone, startTime, endTime := value.Timezone, value.StartTime, value.EndTime
					result = append(result, &models.WeeklyRecurrence{Timezone: &timezone, StartDate: &startDate, EndDate: &endDate, Weekdays: value.Weekdays, StartTime: &startTime, EndTime: &endTime, Exceptions: exceptions})
				}
				return result
			}
			wm[k] = models.Window{Allowed: aa, Denied: dd, RecurringAllowed: convertRecurrences(s.RecurringAllowed), RecurringDenied: convertRecurrences(s.RecurringDenied)}
		}

		mm := models.Manifest{
			Descriptions:                 dm,
			DisplayGuides:                dgm,
			Groups:                       gm,
			Policies:                     pm,
			OperationalWorkflows:         workflows,
			OperationalSchedules:         schedules,
			OperationalJobTemplates:      jobTemplates,
			OperationalPipelineTemplates: pipelineTemplates,
			Resources:                    rm,
			Slots:                        slm,
			Streams:                      stm,
			Uis:                          uim,
			UISets:                       usm,
			Windows:                      wm,
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

// getResourcesHandler
func getResourcesHandler(config config.ServerConfig) func(admin.GetResourcesParams, interface{}) middleware.Responder {
	return func(params admin.GetResourcesParams, principal interface{}) middleware.Responder {

		_, err := isAdmin(principal)

		if err != nil {
			c := "401"
			m := "no scope booking:admin"
			return admin.NewGetResourcesUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		rs := config.Store.GetResources()

		rm := make(map[string]models.Resource)

		for k, v := range rs {
			s := v
			rm[k] = models.Resource{
				ConfigURL:   s.ConfigURL,
				Description: gog.Ptr(s.Description),
				Streams:     s.Streams,
				Tests:       s.Tests,
				TopicStub:   gog.Ptr(s.TopicStub),
			}
		}

		return admin.NewGetResourcesOK().WithPayload(rm)
	}
}

func getOperationalStatusHandler(config config.ServerConfig) func(admin.GetOperationalStatusParams, interface{}) middleware.Responder {
	return func(params admin.GetOperationalStatusParams, principal interface{}) middleware.Responder {
		if _, err := isAdmin(principal); err != nil {
			c, m := "401", "no scope booking:admin"
			return admin.NewGetOperationalStatusUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}
		if err := config.Store.RefreshManifest(context.Background()); err != nil {
			c, m := "500", err.Error()
			return admin.NewGetOperationalStatusInternalServerError().WithPayload(&models.Error{Code: &c, Message: &m})
		}
		snapshot := config.Store.GetOperationalStatus()
		status, err := convertStoreStatusAdminToModel(snapshot.Status)
		if err != nil {
			c, m := "500", err.Error()
			return admin.NewGetOperationalStatusInternalServerError().WithPayload(&models.Error{Code: &c, Message: &m})
		}
		resources := make(models.ResourceStatuses, len(snapshot.Resources))
		for name, value := range snapshot.Resources {
			available, reason := value.Available, value.Reason
			resources[name] = models.ResourceStatus{Available: &available, Reason: &reason}
		}
		slots := make(models.ResourceStatuses, len(snapshot.Slots))
		for name, value := range snapshot.Slots {
			available, reason := value.Available, value.Reason
			slots[name] = models.ResourceStatus{Available: &available, Reason: &reason}
		}
		version := snapshot.ManifestVersion
		return admin.NewGetOperationalStatusOK().WithPayload(&models.OperationalStatus{
			ManifestVersion: &version, Status: &status, Resources: resources, Slots: slots,
		})
	}
}

func listOperationalOccurrencesHandler(config config.ServerConfig) func(admin.ListOperationalOccurrencesParams, interface{}) middleware.Responder {
	return func(params admin.ListOperationalOccurrencesParams, principal interface{}) middleware.Responder {
		if _, err := isAdmin(principal); err != nil {
			c, m := "401", "no scope booking:admin"
			return admin.NewListOperationalOccurrencesUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}
		repository, ok := config.OperationsRepository.(operational.ScheduleReader)
		if !ok {
			c, m := "500", "operational occurrence storage is not configured"
			return admin.NewListOperationalOccurrencesInternalServerError().WithPayload(&models.Error{Code: &c, Message: &m})
		}
		now := time.Now().UTC()
		if config.Now != nil {
			now = config.Now().UTC()
		}
		from, until := now.Add(-24*time.Hour), now.Add(7*24*time.Hour)
		if params.From != nil {
			from = time.Time(*params.From).UTC()
		}
		if params.Until != nil {
			until = time.Time(*params.Until).UTC()
		}
		if !until.After(from) {
			c, m := "400", "until must be after from"
			return admin.NewListOperationalOccurrencesBadRequest().WithPayload(&models.Error{Code: &c, Message: &m})
		}
		state := ""
		if params.State != nil {
			state = *params.State
		}
		limit := 200
		if params.Limit != nil {
			limit = int(*params.Limit)
		}
		ctx := context.Background()
		if params.HTTPRequest != nil {
			ctx = params.HTTPRequest.Context()
		}
		items, err := repository.ListScheduleOccurrences(ctx, from, until, state, limit)
		if err != nil {
			c, m := "500", err.Error()
			return admin.NewListOperationalOccurrencesInternalServerError().WithPayload(&models.Error{Code: &c, Message: &m})
		}
		result := make([]*models.OperationalOccurrence, 0, len(items))
		for _, item := range items {
			occurred := strfmt.DateTime(item.OccurrenceAt)
			schedule, occurrenceState, slot, resource, workflow, version := item.Schedule, item.State, item.Slot, item.Resource, item.Workflow, item.ManifestVersion
			result = append(result, &models.OperationalOccurrence{Schedule: &schedule, OccurrenceAt: &occurred,
				ManifestVersion: &version, State: &occurrenceState, Slot: &slot, Resource: &resource, Workflow: &workflow,
				BookingName: item.BookingName, JobID: item.JobID, Detail: item.Detail})
		}
		return admin.NewListOperationalOccurrencesOK().WithPayload(result)
	}
}

func queryBookingRecordsHandler(config config.ServerConfig) func(admin.QueryBookingRecordsParams, interface{}) middleware.Responder {
	return func(params admin.QueryBookingRecordsParams, principal interface{}) middleware.Responder {
		if _, err := isAdmin(principal); err != nil {
			return admin.NewQueryBookingRecordsUnauthorized().WithPayload(calendarError("401", "no scope booking:admin"))
		}
		query := store.BookingQuery{}
		if params.Resource != nil {
			query.Resource = *params.Resource
		}
		if params.Slot != nil {
			query.Slot = *params.Slot
		}
		if params.Policy != nil {
			query.Policy = *params.Policy
		}
		if params.User != nil {
			query.User = *params.User
		}
		if params.State != nil {
			query.State = *params.State
		}
		if params.Limit != nil {
			query.Limit = int(*params.Limit)
		}
		if params.From != nil {
			value := time.Time(*params.From)
			query.From = &value
		}
		if params.To != nil {
			value := time.Time(*params.To)
			query.To = &value
		}
		records, err := config.Store.QueryBookings(query)
		if err != nil {
			return admin.NewQueryBookingRecordsBadRequest().WithPayload(calendarError("400", err.Error()))
		}
		payload := make([]*models.AdminBookingRecord, 0, len(records))
		for _, record := range records {
			resource, collection, usage := record.Resource, record.Collection, store.HumaniseDuration(record.ActualUsage)
			payload = append(payload, &models.AdminBookingRecord{Booking: bookingToModel(record.Booking), Resource: &resource, Collection: &collection, ActualUsage: &usage})
		}
		return admin.NewQueryBookingRecordsOK().WithPayload(payload)
	}
}

func getBookingEventsHandler(config config.ServerConfig) func(admin.GetBookingEventsParams, interface{}) middleware.Responder {
	return func(params admin.GetBookingEventsParams, principal interface{}) middleware.Responder {
		if _, err := isAdmin(principal); err != nil {
			return admin.NewGetBookingEventsUnauthorized().WithPayload(calendarError("401", "no scope booking:admin"))
		}
		events, err := config.Store.GetBookingEvents(params.BookingName)
		if err != nil {
			if errors.Is(err, store.ErrPersistentNotFound) {
				return admin.NewGetBookingEventsNotFound().WithPayload(calendarError("404", "booking events not found"))
			}
			return admin.NewGetBookingEventsInternalServerError().WithPayload(calendarError("500", err.Error()))
		}
		payload := make([]*models.BookingEvent, 0, len(events))
		for _, event := range events {
			name, kind, actor, at := event.BookingName, event.Type, event.Actor, strfmt.DateTime(event.OccurredAt)
			payload = append(payload, &models.BookingEvent{BookingName: &name, Type: &kind, Actor: &actor, OccurredAt: &at})
		}
		return admin.NewGetBookingEventsOK().WithPayload(payload)
	}
}

func getUsageSummaryHandler(config config.ServerConfig) func(admin.GetUsageSummaryParams, interface{}) middleware.Responder {
	return func(params admin.GetUsageSummaryParams, principal interface{}) middleware.Responder {
		if _, err := isAdmin(principal); err != nil {
			code, message := "401", "no scope booking:admin"
			return admin.NewGetUsageSummaryUnauthorized().WithPayload(&models.Error{Code: &code, Message: &message})
		}
		query := store.UsageQuery{}
		if params.Resource != nil {
			query.Resource = *params.Resource
		}
		if params.Slot != nil {
			query.Slot = *params.Slot
		}
		if params.Policy != nil {
			query.Policy = *params.Policy
		}
		if params.User != nil {
			query.User = *params.User
		}
		if params.From != nil {
			value := time.Time(*params.From)
			query.From = &value
		}
		if params.To != nil {
			value := time.Time(*params.To)
			query.To = &value
		}
		summary, err := config.Store.GetFilteredUsageSummaryPersistent(query)
		if err != nil {
			code, message := "500", "could not read usage summary"
			return admin.NewGetUsageSummaryInternalServerError().WithPayload(&models.Error{Code: &code, Message: &message})
		}
		actual := summary.ActualUsage.String()
		preparation, cleanup, scheduled := summary.PreparationUsage.String(), summary.CleanupUsage.String(), summary.ScheduledUsage.String()
		return admin.NewGetUsageSummaryOK().WithPayload(&models.UsageSummary{
			ActualUsage: &actual, PreparationUsage: &preparation, CleanupUsage: &cleanup, ScheduledUsage: &scheduled, OperationalJobs: &summary.OperationalJobs,
			StartedBookings: &summary.StartedBookings, CompletedBookings: &summary.CompletedBookings,
		})
	}
}

func makeMaintenanceBookingHandler(config config.ServerConfig) func(admin.MakeMaintenanceBookingParams, interface{}) middleware.Responder {
	return func(params admin.MakeMaintenanceBookingParams, principal interface{}) middleware.Responder {
		claims, err := isMaintenanceOrAdmin(principal)
		if err != nil {
			code, message := "401", "no scope booking:maintenance"
			return admin.NewMakeMaintenanceBookingUnauthorized().WithPayload(&models.Error{Code: &code, Message: &message})
		}
		booking, err := config.Store.MakeMaintenanceBooking(params.SlotName, claims.Subject, interval.Interval{
			Start: time.Time(params.From), End: time.Time(params.To),
		})
		if err != nil {
			code, message := "409", err.Error()
			return admin.NewMakeMaintenanceBookingConflict().WithPayload(&models.Error{Code: &code, Message: &message})
		}
		return admin.NewMakeMaintenanceBookingOK().WithPayload(bookingToModel(booking))
	}
}

func makeResourceMaintenanceBookingHandler(config config.ServerConfig) func(admin.MakeResourceMaintenanceBookingParams, interface{}) middleware.Responder {
	return func(params admin.MakeResourceMaintenanceBookingParams, principal interface{}) middleware.Responder {
		claims, err := isMaintenanceOrAdmin(principal)
		if err != nil {
			return admin.NewMakeResourceMaintenanceBookingUnauthorized().WithPayload(calendarError("401", "no scope booking:maintenance"))
		}
		booking, err := config.Store.MakeMaintenanceBookingForResource(params.ResourceName, claims.Subject, interval.Interval{Start: time.Time(params.From), End: time.Time(params.To)})
		if err != nil {
			if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "no slot") {
				return admin.NewMakeResourceMaintenanceBookingNotFound().WithPayload(calendarError("404", err.Error()))
			}
			return admin.NewMakeResourceMaintenanceBookingConflict().WithPayload(calendarError("409", err.Error()))
		}
		return admin.NewMakeResourceMaintenanceBookingOK().WithPayload(bookingToModel(booking))
	}
}

func overrideCancelBookingHandler(config config.ServerConfig) func(admin.OverrideCancelBookingParams, interface{}) middleware.Responder {
	return func(params admin.OverrideCancelBookingParams, principal interface{}) middleware.Responder {
		claims, err := isOverrideOrAdmin(principal)
		if err != nil {
			code, message := "401", "no scope booking:booking-override"
			return admin.NewOverrideCancelBookingUnauthorized().WithPayload(&models.Error{Code: &code, Message: &message})
		}
		if strings.TrimSpace(params.Reason) == "" {
			code, message := "409", "a cancellation reason is required"
			return admin.NewOverrideCancelBookingConflict().WithPayload(&models.Error{Code: &code, Message: &message})
		}
		booking, err := config.Store.GetBooking(params.BookingName)
		if err != nil {
			code, message := "404", "booking not found"
			return admin.NewOverrideCancelBookingNotFound().WithPayload(&models.Error{Code: &code, Message: &message})
		}
		actor := claims.Subject
		if actor == "" {
			actor = "booking-override"
		}
		if err := config.Store.CancelBooking(booking, actor+": "+params.Reason); err != nil {
			code, message := "409", err.Error()
			return admin.NewOverrideCancelBookingConflict().WithPayload(&models.Error{Code: &code, Message: &message})
		}
		return admin.NewOverrideCancelBookingNoContent()
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
			gs := []string{}
			obs := []string{}
			um := make(map[string]string)

			for _, bv := range v.Bookings {
				bs = append(bs, bv)
			}

			for _, obv := range v.OldBookings {
				obs = append(obs, obv)
			}

			// ignore bool in map, has no meaning
			for _, gv := range v.Groups {
				gs = append(gs, gv)
			}

			// store format is map[string]*time.Duration
			for uk, uv := range v.Usage {
				um[uk] = uv
			}

			m := models.User{

				Bookings:    bs,
				OldBookings: obs,
				Groups:      gs,
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

// getResourceIsAvailableHandlerFunc
func getResourceIsAvailableHandler(config config.ServerConfig) func(admin.GetResourceIsAvailableParams, interface{}) middleware.Responder {
	return func(params admin.GetResourceIsAvailableParams, principal interface{}) middleware.Responder {

		_, err := isAdmin(principal)

		if err != nil {
			c := "401"
			m := "no scope booking:admin"
			return admin.NewGetResourceIsAvailableUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		avail, reason, err := config.Store.GetResourceIsAvailable(params.ResourceName)

		if err != nil {
			c := "404"
			m := err.Error()
			return admin.NewGetResourceIsAvailableNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		s := models.ResourceStatus{
			Available: &avail,
			Reason:    &reason,
		}
		return admin.NewGetResourceIsAvailableOK().WithPayload(&s)
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

		s := models.ResourceStatus{
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

		if err := config.Store.SetMaintenance(params.Lock, params.Msg); err != nil {
			c, m := "500", err.Error()
			return admin.NewSetLockInternalServerError().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		s, err := convertStoreStatusAdminToModel(config.Store.GetStoreStatusAdmin())

		if err != nil {
			log.Error("could not convert StoreStatusAdmin to model format")
		}

		return admin.NewSetLockOK().WithPayload(&s)
	}
}

// setResourceIsAvailableHandlerFunc
func setResourceIsAvailableHandler(config config.ServerConfig) func(admin.SetResourceIsAvailableParams, interface{}) middleware.Responder {
	return func(params admin.SetResourceIsAvailableParams, principal interface{}) middleware.Responder {

		_, err := isAdmin(principal)

		if err != nil {
			c := "401"
			m := "no scope booking:admin"
			return admin.NewSetResourceIsAvailableUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		err = config.Store.SetResourceIsAvailable(params.ResourceName, params.Available, params.Reason)

		if err != nil {
			c := "404"
			m := err.Error()
			return admin.NewSetResourceIsAvailableNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		return admin.NewSetResourceIsAvailableNoContent()
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
