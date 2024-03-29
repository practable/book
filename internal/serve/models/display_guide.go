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

// DisplayGuide display guide
//
// swagger:model DisplayGuide
type DisplayGuide struct {

	// book ahead
	// Required: true
	BookAhead *string `json:"book_ahead"`

	// duration
	// Required: true
	Duration *string `json:"duration"`

	// what to display in the tab heading for these slots
	// Required: true
	Label *string `json:"label"`

	// max slots
	// Required: true
	MaxSlots *int64 `json:"max_slots"`
}

// Validate validates this display guide
func (m *DisplayGuide) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateBookAhead(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateDuration(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateLabel(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateMaxSlots(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *DisplayGuide) validateBookAhead(formats strfmt.Registry) error {

	if err := validate.Required("book_ahead", "body", m.BookAhead); err != nil {
		return err
	}

	return nil
}

func (m *DisplayGuide) validateDuration(formats strfmt.Registry) error {

	if err := validate.Required("duration", "body", m.Duration); err != nil {
		return err
	}

	return nil
}

func (m *DisplayGuide) validateLabel(formats strfmt.Registry) error {

	if err := validate.Required("label", "body", m.Label); err != nil {
		return err
	}

	return nil
}

func (m *DisplayGuide) validateMaxSlots(formats strfmt.Registry) error {

	if err := validate.Required("max_slots", "body", m.MaxSlots); err != nil {
		return err
	}

	return nil
}

// ContextValidate validates this display guide based on context it is used
func (m *DisplayGuide) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *DisplayGuide) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *DisplayGuide) UnmarshalBinary(b []byte) error {
	var res DisplayGuide
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
