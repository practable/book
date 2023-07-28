// Code generated by go-swagger; DO NOT EDIT.

package admin

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/practable/book/internal/serve/models"
)

// GetResourceIsAvailableOKCode is the HTTP code returned for type GetResourceIsAvailableOK
const GetResourceIsAvailableOKCode int = 200

/*GetResourceIsAvailableOK OK

swagger:response getResourceIsAvailableOK
*/
type GetResourceIsAvailableOK struct {

	/*
	  In: Body
	*/
	Payload *models.ResourceStatus `json:"body,omitempty"`
}

// NewGetResourceIsAvailableOK creates GetResourceIsAvailableOK with default headers values
func NewGetResourceIsAvailableOK() *GetResourceIsAvailableOK {

	return &GetResourceIsAvailableOK{}
}

// WithPayload adds the payload to the get resource is available o k response
func (o *GetResourceIsAvailableOK) WithPayload(payload *models.ResourceStatus) *GetResourceIsAvailableOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get resource is available o k response
func (o *GetResourceIsAvailableOK) SetPayload(payload *models.ResourceStatus) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetResourceIsAvailableOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetResourceIsAvailableUnauthorizedCode is the HTTP code returned for type GetResourceIsAvailableUnauthorized
const GetResourceIsAvailableUnauthorizedCode int = 401

/*GetResourceIsAvailableUnauthorized Unauthorized

swagger:response getResourceIsAvailableUnauthorized
*/
type GetResourceIsAvailableUnauthorized struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetResourceIsAvailableUnauthorized creates GetResourceIsAvailableUnauthorized with default headers values
func NewGetResourceIsAvailableUnauthorized() *GetResourceIsAvailableUnauthorized {

	return &GetResourceIsAvailableUnauthorized{}
}

// WithPayload adds the payload to the get resource is available unauthorized response
func (o *GetResourceIsAvailableUnauthorized) WithPayload(payload *models.Error) *GetResourceIsAvailableUnauthorized {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get resource is available unauthorized response
func (o *GetResourceIsAvailableUnauthorized) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetResourceIsAvailableUnauthorized) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(401)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetResourceIsAvailableNotFoundCode is the HTTP code returned for type GetResourceIsAvailableNotFound
const GetResourceIsAvailableNotFoundCode int = 404

/*GetResourceIsAvailableNotFound The specified resource was not found

swagger:response getResourceIsAvailableNotFound
*/
type GetResourceIsAvailableNotFound struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetResourceIsAvailableNotFound creates GetResourceIsAvailableNotFound with default headers values
func NewGetResourceIsAvailableNotFound() *GetResourceIsAvailableNotFound {

	return &GetResourceIsAvailableNotFound{}
}

// WithPayload adds the payload to the get resource is available not found response
func (o *GetResourceIsAvailableNotFound) WithPayload(payload *models.Error) *GetResourceIsAvailableNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get resource is available not found response
func (o *GetResourceIsAvailableNotFound) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetResourceIsAvailableNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetResourceIsAvailableInternalServerErrorCode is the HTTP code returned for type GetResourceIsAvailableInternalServerError
const GetResourceIsAvailableInternalServerErrorCode int = 500

/*GetResourceIsAvailableInternalServerError Internal Error

swagger:response getResourceIsAvailableInternalServerError
*/
type GetResourceIsAvailableInternalServerError struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetResourceIsAvailableInternalServerError creates GetResourceIsAvailableInternalServerError with default headers values
func NewGetResourceIsAvailableInternalServerError() *GetResourceIsAvailableInternalServerError {

	return &GetResourceIsAvailableInternalServerError{}
}

// WithPayload adds the payload to the get resource is available internal server error response
func (o *GetResourceIsAvailableInternalServerError) WithPayload(payload *models.Error) *GetResourceIsAvailableInternalServerError {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get resource is available internal server error response
func (o *GetResourceIsAvailableInternalServerError) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetResourceIsAvailableInternalServerError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(500)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
