// Code generated by go-swagger; DO NOT EDIT.

package admin

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/timdrysdale/interval/internal/serve/models"
)

// SetSlotIsAvailableNoContentCode is the HTTP code returned for type SetSlotIsAvailableNoContent
const SetSlotIsAvailableNoContentCode int = 204

/*
SetSlotIsAvailableNoContent OK

swagger:response setSlotIsAvailableNoContent
*/
type SetSlotIsAvailableNoContent struct {
}

// NewSetSlotIsAvailableNoContent creates SetSlotIsAvailableNoContent with default headers values
func NewSetSlotIsAvailableNoContent() *SetSlotIsAvailableNoContent {

	return &SetSlotIsAvailableNoContent{}
}

// WriteResponse to the client
func (o *SetSlotIsAvailableNoContent) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(204)
}

// SetSlotIsAvailableUnauthorizedCode is the HTTP code returned for type SetSlotIsAvailableUnauthorized
const SetSlotIsAvailableUnauthorizedCode int = 401

/*
SetSlotIsAvailableUnauthorized Unauthorized

swagger:response setSlotIsAvailableUnauthorized
*/
type SetSlotIsAvailableUnauthorized struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewSetSlotIsAvailableUnauthorized creates SetSlotIsAvailableUnauthorized with default headers values
func NewSetSlotIsAvailableUnauthorized() *SetSlotIsAvailableUnauthorized {

	return &SetSlotIsAvailableUnauthorized{}
}

// WithPayload adds the payload to the set slot is available unauthorized response
func (o *SetSlotIsAvailableUnauthorized) WithPayload(payload *models.Error) *SetSlotIsAvailableUnauthorized {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the set slot is available unauthorized response
func (o *SetSlotIsAvailableUnauthorized) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *SetSlotIsAvailableUnauthorized) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(401)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// SetSlotIsAvailableNotFoundCode is the HTTP code returned for type SetSlotIsAvailableNotFound
const SetSlotIsAvailableNotFoundCode int = 404

/*
SetSlotIsAvailableNotFound The specified resource was not found

swagger:response setSlotIsAvailableNotFound
*/
type SetSlotIsAvailableNotFound struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewSetSlotIsAvailableNotFound creates SetSlotIsAvailableNotFound with default headers values
func NewSetSlotIsAvailableNotFound() *SetSlotIsAvailableNotFound {

	return &SetSlotIsAvailableNotFound{}
}

// WithPayload adds the payload to the set slot is available not found response
func (o *SetSlotIsAvailableNotFound) WithPayload(payload *models.Error) *SetSlotIsAvailableNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the set slot is available not found response
func (o *SetSlotIsAvailableNotFound) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *SetSlotIsAvailableNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// SetSlotIsAvailableInternalServerErrorCode is the HTTP code returned for type SetSlotIsAvailableInternalServerError
const SetSlotIsAvailableInternalServerErrorCode int = 500

/*
SetSlotIsAvailableInternalServerError Internal Error

swagger:response setSlotIsAvailableInternalServerError
*/
type SetSlotIsAvailableInternalServerError struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewSetSlotIsAvailableInternalServerError creates SetSlotIsAvailableInternalServerError with default headers values
func NewSetSlotIsAvailableInternalServerError() *SetSlotIsAvailableInternalServerError {

	return &SetSlotIsAvailableInternalServerError{}
}

// WithPayload adds the payload to the set slot is available internal server error response
func (o *SetSlotIsAvailableInternalServerError) WithPayload(payload *models.Error) *SetSlotIsAvailableInternalServerError {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the set slot is available internal server error response
func (o *SetSlotIsAvailableInternalServerError) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *SetSlotIsAvailableInternalServerError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(500)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
