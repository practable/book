// This file is safe to edit. Once it exists it will not be overwritten

package restapi

import (
	"crypto/tls"
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"

	"github.com/timdrysdale/interval/internal/serve/restapi/operations"
	"github.com/timdrysdale/interval/internal/serve/restapi/operations/admin"
	"github.com/timdrysdale/interval/internal/serve/restapi/operations/descriptions"
	"github.com/timdrysdale/interval/internal/serve/restapi/operations/policies"
	"github.com/timdrysdale/interval/internal/serve/restapi/operations/users"
)

//go:generate swagger generate server --target ../../serve --name Serve --spec ../../../api/booking.yml --principal interface{} --exclude-main

func configureFlags(api *operations.ServeAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.ServeAPI) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	// api.Logger = log.Printf

	api.UseSwaggerUI()
	// To continue using redoc as your UI, uncomment the following line
	// api.UseRedoc()

	api.JSONConsumer = runtime.JSONConsumer()
	api.TxtConsumer = runtime.TextConsumer()

	api.JSONProducer = runtime.JSONProducer()

	// Applies when the "Authorization" header is set
	if api.BearerAuth == nil {
		api.BearerAuth = func(token string) (interface{}, error) {
			return nil, errors.NotImplemented("api key auth (Bearer) Authorization from header param [Authorization] has not yet been implemented")
		}
	}

	// Set your custom authorizer if needed. Default one is security.Authorized()
	// Expected interface runtime.Authorizer
	//
	// Example:
	// api.APIAuthorizer = security.Authorized()

	if api.UsersAddPolicyForUserHandler == nil {
		api.UsersAddPolicyForUserHandler = users.AddPolicyForUserHandlerFunc(func(params users.AddPolicyForUserParams, principal interface{}) middleware.Responder {
			return middleware.NotImplemented("operation users.AddPolicyForUser has not yet been implemented")
		})
	}
	if api.UsersCancelBookingHandler == nil {
		api.UsersCancelBookingHandler = users.CancelBookingHandlerFunc(func(params users.CancelBookingParams, principal interface{}) middleware.Responder {
			return middleware.NotImplemented("operation users.CancelBooking has not yet been implemented")
		})
	}
	if api.AdminCheckManifestHandler == nil {
		api.AdminCheckManifestHandler = admin.CheckManifestHandlerFunc(func(params admin.CheckManifestParams, principal interface{}) middleware.Responder {
			return middleware.NotImplemented("operation admin.CheckManifest has not yet been implemented")
		})
	}
	if api.AdminExportBookingsHandler == nil {
		api.AdminExportBookingsHandler = admin.ExportBookingsHandlerFunc(func(params admin.ExportBookingsParams, principal interface{}) middleware.Responder {
			return middleware.NotImplemented("operation admin.ExportBookings has not yet been implemented")
		})
	}
	if api.AdminExportManifestHandler == nil {
		api.AdminExportManifestHandler = admin.ExportManifestHandlerFunc(func(params admin.ExportManifestParams, principal interface{}) middleware.Responder {
			return middleware.NotImplemented("operation admin.ExportManifest has not yet been implemented")
		})
	}
	if api.AdminExportOldBookingsHandler == nil {
		api.AdminExportOldBookingsHandler = admin.ExportOldBookingsHandlerFunc(func(params admin.ExportOldBookingsParams, principal interface{}) middleware.Responder {
			return middleware.NotImplemented("operation admin.ExportOldBookings has not yet been implemented")
		})
	}
	if api.AdminExportUsersHandler == nil {
		api.AdminExportUsersHandler = admin.ExportUsersHandlerFunc(func(params admin.ExportUsersParams, principal interface{}) middleware.Responder {
			return middleware.NotImplemented("operation admin.ExportUsers has not yet been implemented")
		})
	}
	if api.UsersGetAccessTokenHandler == nil {
		api.UsersGetAccessTokenHandler = users.GetAccessTokenHandlerFunc(func(params users.GetAccessTokenParams) middleware.Responder {
			return middleware.NotImplemented("operation users.GetAccessToken has not yet been implemented")
		})
	}
	if api.UsersGetActivityHandler == nil {
		api.UsersGetActivityHandler = users.GetActivityHandlerFunc(func(params users.GetActivityParams, principal interface{}) middleware.Responder {
			return middleware.NotImplemented("operation users.GetActivity has not yet been implemented")
		})
	}
	if api.UsersGetAvailabilityHandler == nil {
		api.UsersGetAvailabilityHandler = users.GetAvailabilityHandlerFunc(func(params users.GetAvailabilityParams, principal interface{}) middleware.Responder {
			return middleware.NotImplemented("operation users.GetAvailability has not yet been implemented")
		})
	}
	if api.UsersGetBookingsForUserHandler == nil {
		api.UsersGetBookingsForUserHandler = users.GetBookingsForUserHandlerFunc(func(params users.GetBookingsForUserParams, principal interface{}) middleware.Responder {
			return middleware.NotImplemented("operation users.GetBookingsForUser has not yet been implemented")
		})
	}
	if api.DescriptionsGetDescriptionHandler == nil {
		api.DescriptionsGetDescriptionHandler = descriptions.GetDescriptionHandlerFunc(func(params descriptions.GetDescriptionParams, principal interface{}) middleware.Responder {
			return middleware.NotImplemented("operation descriptions.GetDescription has not yet been implemented")
		})
	}
	if api.UsersGetOldBookingsForUserHandler == nil {
		api.UsersGetOldBookingsForUserHandler = users.GetOldBookingsForUserHandlerFunc(func(params users.GetOldBookingsForUserParams, principal interface{}) middleware.Responder {
			return middleware.NotImplemented("operation users.GetOldBookingsForUser has not yet been implemented")
		})
	}
	if api.UsersGetPoliciesForUserHandler == nil {
		api.UsersGetPoliciesForUserHandler = users.GetPoliciesForUserHandlerFunc(func(params users.GetPoliciesForUserParams, principal interface{}) middleware.Responder {
			return middleware.NotImplemented("operation users.GetPoliciesForUser has not yet been implemented")
		})
	}
	if api.PoliciesGetPolicyHandler == nil {
		api.PoliciesGetPolicyHandler = policies.GetPolicyHandlerFunc(func(params policies.GetPolicyParams, principal interface{}) middleware.Responder {
			return middleware.NotImplemented("operation policies.GetPolicy has not yet been implemented")
		})
	}
	if api.UsersGetPolicyStatusForUserHandler == nil {
		api.UsersGetPolicyStatusForUserHandler = users.GetPolicyStatusForUserHandlerFunc(func(params users.GetPolicyStatusForUserParams, principal interface{}) middleware.Responder {
			return middleware.NotImplemented("operation users.GetPolicyStatusForUser has not yet been implemented")
		})
	}
	if api.AdminGetSlotIsAvailableHandler == nil {
		api.AdminGetSlotIsAvailableHandler = admin.GetSlotIsAvailableHandlerFunc(func(params admin.GetSlotIsAvailableParams, principal interface{}) middleware.Responder {
			return middleware.NotImplemented("operation admin.GetSlotIsAvailable has not yet been implemented")
		})
	}
	if api.UsersMakeBookingHandler == nil {
		api.UsersMakeBookingHandler = users.MakeBookingHandlerFunc(func(params users.MakeBookingParams, principal interface{}) middleware.Responder {
			return middleware.NotImplemented("operation users.MakeBooking has not yet been implemented")
		})
	}
	if api.AdminReplaceBookingsHandler == nil {
		api.AdminReplaceBookingsHandler = admin.ReplaceBookingsHandlerFunc(func(params admin.ReplaceBookingsParams, principal interface{}) middleware.Responder {
			return middleware.NotImplemented("operation admin.ReplaceBookings has not yet been implemented")
		})
	}
	if api.AdminReplaceManifestHandler == nil {
		api.AdminReplaceManifestHandler = admin.ReplaceManifestHandlerFunc(func(params admin.ReplaceManifestParams, principal interface{}) middleware.Responder {
			return middleware.NotImplemented("operation admin.ReplaceManifest has not yet been implemented")
		})
	}
	if api.AdminReplaceOldBookingsHandler == nil {
		api.AdminReplaceOldBookingsHandler = admin.ReplaceOldBookingsHandlerFunc(func(params admin.ReplaceOldBookingsParams, principal interface{}) middleware.Responder {
			return middleware.NotImplemented("operation admin.ReplaceOldBookings has not yet been implemented")
		})
	}
	if api.AdminSetSlotIsAvailableHandler == nil {
		api.AdminSetSlotIsAvailableHandler = admin.SetSlotIsAvailableHandlerFunc(func(params admin.SetSlotIsAvailableParams, principal interface{}) middleware.Responder {
			return middleware.NotImplemented("operation admin.SetSlotIsAvailable has not yet been implemented")
		})
	}
	if api.AdminGetStoreStatusAdminHandler == nil {
		api.AdminGetStoreStatusAdminHandler = admin.GetStoreStatusAdminHandlerFunc(func(params admin.GetStoreStatusAdminParams, principal interface{}) middleware.Responder {
			return middleware.NotImplemented("operation admin.GetStoreStatusAdmin has not yet been implemented")
		})
	}
	if api.UsersGetStoreStatusUserHandler == nil {
		api.UsersGetStoreStatusUserHandler = users.GetStoreStatusUserHandlerFunc(func(params users.GetStoreStatusUserParams, principal interface{}) middleware.Responder {
			return middleware.NotImplemented("operation users.GetStoreStatusUser has not yet been implemented")
		})
	}
	if api.AdminSetLockHandler == nil {
		api.AdminSetLockHandler = admin.SetLockHandlerFunc(func(params admin.SetLockParams, principal interface{}) middleware.Responder {
			return middleware.NotImplemented("operation admin.SetLock has not yet been implemented")
		})
	}

	api.PreServerShutdown = func() {}

	api.ServerShutdown = func() {}

	return setupGlobalMiddleware(api.Serve(setupMiddlewares))
}

// The TLS configuration before HTTPS server starts.
func configureTLS(tlsConfig *tls.Config) {
	// Make all necessary changes to the TLS configuration here.
}

// As soon as server is initialized but not run yet, this function will be called.
// If you need to modify a config, store server instance to stop it individually later, this is the place.
// This function can be called multiple times, depending on the number of serving schemes.
// scheme value will be set accordingly: "http", "https" or "unix".
func configureServer(s *http.Server, scheme, addr string) {
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation.
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics.
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return handler
}
