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

// Resource resource
//
// swagger:model Resource
type Resource struct {

	// config url
	ConfigURL string `json:"config_url,omitempty"`

	// description
	// Required: true
	Description *string `json:"description"`

	// streams
	// Required: true
	Streams []string `json:"streams"`

	// topic stub
	// Required: true
	TopicStub *string `json:"topic_stub"`
}

// Validate validates this resource
func (m *Resource) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateDescription(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateStreams(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateTopicStub(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Resource) validateDescription(formats strfmt.Registry) error {

	if err := validate.Required("description", "body", m.Description); err != nil {
		return err
	}

	return nil
}

func (m *Resource) validateStreams(formats strfmt.Registry) error {

	if err := validate.Required("streams", "body", m.Streams); err != nil {
		return err
	}

	return nil
}

func (m *Resource) validateTopicStub(formats strfmt.Registry) error {

	if err := validate.Required("topic_stub", "body", m.TopicStub); err != nil {
		return err
	}

	return nil
}

// ContextValidate validates this resource based on context it is used
func (m *Resource) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *Resource) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Resource) UnmarshalBinary(b []byte) error {
	var res Resource
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
