// Code generated by go-swagger; DO NOT EDIT.

package users

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/practable/book/internal/serve/models"
)

// GetStoreStatusUserOKCode is the HTTP code returned for type GetStoreStatusUserOK
const GetStoreStatusUserOKCode int = 200

/*GetStoreStatusUserOK OK

swagger:response getStoreStatusUserOK
*/
type GetStoreStatusUserOK struct {

	/*
	  In: Body
	*/
	Payload *models.StoreStatusUser `json:"body,omitempty"`
}

// NewGetStoreStatusUserOK creates GetStoreStatusUserOK with default headers values
func NewGetStoreStatusUserOK() *GetStoreStatusUserOK {

	return &GetStoreStatusUserOK{}
}

// WithPayload adds the payload to the get store status user o k response
func (o *GetStoreStatusUserOK) WithPayload(payload *models.StoreStatusUser) *GetStoreStatusUserOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get store status user o k response
func (o *GetStoreStatusUserOK) SetPayload(payload *models.StoreStatusUser) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetStoreStatusUserOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetStoreStatusUserUnauthorizedCode is the HTTP code returned for type GetStoreStatusUserUnauthorized
const GetStoreStatusUserUnauthorizedCode int = 401

/*GetStoreStatusUserUnauthorized Unauthorized

swagger:response getStoreStatusUserUnauthorized
*/
type GetStoreStatusUserUnauthorized struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetStoreStatusUserUnauthorized creates GetStoreStatusUserUnauthorized with default headers values
func NewGetStoreStatusUserUnauthorized() *GetStoreStatusUserUnauthorized {

	return &GetStoreStatusUserUnauthorized{}
}

// WithPayload adds the payload to the get store status user unauthorized response
func (o *GetStoreStatusUserUnauthorized) WithPayload(payload *models.Error) *GetStoreStatusUserUnauthorized {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get store status user unauthorized response
func (o *GetStoreStatusUserUnauthorized) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetStoreStatusUserUnauthorized) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(401)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetStoreStatusUserNotFoundCode is the HTTP code returned for type GetStoreStatusUserNotFound
const GetStoreStatusUserNotFoundCode int = 404

/*GetStoreStatusUserNotFound The specified resource was not found

swagger:response getStoreStatusUserNotFound
*/
type GetStoreStatusUserNotFound struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetStoreStatusUserNotFound creates GetStoreStatusUserNotFound with default headers values
func NewGetStoreStatusUserNotFound() *GetStoreStatusUserNotFound {

	return &GetStoreStatusUserNotFound{}
}

// WithPayload adds the payload to the get store status user not found response
func (o *GetStoreStatusUserNotFound) WithPayload(payload *models.Error) *GetStoreStatusUserNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get store status user not found response
func (o *GetStoreStatusUserNotFound) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetStoreStatusUserNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetStoreStatusUserInternalServerErrorCode is the HTTP code returned for type GetStoreStatusUserInternalServerError
const GetStoreStatusUserInternalServerErrorCode int = 500

/*GetStoreStatusUserInternalServerError Internal Error

swagger:response getStoreStatusUserInternalServerError
*/
type GetStoreStatusUserInternalServerError struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetStoreStatusUserInternalServerError creates GetStoreStatusUserInternalServerError with default headers values
func NewGetStoreStatusUserInternalServerError() *GetStoreStatusUserInternalServerError {

	return &GetStoreStatusUserInternalServerError{}
}

// WithPayload adds the payload to the get store status user internal server error response
func (o *GetStoreStatusUserInternalServerError) WithPayload(payload *models.Error) *GetStoreStatusUserInternalServerError {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get store status user internal server error response
func (o *GetStoreStatusUserInternalServerError) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetStoreStatusUserInternalServerError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(500)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
