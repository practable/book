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

// AccessToken access token
//
// intended use is for users to access the API, and is tied to their user_name.
//
// swagger:model AccessToken
type AccessToken struct {

	// Audience
	// Required: true
	Aud *string `json:"aud"`

	// Expires At
	// Required: true
	Exp *float64 `json:"exp"`

	// Issued At
	Iat float64 `json:"iat,omitempty"`

	// Not before
	// Required: true
	Nbf *float64 `json:"nbf"`

	// List of scopes
	// Required: true
	Scopes []string `json:"scopes"`

	// Subject
	// Required: true
	Sub *string `json:"sub"`

	// token
	// Required: true
	Token *string `json:"token"`
}

// Validate validates this access token
func (m *AccessToken) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateAud(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateExp(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateNbf(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateScopes(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateSub(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateToken(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *AccessToken) validateAud(formats strfmt.Registry) error {

	if err := validate.Required("aud", "body", m.Aud); err != nil {
		return err
	}

	return nil
}

func (m *AccessToken) validateExp(formats strfmt.Registry) error {

	if err := validate.Required("exp", "body", m.Exp); err != nil {
		return err
	}

	return nil
}

func (m *AccessToken) validateNbf(formats strfmt.Registry) error {

	if err := validate.Required("nbf", "body", m.Nbf); err != nil {
		return err
	}

	return nil
}

func (m *AccessToken) validateScopes(formats strfmt.Registry) error {

	if err := validate.Required("scopes", "body", m.Scopes); err != nil {
		return err
	}

	return nil
}

func (m *AccessToken) validateSub(formats strfmt.Registry) error {

	if err := validate.Required("sub", "body", m.Sub); err != nil {
		return err
	}

	return nil
}

func (m *AccessToken) validateToken(formats strfmt.Registry) error {

	if err := validate.Required("token", "body", m.Token); err != nil {
		return err
	}

	return nil
}

// ContextValidate validates this access token based on context it is used
func (m *AccessToken) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *AccessToken) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *AccessToken) UnmarshalBinary(b []byte) error {
	var res AccessToken
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
