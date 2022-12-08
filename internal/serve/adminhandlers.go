package serve

import (
	"encoding/json"

	"github.com/go-openapi/runtime/middleware"
	log "github.com/sirupsen/logrus"
	"github.com/timdrysdale/interval/internal/config"
	"github.com/timdrysdale/interval/internal/serve/models"
	"github.com/timdrysdale/interval/internal/serve/restapi/operations/admin"
	"github.com/timdrysdale/interval/internal/store"
	"gopkg.in/yaml.v2"
)

// replaceManifestHandler
func replaceManifestHandler(config config.ServerConfig) func(admin.ReplaceManifestParams, interface{}) middleware.Responder {
	return func(params admin.ReplaceManifestParams, principal interface{}) middleware.Responder {

		_, err := isAdmin(principal)

		if err != nil {
			c := "401"
			m := err.Error()
			return admin.NewReplaceManifestUnauthorized().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		if params.Manifest == nil {
			c := "404"
			m := "no manifest in body"
			return admin.NewReplaceManifestNotFound().WithPayload(&models.Error{Code: &c, Message: &m})
		}

		sm, err := convertManifestToStore(*(params.Manifest))
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

func convertStoreStatusAdminToModel(s store.StoreStatusAdmin) (models.StoreStatusAdmin, error) {
	var m models.StoreStatusAdmin

	y, err := json.Marshal(s)

	if err != nil {
		return m, err
	}

	err = json.Unmarshal(y, &m)

	return m, err

}

// convertManifestToStore
func convertManifestToStore(m models.Manifest) (store.Manifest, error) {

	// We don't do manifest replacement often, so using yaml as an intermediate format
	// is not going to be inefficient overall yet reduces maintenance

	var s store.Manifest

	y, err := yaml.Marshal(m)
	if err != nil {
		return s, err
	}

	err = yaml.Unmarshal(y, &s)

	return s, err

	/*
		d := make(map[string]store.Description)
		p := make(map[string]store.Policy)
		r := make(map[string]store.Resource)
		sl := make(map[string]store.Slot)
		st := make(map[string]store.Stream)
		u := make(map[string]store.UI)
		us := make(map[string]store.UISet)
		w := make(map[string]store.Window)



		return store.Manifest{
			Descriptions: d,
			Policies:     p,
			Resources:    r,
			Slots:        sl,
			Streams:      st,
			UIs:          u,
			UISets:       us,
			Windows:      w,
		}, nil
	*/
}
