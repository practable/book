// Code generated by go-swagger; DO NOT EDIT.

package users

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/timdrysdale/interval/serve/models"
)

// GetAccessTokenOKCode is the HTTP code returned for type GetAccessTokenOK
const GetAccessTokenOKCode int = 200

/*GetAccessTokenOK OK

swagger:response getAccessTokenOK
*/
type GetAccessTokenOK struct {

	/*
	  In: Body
	*/
	Payload *models.AccessToken `json:"body,omitempty"`
}

// NewGetAccessTokenOK creates GetAccessTokenOK with default headers values
func NewGetAccessTokenOK() *GetAccessTokenOK {

	return &GetAccessTokenOK{}
}

// WithPayload adds the payload to the get access token o k response
func (o *GetAccessTokenOK) WithPayload(payload *models.AccessToken) *GetAccessTokenOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get access token o k response
func (o *GetAccessTokenOK) SetPayload(payload *models.AccessToken) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetAccessTokenOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetAccessTokenUnauthorizedCode is the HTTP code returned for type GetAccessTokenUnauthorized
const GetAccessTokenUnauthorizedCode int = 401

/*GetAccessTokenUnauthorized Unauthorized

swagger:response getAccessTokenUnauthorized
*/
type GetAccessTokenUnauthorized struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetAccessTokenUnauthorized creates GetAccessTokenUnauthorized with default headers values
func NewGetAccessTokenUnauthorized() *GetAccessTokenUnauthorized {

	return &GetAccessTokenUnauthorized{}
}

// WithPayload adds the payload to the get access token unauthorized response
func (o *GetAccessTokenUnauthorized) WithPayload(payload *models.Error) *GetAccessTokenUnauthorized {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get access token unauthorized response
func (o *GetAccessTokenUnauthorized) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetAccessTokenUnauthorized) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(401)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetAccessTokenNotFoundCode is the HTTP code returned for type GetAccessTokenNotFound
const GetAccessTokenNotFoundCode int = 404

/*GetAccessTokenNotFound The specified resource was not found

swagger:response getAccessTokenNotFound
*/
type GetAccessTokenNotFound struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetAccessTokenNotFound creates GetAccessTokenNotFound with default headers values
func NewGetAccessTokenNotFound() *GetAccessTokenNotFound {

	return &GetAccessTokenNotFound{}
}

// WithPayload adds the payload to the get access token not found response
func (o *GetAccessTokenNotFound) WithPayload(payload *models.Error) *GetAccessTokenNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get access token not found response
func (o *GetAccessTokenNotFound) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetAccessTokenNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetAccessTokenInternalServerErrorCode is the HTTP code returned for type GetAccessTokenInternalServerError
const GetAccessTokenInternalServerErrorCode int = 500

/*GetAccessTokenInternalServerError Internal Error

swagger:response getAccessTokenInternalServerError
*/
type GetAccessTokenInternalServerError struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetAccessTokenInternalServerError creates GetAccessTokenInternalServerError with default headers values
func NewGetAccessTokenInternalServerError() *GetAccessTokenInternalServerError {

	return &GetAccessTokenInternalServerError{}
}

// WithPayload adds the payload to the get access token internal server error response
func (o *GetAccessTokenInternalServerError) WithPayload(payload *models.Error) *GetAccessTokenInternalServerError {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get access token internal server error response
func (o *GetAccessTokenInternalServerError) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetAccessTokenInternalServerError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(500)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
