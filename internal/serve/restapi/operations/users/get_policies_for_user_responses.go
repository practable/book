// Code generated by go-swagger; DO NOT EDIT.

package users

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/timdrysdale/interval/internal/serve/models"
)

// GetPoliciesForUserOKCode is the HTTP code returned for type GetPoliciesForUserOK
const GetPoliciesForUserOKCode int = 200

/*GetPoliciesForUserOK OK

swagger:response getPoliciesForUserOK
*/
type GetPoliciesForUserOK struct {

	/*
	  In: Body
	*/
	Payload models.Policies `json:"body,omitempty"`
}

// NewGetPoliciesForUserOK creates GetPoliciesForUserOK with default headers values
func NewGetPoliciesForUserOK() *GetPoliciesForUserOK {

	return &GetPoliciesForUserOK{}
}

// WithPayload adds the payload to the get policies for user o k response
func (o *GetPoliciesForUserOK) WithPayload(payload models.Policies) *GetPoliciesForUserOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get policies for user o k response
func (o *GetPoliciesForUserOK) SetPayload(payload models.Policies) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetPoliciesForUserOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	payload := o.Payload
	if payload == nil {
		// return empty array
		payload = models.Policies{}
	}

	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}
}

// GetPoliciesForUserUnauthorizedCode is the HTTP code returned for type GetPoliciesForUserUnauthorized
const GetPoliciesForUserUnauthorizedCode int = 401

/*GetPoliciesForUserUnauthorized Unauthorized

swagger:response getPoliciesForUserUnauthorized
*/
type GetPoliciesForUserUnauthorized struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetPoliciesForUserUnauthorized creates GetPoliciesForUserUnauthorized with default headers values
func NewGetPoliciesForUserUnauthorized() *GetPoliciesForUserUnauthorized {

	return &GetPoliciesForUserUnauthorized{}
}

// WithPayload adds the payload to the get policies for user unauthorized response
func (o *GetPoliciesForUserUnauthorized) WithPayload(payload *models.Error) *GetPoliciesForUserUnauthorized {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get policies for user unauthorized response
func (o *GetPoliciesForUserUnauthorized) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetPoliciesForUserUnauthorized) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(401)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetPoliciesForUserNotFoundCode is the HTTP code returned for type GetPoliciesForUserNotFound
const GetPoliciesForUserNotFoundCode int = 404

/*GetPoliciesForUserNotFound The specified resource was not found

swagger:response getPoliciesForUserNotFound
*/
type GetPoliciesForUserNotFound struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetPoliciesForUserNotFound creates GetPoliciesForUserNotFound with default headers values
func NewGetPoliciesForUserNotFound() *GetPoliciesForUserNotFound {

	return &GetPoliciesForUserNotFound{}
}

// WithPayload adds the payload to the get policies for user not found response
func (o *GetPoliciesForUserNotFound) WithPayload(payload *models.Error) *GetPoliciesForUserNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get policies for user not found response
func (o *GetPoliciesForUserNotFound) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetPoliciesForUserNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetPoliciesForUserInternalServerErrorCode is the HTTP code returned for type GetPoliciesForUserInternalServerError
const GetPoliciesForUserInternalServerErrorCode int = 500

/*GetPoliciesForUserInternalServerError Internal Error

swagger:response getPoliciesForUserInternalServerError
*/
type GetPoliciesForUserInternalServerError struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetPoliciesForUserInternalServerError creates GetPoliciesForUserInternalServerError with default headers values
func NewGetPoliciesForUserInternalServerError() *GetPoliciesForUserInternalServerError {

	return &GetPoliciesForUserInternalServerError{}
}

// WithPayload adds the payload to the get policies for user internal server error response
func (o *GetPoliciesForUserInternalServerError) WithPayload(payload *models.Error) *GetPoliciesForUserInternalServerError {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get policies for user internal server error response
func (o *GetPoliciesForUserInternalServerError) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetPoliciesForUserInternalServerError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(500)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}