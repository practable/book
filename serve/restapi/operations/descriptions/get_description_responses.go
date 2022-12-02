// Code generated by go-swagger; DO NOT EDIT.

package descriptions

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/timdrysdale/interval/serve/models"
)

// GetDescriptionOKCode is the HTTP code returned for type GetDescriptionOK
const GetDescriptionOKCode int = 200

/*GetDescriptionOK OK

swagger:response getDescriptionOK
*/
type GetDescriptionOK struct {

	/*
	  In: Body
	*/
	Payload *models.Description `json:"body,omitempty"`
}

// NewGetDescriptionOK creates GetDescriptionOK with default headers values
func NewGetDescriptionOK() *GetDescriptionOK {

	return &GetDescriptionOK{}
}

// WithPayload adds the payload to the get description o k response
func (o *GetDescriptionOK) WithPayload(payload *models.Description) *GetDescriptionOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get description o k response
func (o *GetDescriptionOK) SetPayload(payload *models.Description) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetDescriptionOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetDescriptionUnauthorizedCode is the HTTP code returned for type GetDescriptionUnauthorized
const GetDescriptionUnauthorizedCode int = 401

/*GetDescriptionUnauthorized Unauthorized

swagger:response getDescriptionUnauthorized
*/
type GetDescriptionUnauthorized struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetDescriptionUnauthorized creates GetDescriptionUnauthorized with default headers values
func NewGetDescriptionUnauthorized() *GetDescriptionUnauthorized {

	return &GetDescriptionUnauthorized{}
}

// WithPayload adds the payload to the get description unauthorized response
func (o *GetDescriptionUnauthorized) WithPayload(payload *models.Error) *GetDescriptionUnauthorized {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get description unauthorized response
func (o *GetDescriptionUnauthorized) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetDescriptionUnauthorized) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(401)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetDescriptionNotFoundCode is the HTTP code returned for type GetDescriptionNotFound
const GetDescriptionNotFoundCode int = 404

/*GetDescriptionNotFound The specified resource was not found

swagger:response getDescriptionNotFound
*/
type GetDescriptionNotFound struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetDescriptionNotFound creates GetDescriptionNotFound with default headers values
func NewGetDescriptionNotFound() *GetDescriptionNotFound {

	return &GetDescriptionNotFound{}
}

// WithPayload adds the payload to the get description not found response
func (o *GetDescriptionNotFound) WithPayload(payload *models.Error) *GetDescriptionNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get description not found response
func (o *GetDescriptionNotFound) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetDescriptionNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetDescriptionInternalServerErrorCode is the HTTP code returned for type GetDescriptionInternalServerError
const GetDescriptionInternalServerErrorCode int = 500

/*GetDescriptionInternalServerError Internal Error

swagger:response getDescriptionInternalServerError
*/
type GetDescriptionInternalServerError struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetDescriptionInternalServerError creates GetDescriptionInternalServerError with default headers values
func NewGetDescriptionInternalServerError() *GetDescriptionInternalServerError {

	return &GetDescriptionInternalServerError{}
}

// WithPayload adds the payload to the get description internal server error response
func (o *GetDescriptionInternalServerError) WithPayload(payload *models.Error) *GetDescriptionInternalServerError {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get description internal server error response
func (o *GetDescriptionInternalServerError) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetDescriptionInternalServerError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(500)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}