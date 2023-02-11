// Code generated by go-swagger; DO NOT EDIT.

package users

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// AddGroupForUserHandlerFunc turns a function with the right signature into a add group for user handler
type AddGroupForUserHandlerFunc func(AddGroupForUserParams, interface{}) middleware.Responder

// Handle executing the request and returning a response
func (fn AddGroupForUserHandlerFunc) Handle(params AddGroupForUserParams, principal interface{}) middleware.Responder {
	return fn(params, principal)
}

// AddGroupForUserHandler interface for that can handle valid add group for user params
type AddGroupForUserHandler interface {
	Handle(AddGroupForUserParams, interface{}) middleware.Responder
}

// NewAddGroupForUser creates a new http.Handler for the add group for user operation
func NewAddGroupForUser(ctx *middleware.Context, handler AddGroupForUserHandler) *AddGroupForUser {
	return &AddGroupForUser{Context: ctx, Handler: handler}
}

/* AddGroupForUser swagger:route POST /users/{user_name}/groups/{group_name} users addGroupForUser

Add group to user account

Add group to the list of groups with which this user is allowed to make bookings

*/
type AddGroupForUser struct {
	Context *middleware.Context
	Handler AddGroupForUserHandler
}

func (o *AddGroupForUser) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewAddGroupForUserParams()
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