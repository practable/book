// Code generated by go-swagger; DO NOT EDIT.

package admin

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/timdrysdale/interval/internal/serve/models"
)

// GetSlotIsAvailableOKCode is the HTTP code returned for type GetSlotIsAvailableOK
const GetSlotIsAvailableOKCode int = 200

/*GetSlotIsAvailableOK OK

swagger:response getSlotIsAvailableOK
*/
type GetSlotIsAvailableOK struct {

	/*
	  In: Body
	*/
	Payload *models.SlotStatus `json:"body,omitempty"`
}

// NewGetSlotIsAvailableOK creates GetSlotIsAvailableOK with default headers values
func NewGetSlotIsAvailableOK() *GetSlotIsAvailableOK {

	return &GetSlotIsAvailableOK{}
}

// WithPayload adds the payload to the get slot is available o k response
func (o *GetSlotIsAvailableOK) WithPayload(payload *models.SlotStatus) *GetSlotIsAvailableOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get slot is available o k response
func (o *GetSlotIsAvailableOK) SetPayload(payload *models.SlotStatus) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetSlotIsAvailableOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetSlotIsAvailableUnauthorizedCode is the HTTP code returned for type GetSlotIsAvailableUnauthorized
const GetSlotIsAvailableUnauthorizedCode int = 401

/*GetSlotIsAvailableUnauthorized Unauthorized

swagger:response getSlotIsAvailableUnauthorized
*/
type GetSlotIsAvailableUnauthorized struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetSlotIsAvailableUnauthorized creates GetSlotIsAvailableUnauthorized with default headers values
func NewGetSlotIsAvailableUnauthorized() *GetSlotIsAvailableUnauthorized {

	return &GetSlotIsAvailableUnauthorized{}
}

// WithPayload adds the payload to the get slot is available unauthorized response
func (o *GetSlotIsAvailableUnauthorized) WithPayload(payload *models.Error) *GetSlotIsAvailableUnauthorized {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get slot is available unauthorized response
func (o *GetSlotIsAvailableUnauthorized) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetSlotIsAvailableUnauthorized) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(401)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetSlotIsAvailableNotFoundCode is the HTTP code returned for type GetSlotIsAvailableNotFound
const GetSlotIsAvailableNotFoundCode int = 404

/*GetSlotIsAvailableNotFound The specified resource was not found

swagger:response getSlotIsAvailableNotFound
*/
type GetSlotIsAvailableNotFound struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetSlotIsAvailableNotFound creates GetSlotIsAvailableNotFound with default headers values
func NewGetSlotIsAvailableNotFound() *GetSlotIsAvailableNotFound {

	return &GetSlotIsAvailableNotFound{}
}

// WithPayload adds the payload to the get slot is available not found response
func (o *GetSlotIsAvailableNotFound) WithPayload(payload *models.Error) *GetSlotIsAvailableNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get slot is available not found response
func (o *GetSlotIsAvailableNotFound) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetSlotIsAvailableNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetSlotIsAvailableInternalServerErrorCode is the HTTP code returned for type GetSlotIsAvailableInternalServerError
const GetSlotIsAvailableInternalServerErrorCode int = 500

/*GetSlotIsAvailableInternalServerError Internal Error

swagger:response getSlotIsAvailableInternalServerError
*/
type GetSlotIsAvailableInternalServerError struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetSlotIsAvailableInternalServerError creates GetSlotIsAvailableInternalServerError with default headers values
func NewGetSlotIsAvailableInternalServerError() *GetSlotIsAvailableInternalServerError {

	return &GetSlotIsAvailableInternalServerError{}
}

// WithPayload adds the payload to the get slot is available internal server error response
func (o *GetSlotIsAvailableInternalServerError) WithPayload(payload *models.Error) *GetSlotIsAvailableInternalServerError {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get slot is available internal server error response
func (o *GetSlotIsAvailableInternalServerError) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetSlotIsAvailableInternalServerError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(500)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
