package serve

import (
	"encoding/json"
	"strings"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/icza/gog"
	log "github.com/sirupsen/logrus"
	"github.com/timdrysdale/interval/internal/config"
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

		if params.Manifest == "" {
			c := "404"
			m := "no manifest in body"
			return admin.NewCheckManifestNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		sm, err := convertManifestToStore(params.Manifest)
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

// convertBookingsToStore converts from YAML string to internal type
func convertBookingsToStore(m string) (map[string]store.Booking, error) {

	var s map[string]store.Booking

	err := yaml.Unmarshal([]byte(m), &s)

	return s, err
}

// convertManifestToStore converts from YAML string to internal type
func convertManifestToStore(m string) (store.Manifest, error) {

	var s store.Manifest

	err := yaml.Unmarshal([]byte(m), &s)

	return s, err
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

		//b, err := json.Marshal(m)

		//if err != nil {
		//	c := "500"
		//	m := err.Error()
		//	return admin.NewExportBookingsInternalServerError().WithPayload(&models.Error{Code: &c, Message: &m})
		//}

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

		m := config.Store.ExportManifest()

		b, err := json.Marshal(m)

		if err != nil {
			c := "500"
			m := err.Error()
			return admin.NewExportManifestInternalServerError().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		return admin.NewExportManifestOK().WithPayload(string(b))
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

		m := config.Store.ExportOldBookings()

		b, err := json.Marshal(m)

		if err != nil {
			c := "500"
			m := err.Error()
			return admin.NewExportOldBookingsInternalServerError().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		return admin.NewExportOldBookingsOK().WithPayload(string(b))
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

		m := config.Store.ExportUsers()

		b, err := json.Marshal(m)

		if err != nil {
			c := "500"
			m := err.Error()
			return admin.NewExportUsersInternalServerError().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		return admin.NewExportUsersOK().WithPayload(string(b))
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

		if params.Bookings == "" {
			c := "404"
			m := "no manifest in body"
			return admin.NewReplaceBookingsNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		sm, err := convertBookingsToStore(params.Bookings)
		if err != nil {
			c := "500"
			m := err.Error()
			return admin.NewReplaceBookingsInternalServerError().WithPayload(&models.Error{Code: &c, Message: &m})
		}

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

		if params.Manifest == "" {
			c := "404"
			m := "no manifest in body"
			return admin.NewReplaceManifestNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		sm, err := convertManifestToStore(params.Manifest)
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

		if params.Bookings == "" {
			c := "404"
			m := "no manifest in body"
			return admin.NewReplaceOldBookingsNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
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
