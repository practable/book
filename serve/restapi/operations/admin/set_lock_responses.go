// Code generated by go-swagger; DO NOT EDIT.

package admin

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/timdrysdale/interval/serve/models"
)

// SetLockOKCode is the HTTP code returned for type SetLockOK
const SetLockOKCode int = 200

/*SetLockOK OK

swagger:response setLockOK
*/
type SetLockOK struct {

	/*
	  In: Body
	*/
	Payload *models.StoreStatusAdmin `json:"body,omitempty"`
}

// NewSetLockOK creates SetLockOK with default headers values
func NewSetLockOK() *SetLockOK {

	return &SetLockOK{}
}

// WithPayload adds the payload to the set lock o k response
func (o *SetLockOK) WithPayload(payload *models.StoreStatusAdmin) *SetLockOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the set lock o k response
func (o *SetLockOK) SetPayload(payload *models.StoreStatusAdmin) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *SetLockOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// SetLockUnauthorizedCode is the HTTP code returned for type SetLockUnauthorized
const SetLockUnauthorizedCode int = 401

/*SetLockUnauthorized Unauthorized

swagger:response setLockUnauthorized
*/
type SetLockUnauthorized struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewSetLockUnauthorized creates SetLockUnauthorized with default headers values
func NewSetLockUnauthorized() *SetLockUnauthorized {

	return &SetLockUnauthorized{}
}

// WithPayload adds the payload to the set lock unauthorized response
func (o *SetLockUnauthorized) WithPayload(payload *models.Error) *SetLockUnauthorized {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the set lock unauthorized response
func (o *SetLockUnauthorized) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *SetLockUnauthorized) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(401)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// SetLockNotFoundCode is the HTTP code returned for type SetLockNotFound
const SetLockNotFoundCode int = 404

/*SetLockNotFound The specified resource was not found

swagger:response setLockNotFound
*/
type SetLockNotFound struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewSetLockNotFound creates SetLockNotFound with default headers values
func NewSetLockNotFound() *SetLockNotFound {

	return &SetLockNotFound{}
}

// WithPayload adds the payload to the set lock not found response
func (o *SetLockNotFound) WithPayload(payload *models.Error) *SetLockNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the set lock not found response
func (o *SetLockNotFound) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *SetLockNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// SetLockInternalServerErrorCode is the HTTP code returned for type SetLockInternalServerError
const SetLockInternalServerErrorCode int = 500

/*SetLockInternalServerError Internal Error

swagger:response setLockInternalServerError
*/
type SetLockInternalServerError struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewSetLockInternalServerError creates SetLockInternalServerError with default headers values
func NewSetLockInternalServerError() *SetLockInternalServerError {

	return &SetLockInternalServerError{}
}

// WithPayload adds the payload to the set lock internal server error response
func (o *SetLockInternalServerError) WithPayload(payload *models.Error) *SetLockInternalServerError {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the set lock internal server error response
func (o *SetLockInternalServerError) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *SetLockInternalServerError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(500)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}