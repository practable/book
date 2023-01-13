// Code generated by go-swagger; DO NOT EDIT.

package users

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// AddPolicyForUserHandlerFunc turns a function with the right signature into a add policy for user handler
type AddPolicyForUserHandlerFunc func(AddPolicyForUserParams, interface{}) middleware.Responder

// Handle executing the request and returning a response
func (fn AddPolicyForUserHandlerFunc) Handle(params AddPolicyForUserParams, principal interface{}) middleware.Responder {
	return fn(params, principal)
}

// AddPolicyForUserHandler interface for that can handle valid add policy for user params
type AddPolicyForUserHandler interface {
	Handle(AddPolicyForUserParams, interface{}) middleware.Responder
}

// NewAddPolicyForUser creates a new http.Handler for the add policy for user operation
func NewAddPolicyForUser(ctx *middleware.Context, handler AddPolicyForUserHandler) *AddPolicyForUser {
	return &AddPolicyForUser{Context: ctx, Handler: handler}
}

/* AddPolicyForUser swagger:route POST /users/{user_name}/policies/{policy_name} users addPolicyForUser

Add policy to user account

Add policy to the list of policies with which this user is allowed to make bookings

*/
type AddPolicyForUser struct {
	Context *middleware.Context
	Handler AddPolicyForUserHandler
}

func (o *AddPolicyForUser) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewAddPolicyForUserParams()
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
