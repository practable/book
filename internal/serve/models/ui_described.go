// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// UIDescribed User Interface with description including
//
// swagger:model UIDescribed
type UIDescribed struct {

	// description
	// Required: true
	Description *Description `json:"description"`

	// list of names of required streams
	// Example: ["data","video"]
	StreamsRequired []string `json:"streams_required"`

	// template for the URL for the user interface
	// Example: https://static.practable.io/ui/penduino-basic.html?video={{video}}\u0026data={{data}}
	// Required: true
	URL *string `json:"url"`
}

// Validate validates this UI described
func (m *UIDescribed) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateDescription(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateURL(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *UIDescribed) validateDescription(formats strfmt.Registry) error {

	if err := validate.Required("description", "body", m.Description); err != nil {
		return err
	}

	if m.Description != nil {
		if err := m.Description.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("description")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("description")
			}
			return err
		}
	}

	return nil
}

func (m *UIDescribed) validateURL(formats strfmt.Registry) error {

	if err := validate.Required("url", "body", m.URL); err != nil {
		return err
	}

	return nil
}

// ContextValidate validate this UI described based on the context it is used
func (m *UIDescribed) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := m.contextValidateDescription(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *UIDescribed) contextValidateDescription(ctx context.Context, formats strfmt.Registry) error {

	if m.Description != nil {
		if err := m.Description.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("description")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("description")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (m *UIDescribed) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *UIDescribed) UnmarshalBinary(b []byte) error {
	var res UIDescribed
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
