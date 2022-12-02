// Code generated by go-swagger; DO NOT EDIT.

package policies

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// GetPolicyHandlerFunc turns a function with the right signature into a get policy handler
type GetPolicyHandlerFunc func(GetPolicyParams, interface{}) middleware.Responder

// Handle executing the request and returning a response
func (fn GetPolicyHandlerFunc) Handle(params GetPolicyParams, principal interface{}) middleware.Responder {
	return fn(params, principal)
}

// GetPolicyHandler interface for that can handle valid get policy params
type GetPolicyHandler interface {
	Handle(GetPolicyParams, interface{}) middleware.Responder
}

// NewGetPolicy creates a new http.Handler for the get policy operation
func NewGetPolicy(ctx *middleware.Context, handler GetPolicyHandler) *GetPolicy {
	return &GetPolicy{Context: ctx, Handler: handler}
}

/* GetPolicy swagger:route GET /policies/{policy_name} policies getPolicy

Get policy

Get policy

*/
type GetPolicy struct {
	Context *middleware.Context
	Handler GetPolicyHandler
}

func (o *GetPolicy) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewGetPolicyParams()
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
