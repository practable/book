// Code generated by go-swagger; DO NOT EDIT.

package admin

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// ExportOldBookingsHandlerFunc turns a function with the right signature into a export old bookings handler
type ExportOldBookingsHandlerFunc func(ExportOldBookingsParams, interface{}) middleware.Responder

// Handle executing the request and returning a response
func (fn ExportOldBookingsHandlerFunc) Handle(params ExportOldBookingsParams, principal interface{}) middleware.Responder {
	return fn(params, principal)
}

// ExportOldBookingsHandler interface for that can handle valid export old bookings params
type ExportOldBookingsHandler interface {
	Handle(ExportOldBookingsParams, interface{}) middleware.Responder
}

// NewExportOldBookings creates a new http.Handler for the export old bookings operation
func NewExportOldBookings(ctx *middleware.Context, handler ExportOldBookingsHandler) *ExportOldBookings {
	return &ExportOldBookings{Context: ctx, Handler: handler}
}

/* ExportOldBookings swagger:route GET /admin/oldbookings admin exportOldBookings

Export a copy of all old bookings

Exports a copy of the old bookings, with sufficient information to allow editing and replacement. If successful produces JSON-formatted bookings list.

*/
type ExportOldBookings struct {
	Context *middleware.Context
	Handler ExportOldBookingsHandler
}

func (o *ExportOldBookings) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewExportOldBookingsParams()
	uprinc, aCtx, err := o.Context.Authorize(r, route)
	if err != nil {
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}
	if aCtx != nil {
		*r = *aCtx
	}
	var principal interface{}
	if uprinc != nil {
		principal = uprinc.(interface{}) // this is really a interface{}, I promise
	}

	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params, principal) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}
