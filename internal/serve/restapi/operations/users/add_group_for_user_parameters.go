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

// NewAddGroupForUserParams creates a new AddGroupForUserParams object
//
// There are no default values defined in the spec.
func NewAddGroupForUserParams() AddGroupForUserParams {

	return AddGroupForUserParams{}
}

// AddGroupForUserParams contains all the bound params for the add group for user operation
// typically these are obtained from a http.Request
//
// swagger:parameters AddGroupForUser
type AddGroupForUserParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*
	  Required: true
	  In: path
	*/
	GroupName string
	/*
	  Required: true
	  In: path
	*/
	UserName string
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewAddGroupForUserParams() beforehand.
func (o *AddGroupForUserParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	rGroupName, rhkGroupName, _ := route.Params.GetOK("group_name")
	if err := o.bindGroupName(rGroupName, rhkGroupName, route.Formats); err != nil {
		res = append(res, err)
	}

	rUserName, rhkUserName, _ := route.Params.GetOK("user_name")
	if err := o.bindUserName(rUserName, rhkUserName, route.Formats); err != nil {
		res = append(res, err)
	}
	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// bindGroupName binds and validates parameter GroupName from path.
func (o *AddGroupForUserParams) bindGroupName(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true
	// Parameter is provided by construction from the route
	o.GroupName = raw

	return nil
}

// bindUserName binds and validates parameter UserName from path.
func (o *AddGroupForUserParams) bindUserName(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true
	// Parameter is provided by construction from the route
	o.UserName = raw

	return nil
}
