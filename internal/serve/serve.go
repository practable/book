// Package booking provides an API for booking experiments
package serve

import (
	"context"
	"flag"

	"github.com/go-openapi/loads"
	"github.com/practable/book/internal/config"
	"github.com/practable/book/internal/serve/restapi"
	"github.com/practable/book/internal/serve/restapi/operations"
	"github.com/practable/book/internal/serve/restapi/operations/admin"
	"github.com/practable/book/internal/serve/restapi/operations/users"
	log "github.com/sirupsen/logrus"
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
func API(ctx context.Context, config config.ServerConfig, cancelOthers func()) {
	defer func() {
		log.Trace("serve.API stopped")
	}()

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
	api.AdminGetStoreStatusAdminHandler = admin.GetStoreStatusAdminHandlerFunc(getStoreStatusAdminHandler(config))
	api.AdminGetSlotIsAvailableHandler = admin.GetSlotIsAvailableHandlerFunc(getSlotIsAvailableHandler(config))
	api.AdminExportBookingsHandler = admin.ExportBookingsHandlerFunc(exportBookingsHandler(config))
	api.AdminExportManifestHandler = admin.ExportManifestHandlerFunc(exportManifestHandler(config))
	api.AdminExportOldBookingsHandler = admin.ExportOldBookingsHandlerFunc(exportOldBookingsHandler(config))
	api.AdminExportUsersHandler = admin.ExportUsersHandlerFunc(exportUsersHandler(config))
	api.AdminReplaceBookingsHandler = admin.ReplaceBookingsHandlerFunc(replaceBookingsHandler(config))
	api.AdminReplaceManifestHandler = admin.ReplaceManifestHandlerFunc(replaceManifestHandler(config))
	api.AdminReplaceOldBookingsHandler = admin.ReplaceOldBookingsHandlerFunc(replaceOldBookingsHandler(config))
	api.AdminSetLockHandler = admin.SetLockHandlerFunc(setLockHandler(config))
	api.AdminSetSlotIsAvailableHandler = admin.SetSlotIsAvailableHandlerFunc(setSlotIsAvailableHandler(config))

	// *** USERS *** //
	api.UsersAddGroupForUserHandler = users.AddGroupForUserHandlerFunc(addGroupForUserHandler(config))
	api.UsersCancelBookingHandler = users.CancelBookingHandlerFunc(cancelBookingHandler(config))
	api.UsersGetAccessTokenHandler = users.GetAccessTokenHandlerFunc(getAccessTokenHandler(config))
	api.UsersGetActivityHandler = users.GetActivityHandlerFunc(getActivityHandler(config))
	api.UsersGetAvailabilityHandler = users.GetAvailabilityHandlerFunc(getAvailabilityHandler(config))
	api.UsersGetBookingsForUserHandler = users.GetBookingsForUserHandlerFunc(getBookingsForUserHandler(config))
	api.UsersGetDescriptionHandler = users.GetDescriptionHandlerFunc(getDescriptionHandler(config))
	api.UsersGetGroupHandler = users.GetGroupHandlerFunc(getGroupHandler(config))
	api.UsersGetOldBookingsForUserHandler = users.GetOldBookingsForUserHandlerFunc(getOldBookingsForUserHandler(config))
	api.UsersGetGroupsForUserHandler = users.GetGroupsForUserHandlerFunc(getGroupsForUserHandler(config))
	api.UsersGetPolicyHandler = users.GetPolicyHandlerFunc(getPolicyHandler(config))
	api.UsersGetPolicyStatusForUserHandler = users.GetPolicyStatusForUserHandlerFunc(getPolicyStatusForUserHandler(config))
	api.UsersGetStoreStatusUserHandler = users.GetStoreStatusUserHandlerFunc(getStoreStatusUserHandler(config))
	api.UsersMakeBookingHandler = users.MakeBookingHandlerFunc(makeBookingHandler(config))
	api.UsersUniqueNameHandler = users.UniqueNameHandlerFunc(uniqueNameHandler(config))

	c := make(chan struct{})

	go func() {
		defer func() {
			log.Trace("serve.API cancel checker goro stopped")
		}()
		log.Trace("serve API awaiting context cancellation")
		select {
		case <-ctx.Done():
			log.Trace("serve API context cancelled")
			if err := server.Shutdown(); err != nil {
				log.Fatalln(err)
			}
		case <-c:
			// clean up this goro when we stop
			// for another reason
			return
		}

	}()

	//serve API
	if err := server.Serve(); err != nil {
		log.Fatalln(err)
	}
	log.Trace("serve API stopped without error")

	close(c) // clean up the goro checking for cancellation of our own context

	cancelOthers()

	log.Trace("serve API called cancelOthers")

}
