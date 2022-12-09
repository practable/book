// Code generated by go-swagger; DO NOT EDIT.

package users

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// NewGetAvailabilityParams creates a new GetAvailabilityParams object
//
// There are no default values defined in the spec.
func NewGetAvailabilityParams() GetAvailabilityParams {

	return GetAvailabilityParams{}
}

// GetAvailabilityParams contains all the bound params for the get availability operation
// typically these are obtained from a http.Request
//
// swagger:parameters GetAvailability
type GetAvailabilityParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*
	  In: query
	*/
	Limit *int64
	/*
	  In: query
	*/
	Offset *int64
	/*
	  Required: true
	  In: path
	*/
	PolicyName string
	/*
	  Required: true
	  In: path
	*/
	SlotName string
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewGetAvailabilityParams() beforehand.
func (o *GetAvailabilityParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	qs := runtime.Values(r.URL.Query())

	qLimit, qhkLimit, _ := qs.GetOK("limit")
	if err := o.bindLimit(qLimit, qhkLimit, route.Formats); err != nil {
		res = append(res, err)
	}

	qOffset, qhkOffset, _ := qs.GetOK("offset")
	if err := o.bindOffset(qOffset, qhkOffset, route.Formats); err != nil {
		res = append(res, err)
	}

	rPolicyName, rhkPolicyName, _ := route.Params.GetOK("policy_name")
	if err := o.bindPolicyName(rPolicyName, rhkPolicyName, route.Formats); err != nil {
		res = append(res, err)
	}

	rSlotName, rhkSlotName, _ := route.Params.GetOK("slot_name")
	if err := o.bindSlotName(rSlotName, rhkSlotName, route.Formats); err != nil {
		res = append(res, err)
	}
	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// bindLimit binds and validates parameter Limit from query.
func (o *GetAvailabilityParams) bindLimit(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: false
	// AllowEmptyValue: false

	if raw == "" { // empty values pass all other validations
		return nil
	}

	value, err := swag.ConvertInt64(raw)
	if err != nil {
		return errors.InvalidType("limit", "query", "int64", raw)
	}
	o.Limit = &value

	return nil
}

// bindOffset binds and validates parameter Offset from query.
func (o *GetAvailabilityParams) bindOffset(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: false
	// AllowEmptyValue: false

	if raw == "" { // empty values pass all other validations
		return nil
	}

	value, err := swag.ConvertInt64(raw)
	if err != nil {
		return errors.InvalidType("offset", "query", "int64", raw)
	}
	o.Offset = &value

	return nil
}

// bindPolicyName binds and validates parameter PolicyName from path.
func (o *GetAvailabilityParams) bindPolicyName(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true
	// Parameter is provided by construction from the route
	o.PolicyName = raw

	return nil
}

// bindSlotName binds and validates parameter SlotName from path.
func (o *GetAvailabilityParams) bindSlotName(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true
	// Parameter is provided by construction from the route
	o.SlotName = raw

	return nil
}
