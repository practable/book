// Code generated by go-swagger; DO NOT EDIT.

package users

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/practable/book/internal/serve/models"
)

// GetActivityOKCode is the HTTP code returned for type GetActivityOK
const GetActivityOKCode int = 200

/*GetActivityOK OK

swagger:response getActivityOK
*/
type GetActivityOK struct {

	/*
	  In: Body
	*/
	Payload *models.Activity `json:"body,omitempty"`
}

// NewGetActivityOK creates GetActivityOK with default headers values
func NewGetActivityOK() *GetActivityOK {

	return &GetActivityOK{}
}

// WithPayload adds the payload to the get activity o k response
func (o *GetActivityOK) WithPayload(payload *models.Activity) *GetActivityOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get activity o k response
func (o *GetActivityOK) SetPayload(payload *models.Activity) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetActivityOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetActivityUnauthorizedCode is the HTTP code returned for type GetActivityUnauthorized
const GetActivityUnauthorizedCode int = 401

/*GetActivityUnauthorized Unauthorized

swagger:response getActivityUnauthorized
*/
type GetActivityUnauthorized struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetActivityUnauthorized creates GetActivityUnauthorized with default headers values
func NewGetActivityUnauthorized() *GetActivityUnauthorized {

	return &GetActivityUnauthorized{}
}

// WithPayload adds the payload to the get activity unauthorized response
func (o *GetActivityUnauthorized) WithPayload(payload *models.Error) *GetActivityUnauthorized {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get activity unauthorized response
func (o *GetActivityUnauthorized) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetActivityUnauthorized) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(401)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetActivityNotFoundCode is the HTTP code returned for type GetActivityNotFound
const GetActivityNotFoundCode int = 404

/*GetActivityNotFound The specified resource was not found

swagger:response getActivityNotFound
*/
type GetActivityNotFound struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetActivityNotFound creates GetActivityNotFound with default headers values
func NewGetActivityNotFound() *GetActivityNotFound {

	return &GetActivityNotFound{}
}

// WithPayload adds the payload to the get activity not found response
func (o *GetActivityNotFound) WithPayload(payload *models.Error) *GetActivityNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get activity not found response
func (o *GetActivityNotFound) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetActivityNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetActivityInternalServerErrorCode is the HTTP code returned for type GetActivityInternalServerError
const GetActivityInternalServerErrorCode int = 500

/*GetActivityInternalServerError Internal Error

swagger:response getActivityInternalServerError
*/
type GetActivityInternalServerError struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetActivityInternalServerError creates GetActivityInternalServerError with default headers values
func NewGetActivityInternalServerError() *GetActivityInternalServerError {

	return &GetActivityInternalServerError{}
}

// WithPayload adds the payload to the get activity internal server error response
func (o *GetActivityInternalServerError) WithPayload(payload *models.Error) *GetActivityInternalServerError {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get activity internal server error response
func (o *GetActivityInternalServerError) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetActivityInternalServerError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(500)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
