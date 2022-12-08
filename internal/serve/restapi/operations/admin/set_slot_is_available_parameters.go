// Code generated by go-swagger; DO NOT EDIT.

package admin

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// NewSetSlotIsAvailableParams creates a new SetSlotIsAvailableParams object
//
// There are no default values defined in the spec.
func NewSetSlotIsAvailableParams() SetSlotIsAvailableParams {

	return SetSlotIsAvailableParams{}
}

// SetSlotIsAvailableParams contains all the bound params for the set slot is available operation
// typically these are obtained from a http.Request
//
// swagger:parameters SetSlotIsAvailable
type SetSlotIsAvailableParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*
	  Required: true
	  In: query
	*/
	Available bool
	/*
	  Required: true
	  In: query
	*/
	Reason string
	/*
	  Required: true
	  In: path
	*/
	SlotName string
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewSetSlotIsAvailableParams() beforehand.
func (o *SetSlotIsAvailableParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	qs := runtime.Values(r.URL.Query())

	qAvailable, qhkAvailable, _ := qs.GetOK("available")
	if err := o.bindAvailable(qAvailable, qhkAvailable, route.Formats); err != nil {
		res = append(res, err)
	}

	qReason, qhkReason, _ := qs.GetOK("reason")
	if err := o.bindReason(qReason, qhkReason, route.Formats); err != nil {
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

// bindAvailable binds and validates parameter Available from query.
func (o *SetSlotIsAvailableParams) bindAvailable(rawData []string, hasKey bool, formats strfmt.Registry) error {
	if !hasKey {
		return errors.Required("available", "query", rawData)
	}
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true
	// AllowEmptyValue: false

	if err := validate.RequiredString("available", "query", raw); err != nil {
		return err
	}

	value, err := swag.ConvertBool(raw)
	if err != nil {
		return errors.InvalidType("available", "query", "bool", raw)
	}
	o.Available = value

	return nil
}

// bindReason binds and validates parameter Reason from query.
func (o *SetSlotIsAvailableParams) bindReason(rawData []string, hasKey bool, formats strfmt.Registry) error {
	if !hasKey {
		return errors.Required("reason", "query", rawData)
	}
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true
	// AllowEmptyValue: false

	if err := validate.RequiredString("reason", "query", raw); err != nil {
		return err
	}
	o.Reason = raw

	return nil
}

// bindSlotName binds and validates parameter SlotName from path.
func (o *SetSlotIsAvailableParams) bindSlotName(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true
	// Parameter is provided by construction from the route
	o.SlotName = raw

	return nil
}