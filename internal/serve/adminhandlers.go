package serve

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/timdrysdale/interval/internal/config"
)

// replaceManifestHandler
func replaceManifestHandler(config config.ServerConfig) func(params admin.ReplaceManifestParams, principal interface{}) middleware.Responder {

	return func(params admin.ReplaceManifestParams) middleware.Responder {

	}

}
