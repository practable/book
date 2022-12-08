// Code generated by go-swagger; DO NOT EDIT.

package admin

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// GetStoreStatusAdminHandlerFunc turns a function with the right signature into a get store status admin handler
type GetStoreStatusAdminHandlerFunc func(GetStoreStatusAdminParams, interface{}) middleware.Responder

// Handle executing the request and returning a response
func (fn GetStoreStatusAdminHandlerFunc) Handle(params GetStoreStatusAdminParams, principal interface{}) middleware.Responder {
	return fn(params, principal)
}

// GetStoreStatusAdminHandler interface for that can handle valid get store status admin params
type GetStoreStatusAdminHandler interface {
	Handle(GetStoreStatusAdminParams, interface{}) middleware.Responder
}

// NewGetStoreStatusAdmin creates a new http.Handler for the get store status admin operation
func NewGetStoreStatusAdmin(ctx *middleware.Context, handler GetStoreStatusAdminHandler) *GetStoreStatusAdmin {
	return &GetStoreStatusAdmin{Context: ctx, Handler: handler}
}

/* GetStoreStatusAdmin swagger:route GET /admin/status admin status getStoreStatusAdmin

Get current store status

Gets a count of the number of elements in the store, e.g. Bookings, Descriptions etc to facilitate a necessary but not sufficient check that replace manifest and replace bookings have produced the correct results.

*/
type GetStoreStatusAdmin struct {
	Context *middleware.Context
	Handler GetStoreStatusAdminHandler
}

func (o *GetStoreStatusAdmin) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewGetStoreStatusAdminParams()
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