// Package booking provides an API for booking experiments
package serve

import (
	"context"
	"flag"

	"github.com/go-openapi/loads"
	log "github.com/sirupsen/logrus"
	"github.com/timdrysdale/interval/internal/config"
	"github.com/timdrysdale/interval/internal/serve/restapi"
	"github.com/timdrysdale/interval/internal/serve/restapi/operations"
	"github.com/timdrysdale/interval/internal/serve/restapi/operations/admin"
	"github.com/timdrysdale/interval/internal/serve/restapi/operations/users"
)

// API starts the API
// Inputs
// @closed - channel will be closed when server shutsdown
// @wg - waitgroup, we must wg.Done() when we are shutdown
// @port - where to listen locally
// @host - external FQDN of the host (for checking against tokens) e.g. https://relay-access.practable.io
// @target - FQDN of the relay instance e.g. wss://relay.practable.io
// @secret- HMAC shared secret which incoming tokens will be signed with
// @cs - pointer to the CodeStore this API shares with the shellbar websocket relay
// @options - for future backwards compatibility (no options currently available)
func API(ctx context.Context, config config.ServerConfig) {

	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		log.Fatalln(err)
	}

	//create new service API
	api := operations.NewServeAPI(swaggerSpec)
	server := restapi.NewServer(api)

	//parse flags
	flag.Parse()

	// set the port this service will run on
	server.Port = config.Port

	// set the Authorizer
	api.BearerAuth = validateHeader(config.StoreSecret, config.Host)

	// set the Handlers

	// *** ADMIN *** //
	api.AdminCheckManifestHandler = admin.CheckManifestHandlerFunc(checkManifestHandler(config))
	api.AdminExportBookingsHandler = admin.ExportBookingsHandlerFunc(exportBookingsHandler(config))
	api.AdminExportManifestHandler = admin.ExportManifestHandlerFunc(exportManifestHandler(config))
	api.AdminReplaceBookingsHandler = admin.ReplaceBookingsHandlerFunc(replaceBookingsHandler(config))
	api.AdminReplaceManifestHandler = admin.ReplaceManifestHandlerFunc(replaceManifestHandler(config))

	// *** USERS *** //
	api.UsersGetAccessTokenHandler = users.GetAccessTokenHandlerFunc(getAccessTokenHandler(config))

	go func() {
		<-ctx.Done()
		if err := server.Shutdown(); err != nil {
			log.Fatalln(err)
		}

	}()

	//serve API
	if err := server.Serve(); err != nil {
		log.Fatalln(err)
	}

}
