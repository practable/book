// Code generated by go-swagger; DO NOT EDIT.

package users

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/timdrysdale/interval/internal/serve/models"
)

// MakeBookingNoContentCode is the HTTP code returned for type MakeBookingNoContent
const MakeBookingNoContentCode int = 204

/*MakeBookingNoContent OK - No Content

swagger:response makeBookingNoContent
*/
type MakeBookingNoContent struct {
}

// NewMakeBookingNoContent creates MakeBookingNoContent with default headers values
func NewMakeBookingNoContent() *MakeBookingNoContent {

	return &MakeBookingNoContent{}
}

// WriteResponse to the client
func (o *MakeBookingNoContent) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(204)
}

// MakeBookingUnauthorizedCode is the HTTP code returned for type MakeBookingUnauthorized
const MakeBookingUnauthorizedCode int = 401

/*MakeBookingUnauthorized Unauthorized

swagger:response makeBookingUnauthorized
*/
type MakeBookingUnauthorized struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewMakeBookingUnauthorized creates MakeBookingUnauthorized with default headers values
func NewMakeBookingUnauthorized() *MakeBookingUnauthorized {

	return &MakeBookingUnauthorized{}
}

// WithPayload adds the payload to the make booking unauthorized response
func (o *MakeBookingUnauthorized) WithPayload(payload *models.Error) *MakeBookingUnauthorized {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the make booking unauthorized response
func (o *MakeBookingUnauthorized) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *MakeBookingUnauthorized) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(401)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// MakeBookingNotFoundCode is the HTTP code returned for type MakeBookingNotFound
const MakeBookingNotFoundCode int = 404

/*MakeBookingNotFound The specified resource was not found

swagger:response makeBookingNotFound
*/
type MakeBookingNotFound struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewMakeBookingNotFound creates MakeBookingNotFound with default headers values
func NewMakeBookingNotFound() *MakeBookingNotFound {

	return &MakeBookingNotFound{}
}

// WithPayload adds the payload to the make booking not found response
func (o *MakeBookingNotFound) WithPayload(payload *models.Error) *MakeBookingNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the make booking not found response
func (o *MakeBookingNotFound) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *MakeBookingNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// MakeBookingConflictCode is the HTTP code returned for type MakeBookingConflict
const MakeBookingConflictCode int = 409

/*MakeBookingConflict Conflict - unavailable for the requested interval

swagger:response makeBookingConflict
*/
type MakeBookingConflict struct {

	/*
	  In: Body
	*/
	Payload interface{} `json:"body,omitempty"`
}

// NewMakeBookingConflict creates MakeBookingConflict with default headers values
func NewMakeBookingConflict() *MakeBookingConflict {

	return &MakeBookingConflict{}
}

// WithPayload adds the payload to the make booking conflict response
func (o *MakeBookingConflict) WithPayload(payload interface{}) *MakeBookingConflict {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the make booking conflict response
func (o *MakeBookingConflict) SetPayload(payload interface{}) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *MakeBookingConflict) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(409)
	payload := o.Payload
	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}
}

// MakeBookingInternalServerErrorCode is the HTTP code returned for type MakeBookingInternalServerError
const MakeBookingInternalServerErrorCode int = 500

/*MakeBookingInternalServerError Internal Error

swagger:response makeBookingInternalServerError
*/
type MakeBookingInternalServerError struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewMakeBookingInternalServerError creates MakeBookingInternalServerError with default headers values
func NewMakeBookingInternalServerError() *MakeBookingInternalServerError {

	return &MakeBookingInternalServerError{}
}

// WithPayload adds the payload to the make booking internal server error response
func (o *MakeBookingInternalServerError) WithPayload(payload *models.Error) *MakeBookingInternalServerError {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the make booking internal server error response
func (o *MakeBookingInternalServerError) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *MakeBookingInternalServerError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(500)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
