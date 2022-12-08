package serve

import (
	"encoding/json"
	"strings"

	"github.com/go-openapi/runtime/middleware"
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
			m := err.Error()
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

// getStoreStatusAdminHandler
func getStoreStatusAdminHandler(config config.ServerConfig) func(admin.GetStoreStatusAdminParams, interface{}) middleware.Responder {
	return func(params admin.GetStoreStatusAdminParams, principal interface{}) middleware.Responder {

		_, err := isAdmin(principal)

		if err != nil {
			c := "401"
			m := err.Error()
			return admin.NewGetStoreStatusAdminUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		s, err := convertStoreStatusAdminToModel(config.Store.GetStoreStatusAdmin())

		if err != nil {
			log.Error("could not convert StoreStatusAdmin to model format")
		}

		return admin.NewGetStoreStatusAdminOK().WithPayload(&s)
	}
}

// exportBookingsHandler
func exportBookingsHandler(config config.ServerConfig) func(admin.ExportBookingsParams, interface{}) middleware.Responder {
	return func(params admin.ExportBookingsParams, principal interface{}) middleware.Responder {

		_, err := isAdmin(principal)

		if err != nil {
			c := "401"
			m := err.Error()
			return admin.NewExportBookingsUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		m := config.Store.ExportBookings()

		b, err := json.Marshal(m)

		if err != nil {
			c := "500"
			m := err.Error()
			return admin.NewExportBookingsInternalServerError().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		return admin.NewExportBookingsOK().WithPayload(string(b))
	}
}

// exportManifestHandler
func exportManifestHandler(config config.ServerConfig) func(admin.ExportManifestParams, interface{}) middleware.Responder {
	return func(params admin.ExportManifestParams, principal interface{}) middleware.Responder {

		_, err := isAdmin(principal)

		if err != nil {
			c := "401"
			m := err.Error()
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
			m := err.Error()
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
			m := err.Error()
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

// replaceBookingsHandler
func replaceBookingsHandler(config config.ServerConfig) func(admin.ReplaceBookingsParams, interface{}) middleware.Responder {
	return func(params admin.ReplaceBookingsParams, principal interface{}) middleware.Responder {

		_, err := isAdmin(principal)

		if err != nil {
			c := "401"
			m := err.Error()
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
			m := err.Error()
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
			m := err.Error()
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
