// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// UISet set of User Interfaces
//
// swagger:model UISet
type UISet struct {

	// u is
	UIs []string `json:"UIs"`
}

// Validate validates this UI set
func (m *UISet) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this UI set based on context it is used
func (m *UISet) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *UISet) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *UISet) UnmarshalBinary(b []byte) error {
	var res UISet
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
