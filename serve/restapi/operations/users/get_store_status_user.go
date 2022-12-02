// Code generated by go-swagger; DO NOT EDIT.

package users

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// GetStoreStatusUserHandlerFunc turns a function with the right signature into a get store status user handler
type GetStoreStatusUserHandlerFunc func(GetStoreStatusUserParams, interface{}) middleware.Responder

// Handle executing the request and returning a response
func (fn GetStoreStatusUserHandlerFunc) Handle(params GetStoreStatusUserParams, principal interface{}) middleware.Responder {
	return fn(params, principal)
}

// GetStoreStatusUserHandler interface for that can handle valid get store status user params
type GetStoreStatusUserHandler interface {
	Handle(GetStoreStatusUserParams, interface{}) middleware.Responder
}

// NewGetStoreStatusUser creates a new http.Handler for the get store status user operation
func NewGetStoreStatusUser(ctx *middleware.Context, handler GetStoreStatusUserHandler) *GetStoreStatusUser {
	return &GetStoreStatusUser{Context: ctx, Handler: handler}
}

/* GetStoreStatusUser swagger:route GET /users/status users status getStoreStatusUser

Get current store status

Gets the current store status from a user perspective (e.g. is it locked? what is the reason?)

*/
type GetStoreStatusUser struct {
	Context *middleware.Context
	Handler GetStoreStatusUserHandler
}

func (o *GetStoreStatusUser) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewGetStoreStatusUserParams()
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
