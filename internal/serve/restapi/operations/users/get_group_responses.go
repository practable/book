// Code generated by go-swagger; DO NOT EDIT.

package users

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/practable/book/internal/serve/models"
)

// GetGroupOKCode is the HTTP code returned for type GetGroupOK
const GetGroupOKCode int = 200

/*GetGroupOK OK

swagger:response getGroupOK
*/
type GetGroupOK struct {

	/*
	  In: Body
	*/
	Payload *models.GroupDescribedWithPolicies `json:"body,omitempty"`
}

// NewGetGroupOK creates GetGroupOK with default headers values
func NewGetGroupOK() *GetGroupOK {

	return &GetGroupOK{}
}

// WithPayload adds the payload to the get group o k response
func (o *GetGroupOK) WithPayload(payload *models.GroupDescribedWithPolicies) *GetGroupOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get group o k response
func (o *GetGroupOK) SetPayload(payload *models.GroupDescribedWithPolicies) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetGroupOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetGroupUnauthorizedCode is the HTTP code returned for type GetGroupUnauthorized
const GetGroupUnauthorizedCode int = 401

/*GetGroupUnauthorized Unauthorized

swagger:response getGroupUnauthorized
*/
type GetGroupUnauthorized struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetGroupUnauthorized creates GetGroupUnauthorized with default headers values
func NewGetGroupUnauthorized() *GetGroupUnauthorized {

	return &GetGroupUnauthorized{}
}

// WithPayload adds the payload to the get group unauthorized response
func (o *GetGroupUnauthorized) WithPayload(payload *models.Error) *GetGroupUnauthorized {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get group unauthorized response
func (o *GetGroupUnauthorized) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetGroupUnauthorized) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(401)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetGroupNotFoundCode is the HTTP code returned for type GetGroupNotFound
const GetGroupNotFoundCode int = 404

/*GetGroupNotFound The specified resource was not found

swagger:response getGroupNotFound
*/
type GetGroupNotFound struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetGroupNotFound creates GetGroupNotFound with default headers values
func NewGetGroupNotFound() *GetGroupNotFound {

	return &GetGroupNotFound{}
}

// WithPayload adds the payload to the get group not found response
func (o *GetGroupNotFound) WithPayload(payload *models.Error) *GetGroupNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get group not found response
func (o *GetGroupNotFound) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetGroupNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetGroupInternalServerErrorCode is the HTTP code returned for type GetGroupInternalServerError
const GetGroupInternalServerErrorCode int = 500

/*GetGroupInternalServerError Internal Error

swagger:response getGroupInternalServerError
*/
type GetGroupInternalServerError struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetGroupInternalServerError creates GetGroupInternalServerError with default headers values
func NewGetGroupInternalServerError() *GetGroupInternalServerError {

	return &GetGroupInternalServerError{}
}

// WithPayload adds the payload to the get group internal server error response
func (o *GetGroupInternalServerError) WithPayload(payload *models.Error) *GetGroupInternalServerError {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get group internal server error response
func (o *GetGroupInternalServerError) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetGroupInternalServerError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(500)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
