// Code generated by go-swagger; DO NOT EDIT.

package users

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
)

// NewGetDescriptionParams creates a new GetDescriptionParams object
//
// There are no default values defined in the spec.
func NewGetDescriptionParams() GetDescriptionParams {

	return GetDescriptionParams{}
}

// GetDescriptionParams contains all the bound params for the get description operation
// typically these are obtained from a http.Request
//
// swagger:parameters GetDescription
type GetDescriptionParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*
	  Required: true
	  In: path
	*/
	DescriptionName string
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewGetDescriptionParams() beforehand.
func (o *GetDescriptionParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	rDescriptionName, rhkDescriptionName, _ := route.Params.GetOK("description_name")
	if err := o.bindDescriptionName(rDescriptionName, rhkDescriptionName, route.Formats); err != nil {
		res = append(res, err)
	}
	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// bindDescriptionName binds and validates parameter DescriptionName from path.
func (o *GetDescriptionParams) bindDescriptionName(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true
	// Parameter is provided by construction from the route
	o.DescriptionName = raw

	return nil
}
