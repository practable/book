// Code generated by go-swagger; DO NOT EDIT.

package admin

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/practable/book/internal/serve/models"
)

// GetResourcesOKCode is the HTTP code returned for type GetResourcesOK
const GetResourcesOKCode int = 200

/*GetResourcesOK OK

swagger:response getResourcesOK
*/
type GetResourcesOK struct {

	/*
	  In: Body
	*/
	Payload models.Resources `json:"body,omitempty"`
}

// NewGetResourcesOK creates GetResourcesOK with default headers values
func NewGetResourcesOK() *GetResourcesOK {

	return &GetResourcesOK{}
}

// WithPayload adds the payload to the get resources o k response
func (o *GetResourcesOK) WithPayload(payload models.Resources) *GetResourcesOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get resources o k response
func (o *GetResourcesOK) SetPayload(payload models.Resources) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetResourcesOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	payload := o.Payload
	if payload == nil {
		// return empty map
		payload = models.Resources{}
	}

	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}
}

// GetResourcesUnauthorizedCode is the HTTP code returned for type GetResourcesUnauthorized
const GetResourcesUnauthorizedCode int = 401

/*GetResourcesUnauthorized Unauthorized

swagger:response getResourcesUnauthorized
*/
type GetResourcesUnauthorized struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetResourcesUnauthorized creates GetResourcesUnauthorized with default headers values
func NewGetResourcesUnauthorized() *GetResourcesUnauthorized {

	return &GetResourcesUnauthorized{}
}

// WithPayload adds the payload to the get resources unauthorized response
func (o *GetResourcesUnauthorized) WithPayload(payload *models.Error) *GetResourcesUnauthorized {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get resources unauthorized response
func (o *GetResourcesUnauthorized) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetResourcesUnauthorized) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(401)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetResourcesNotFoundCode is the HTTP code returned for type GetResourcesNotFound
const GetResourcesNotFoundCode int = 404

/*GetResourcesNotFound The specified resource was not found

swagger:response getResourcesNotFound
*/
type GetResourcesNotFound struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetResourcesNotFound creates GetResourcesNotFound with default headers values
func NewGetResourcesNotFound() *GetResourcesNotFound {

	return &GetResourcesNotFound{}
}

// WithPayload adds the payload to the get resources not found response
func (o *GetResourcesNotFound) WithPayload(payload *models.Error) *GetResourcesNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get resources not found response
func (o *GetResourcesNotFound) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetResourcesNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetResourcesInternalServerErrorCode is the HTTP code returned for type GetResourcesInternalServerError
const GetResourcesInternalServerErrorCode int = 500

/*GetResourcesInternalServerError Internal Error

swagger:response getResourcesInternalServerError
*/
type GetResourcesInternalServerError struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetResourcesInternalServerError creates GetResourcesInternalServerError with default headers values
func NewGetResourcesInternalServerError() *GetResourcesInternalServerError {

	return &GetResourcesInternalServerError{}
}

// WithPayload adds the payload to the get resources internal server error response
func (o *GetResourcesInternalServerError) WithPayload(payload *models.Error) *GetResourcesInternalServerError {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get resources internal server error response
func (o *GetResourcesInternalServerError) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetResourcesInternalServerError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(500)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}