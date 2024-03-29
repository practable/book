// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"strconv"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// Activity activity
//
// An activity represents connection details to instances of pre-agreed resource types such as video, data streams and user interfaces.
//
// swagger:model Activity
type Activity struct {

	// URL at which to GET the configuration object
	// Example: https://assets.practable.io/config/experiments/pvna/pvna01-0.0.json
	Config string `json:"config,omitempty"`

	// description
	// Required: true
	Description *Description `json:"description"`

	// Expires At
	// Required: true
	Exp *float64 `json:"exp"`

	// Expires At
	// Required: true
	Nbf *float64 `json:"nbf"`

	// A list of streams
	// Required: true
	Streams []*ActivityStream `json:"streams"`

	// User interfaces
	// Required: true
	Uis []*UIDescribed `json:"uis"`
}

// Validate validates this activity
func (m *Activity) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateDescription(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateExp(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateNbf(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateStreams(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateUis(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Activity) validateDescription(formats strfmt.Registry) error {

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

func (m *Activity) validateExp(formats strfmt.Registry) error {

	if err := validate.Required("exp", "body", m.Exp); err != nil {
		return err
	}

	return nil
}

func (m *Activity) validateNbf(formats strfmt.Registry) error {

	if err := validate.Required("nbf", "body", m.Nbf); err != nil {
		return err
	}

	return nil
}

func (m *Activity) validateStreams(formats strfmt.Registry) error {

	if err := validate.Required("streams", "body", m.Streams); err != nil {
		return err
	}

	for i := 0; i < len(m.Streams); i++ {
		if swag.IsZero(m.Streams[i]) { // not required
			continue
		}

		if m.Streams[i] != nil {
			if err := m.Streams[i].Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("streams" + "." + strconv.Itoa(i))
				} else if ce, ok := err.(*errors.CompositeError); ok {
					return ce.ValidateName("streams" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

func (m *Activity) validateUis(formats strfmt.Registry) error {

	if err := validate.Required("uis", "body", m.Uis); err != nil {
		return err
	}

	for i := 0; i < len(m.Uis); i++ {
		if swag.IsZero(m.Uis[i]) { // not required
			continue
		}

		if m.Uis[i] != nil {
			if err := m.Uis[i].Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("uis" + "." + strconv.Itoa(i))
				} else if ce, ok := err.(*errors.CompositeError); ok {
					return ce.ValidateName("uis" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

// ContextValidate validate this activity based on the context it is used
func (m *Activity) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := m.contextValidateDescription(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := m.contextValidateStreams(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := m.contextValidateUis(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Activity) contextValidateDescription(ctx context.Context, formats strfmt.Registry) error {

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

func (m *Activity) contextValidateStreams(ctx context.Context, formats strfmt.Registry) error {

	for i := 0; i < len(m.Streams); i++ {

		if m.Streams[i] != nil {
			if err := m.Streams[i].ContextValidate(ctx, formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("streams" + "." + strconv.Itoa(i))
				} else if ce, ok := err.(*errors.CompositeError); ok {
					return ce.ValidateName("streams" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

func (m *Activity) contextValidateUis(ctx context.Context, formats strfmt.Registry) error {

	for i := 0; i < len(m.Uis); i++ {

		if m.Uis[i] != nil {
			if err := m.Uis[i].ContextValidate(ctx, formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("uis" + "." + strconv.Itoa(i))
				} else if ce, ok := err.(*errors.CompositeError); ok {
					return ce.ValidateName("uis" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

// MarshalBinary interface implementation
func (m *Activity) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Activity) UnmarshalBinary(b []byte) error {
	var res Activity
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
