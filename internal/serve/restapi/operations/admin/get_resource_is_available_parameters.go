// Code generated by go-swagger; DO NOT EDIT.

package admin

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
)

// NewGetResourceIsAvailableParams creates a new GetResourceIsAvailableParams object
//
// There are no default values defined in the spec.
func NewGetResourceIsAvailableParams() GetResourceIsAvailableParams {

	return GetResourceIsAvailableParams{}
}

// GetResourceIsAvailableParams contains all the bound params for the get resource is available operation
// typically these are obtained from a http.Request
//
// swagger:parameters GetResourceIsAvailable
type GetResourceIsAvailableParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*
	  Required: true
	  In: path
	*/
	ResourceName string
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewGetResourceIsAvailableParams() beforehand.
func (o *GetResourceIsAvailableParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	rResourceName, rhkResourceName, _ := route.Params.GetOK("resource_name")
	if err := o.bindResourceName(rResourceName, rhkResourceName, route.Formats); err != nil {
		res = append(res, err)
	}
	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// bindResourceName binds and validates parameter ResourceName from path.
func (o *GetResourceIsAvailableParams) bindResourceName(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true
	// Parameter is provided by construction from the route
	o.ResourceName = raw

	return nil
}