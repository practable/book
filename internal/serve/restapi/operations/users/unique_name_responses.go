// Code generated by go-swagger; DO NOT EDIT.

package users

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/practable/book/internal/serve/models"
)

// UniqueNameOKCode is the HTTP code returned for type UniqueNameOK
const UniqueNameOKCode int = 200

/*UniqueNameOK OK

swagger:response uniqueNameOK
*/
type UniqueNameOK struct {

	/*
	  In: Body
	*/
	Payload *models.UserName `json:"body,omitempty"`
}

// NewUniqueNameOK creates UniqueNameOK with default headers values
func NewUniqueNameOK() *UniqueNameOK {

	return &UniqueNameOK{}
}

// WithPayload adds the payload to the unique name o k response
func (o *UniqueNameOK) WithPayload(payload *models.UserName) *UniqueNameOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the unique name o k response
func (o *UniqueNameOK) SetPayload(payload *models.UserName) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *UniqueNameOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// UniqueNameUnauthorizedCode is the HTTP code returned for type UniqueNameUnauthorized
const UniqueNameUnauthorizedCode int = 401

/*UniqueNameUnauthorized Unauthorized

swagger:response uniqueNameUnauthorized
*/
type UniqueNameUnauthorized struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewUniqueNameUnauthorized creates UniqueNameUnauthorized with default headers values
func NewUniqueNameUnauthorized() *UniqueNameUnauthorized {

	return &UniqueNameUnauthorized{}
}

// WithPayload adds the payload to the unique name unauthorized response
func (o *UniqueNameUnauthorized) WithPayload(payload *models.Error) *UniqueNameUnauthorized {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the unique name unauthorized response
func (o *UniqueNameUnauthorized) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *UniqueNameUnauthorized) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(401)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// UniqueNameInternalServerErrorCode is the HTTP code returned for type UniqueNameInternalServerError
const UniqueNameInternalServerErrorCode int = 500

/*UniqueNameInternalServerError Internal Error

swagger:response uniqueNameInternalServerError
*/
type UniqueNameInternalServerError struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewUniqueNameInternalServerError creates UniqueNameInternalServerError with default headers values
func NewUniqueNameInternalServerError() *UniqueNameInternalServerError {

	return &UniqueNameInternalServerError{}
}

// WithPayload adds the payload to the unique name internal server error response
func (o *UniqueNameInternalServerError) WithPayload(payload *models.Error) *UniqueNameInternalServerError {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the unique name internal server error response
func (o *UniqueNameInternalServerError) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *UniqueNameInternalServerError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(500)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
