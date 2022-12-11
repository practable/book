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

// Manifest manifest
//
// Represents resources that can be booked
//
// swagger:model Manifest
type Manifest struct {

	// descriptions
	// Required: true
	Descriptions map[string]Description `json:"descriptions"`

	// policies
	// Required: true
	Policies map[string]Policy `json:"policies"`

	// resources
	// Required: true
	Resources map[string]Resource `json:"resources"`

	// slots
	// Required: true
	Slots map[string]Slot `json:"slots"`

	// streams
	// Required: true
	Streams map[string]Stream `json:"streams"`

	// ui sets
	// Required: true
	UISets map[string]UISet `json:"ui_sets"`

	// uis
	// Required: true
	Uis map[string]UI `json:"uis"`

	// windows
	// Required: true
	Windows map[string]Window `json:"windows"`
}

// Validate validates this manifest
func (m *Manifest) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateDescriptions(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validatePolicies(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateResources(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateSlots(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateStreams(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateUISets(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateUis(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateWindows(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Manifest) validateDescriptions(formats strfmt.Registry) error {

	if err := validate.Required("descriptions", "body", m.Descriptions); err != nil {
		return err
	}

	for k := range m.Descriptions {

		if err := validate.Required("descriptions"+"."+k, "body", m.Descriptions[k]); err != nil {
			return err
		}
		if val, ok := m.Descriptions[k]; ok {
			if err := val.Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("descriptions" + "." + k)
				} else if ce, ok := err.(*errors.CompositeError); ok {
					return ce.ValidateName("descriptions" + "." + k)
				}
				return err
			}
		}

	}

	return nil
}

func (m *Manifest) validatePolicies(formats strfmt.Registry) error {

	if err := validate.Required("policies", "body", m.Policies); err != nil {
		return err
	}

	for k := range m.Policies {

		if err := validate.Required("policies"+"."+k, "body", m.Policies[k]); err != nil {
			return err
		}
		if val, ok := m.Policies[k]; ok {
			if err := val.Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("policies" + "." + k)
				} else if ce, ok := err.(*errors.CompositeError); ok {
					return ce.ValidateName("policies" + "." + k)
				}
				return err
			}
		}

	}

	return nil
}

func (m *Manifest) validateResources(formats strfmt.Registry) error {

	if err := validate.Required("resources", "body", m.Resources); err != nil {
		return err
	}

	for k := range m.Resources {

		if err := validate.Required("resources"+"."+k, "body", m.Resources[k]); err != nil {
			return err
		}
		if val, ok := m.Resources[k]; ok {
			if err := val.Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("resources" + "." + k)
				} else if ce, ok := err.(*errors.CompositeError); ok {
					return ce.ValidateName("resources" + "." + k)
				}
				return err
			}
		}

	}

	return nil
}

func (m *Manifest) validateSlots(formats strfmt.Registry) error {

	if err := validate.Required("slots", "body", m.Slots); err != nil {
		return err
	}

	for k := range m.Slots {

		if err := validate.Required("slots"+"."+k, "body", m.Slots[k]); err != nil {
			return err
		}
		if val, ok := m.Slots[k]; ok {
			if err := val.Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("slots" + "." + k)
				} else if ce, ok := err.(*errors.CompositeError); ok {
					return ce.ValidateName("slots" + "." + k)
				}
				return err
			}
		}

	}

	return nil
}

func (m *Manifest) validateStreams(formats strfmt.Registry) error {

	if err := validate.Required("streams", "body", m.Streams); err != nil {
		return err
	}

	for k := range m.Streams {

		if err := validate.Required("streams"+"."+k, "body", m.Streams[k]); err != nil {
			return err
		}
		if val, ok := m.Streams[k]; ok {
			if err := val.Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("streams" + "." + k)
				} else if ce, ok := err.(*errors.CompositeError); ok {
					return ce.ValidateName("streams" + "." + k)
				}
				return err
			}
		}

	}

	return nil
}

func (m *Manifest) validateUISets(formats strfmt.Registry) error {

	if err := validate.Required("ui_sets", "body", m.UISets); err != nil {
		return err
	}

	for k := range m.UISets {

		if err := validate.Required("ui_sets"+"."+k, "body", m.UISets[k]); err != nil {
			return err
		}

		if err := m.UISets[k].Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("ui_sets" + "." + k)
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("ui_sets" + "." + k)
			}
			return err
		}

	}

	return nil
}

func (m *Manifest) validateUis(formats strfmt.Registry) error {

	if err := validate.Required("uis", "body", m.Uis); err != nil {
		return err
	}

	for k := range m.Uis {

		if err := validate.Required("uis"+"."+k, "body", m.Uis[k]); err != nil {
			return err
		}
		if val, ok := m.Uis[k]; ok {
			if err := val.Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("uis" + "." + k)
				} else if ce, ok := err.(*errors.CompositeError); ok {
					return ce.ValidateName("uis" + "." + k)
				}
				return err
			}
		}

	}

	return nil
}

func (m *Manifest) validateWindows(formats strfmt.Registry) error {

	if err := validate.Required("windows", "body", m.Windows); err != nil {
		return err
	}

	for k := range m.Windows {

		if err := validate.Required("windows"+"."+k, "body", m.Windows[k]); err != nil {
			return err
		}
		if val, ok := m.Windows[k]; ok {
			if err := val.Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("windows" + "." + k)
				} else if ce, ok := err.(*errors.CompositeError); ok {
					return ce.ValidateName("windows" + "." + k)
				}
				return err
			}
		}

	}

	return nil
}

// ContextValidate validate this manifest based on the context it is used
func (m *Manifest) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := m.contextValidateDescriptions(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := m.contextValidatePolicies(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := m.contextValidateResources(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := m.contextValidateSlots(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := m.contextValidateStreams(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := m.contextValidateUISets(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := m.contextValidateUis(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := m.contextValidateWindows(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Manifest) contextValidateDescriptions(ctx context.Context, formats strfmt.Registry) error {

	if err := validate.Required("descriptions", "body", m.Descriptions); err != nil {
		return err
	}

	for k := range m.Descriptions {

		if val, ok := m.Descriptions[k]; ok {
			if err := val.ContextValidate(ctx, formats); err != nil {
				return err
			}
		}

	}

	return nil
}

func (m *Manifest) contextValidatePolicies(ctx context.Context, formats strfmt.Registry) error {

	if err := validate.Required("policies", "body", m.Policies); err != nil {
		return err
	}

	for k := range m.Policies {

		if val, ok := m.Policies[k]; ok {
			if err := val.ContextValidate(ctx, formats); err != nil {
				return err
			}
		}

	}

	return nil
}

func (m *Manifest) contextValidateResources(ctx context.Context, formats strfmt.Registry) error {

	if err := validate.Required("resources", "body", m.Resources); err != nil {
		return err
	}

	for k := range m.Resources {

		if val, ok := m.Resources[k]; ok {
			if err := val.ContextValidate(ctx, formats); err != nil {
				return err
			}
		}

	}

	return nil
}

func (m *Manifest) contextValidateSlots(ctx context.Context, formats strfmt.Registry) error {

	if err := validate.Required("slots", "body", m.Slots); err != nil {
		return err
	}

	for k := range m.Slots {

		if val, ok := m.Slots[k]; ok {
			if err := val.ContextValidate(ctx, formats); err != nil {
				return err
			}
		}

	}

	return nil
}

func (m *Manifest) contextValidateStreams(ctx context.Context, formats strfmt.Registry) error {

	if err := validate.Required("streams", "body", m.Streams); err != nil {
		return err
	}

	for k := range m.Streams {

		if val, ok := m.Streams[k]; ok {
			if err := val.ContextValidate(ctx, formats); err != nil {
				return err
			}
		}

	}

	return nil
}

func (m *Manifest) contextValidateUISets(ctx context.Context, formats strfmt.Registry) error {

	if err := validate.Required("ui_sets", "body", m.UISets); err != nil {
		return err
	}

	for k := range m.UISets {

		if err := m.UISets[k].ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("ui_sets" + "." + k)
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("ui_sets" + "." + k)
			}
			return err
		}

	}

	return nil
}

func (m *Manifest) contextValidateUis(ctx context.Context, formats strfmt.Registry) error {

	if err := validate.Required("uis", "body", m.Uis); err != nil {
		return err
	}

	for k := range m.Uis {

		if val, ok := m.Uis[k]; ok {
			if err := val.ContextValidate(ctx, formats); err != nil {
				return err
			}
		}

	}

	return nil
}

func (m *Manifest) contextValidateWindows(ctx context.Context, formats strfmt.Registry) error {

	if err := validate.Required("windows", "body", m.Windows); err != nil {
		return err
	}

	for k := range m.Windows {

		if val, ok := m.Windows[k]; ok {
			if err := val.ContextValidate(ctx, formats); err != nil {
				return err
			}
		}

	}

	return nil
}

// MarshalBinary interface implementation
func (m *Manifest) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Manifest) UnmarshalBinary(b []byte) error {
	var res Manifest
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
