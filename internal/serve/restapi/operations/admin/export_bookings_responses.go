// Code generated by go-swagger; DO NOT EDIT.

package admin

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/timdrysdale/interval/internal/serve/models"
)

// ExportBookingsOKCode is the HTTP code returned for type ExportBookingsOK
const ExportBookingsOKCode int = 200

/*ExportBookingsOK OK

swagger:response exportBookingsOK
*/
type ExportBookingsOK struct {

	/*
	  In: Body
	*/
	Payload []*models.Booking `json:"body,omitempty"`
}

// NewExportBookingsOK creates ExportBookingsOK with default headers values
func NewExportBookingsOK() *ExportBookingsOK {

	return &ExportBookingsOK{}
}

// WithPayload adds the payload to the export bookings o k response
func (o *ExportBookingsOK) WithPayload(payload []*models.Booking) *ExportBookingsOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the export bookings o k response
func (o *ExportBookingsOK) SetPayload(payload []*models.Booking) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *ExportBookingsOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	payload := o.Payload
	if payload == nil {
		// return empty array
		payload = make([]*models.Booking, 0, 50)
	}

	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}
}

// ExportBookingsUnauthorizedCode is the HTTP code returned for type ExportBookingsUnauthorized
const ExportBookingsUnauthorizedCode int = 401

/*ExportBookingsUnauthorized Unauthorized

swagger:response exportBookingsUnauthorized
*/
type ExportBookingsUnauthorized struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewExportBookingsUnauthorized creates ExportBookingsUnauthorized with default headers values
func NewExportBookingsUnauthorized() *ExportBookingsUnauthorized {

	return &ExportBookingsUnauthorized{}
}

// WithPayload adds the payload to the export bookings unauthorized response
func (o *ExportBookingsUnauthorized) WithPayload(payload *models.Error) *ExportBookingsUnauthorized {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the export bookings unauthorized response
func (o *ExportBookingsUnauthorized) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *ExportBookingsUnauthorized) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(401)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// ExportBookingsNotFoundCode is the HTTP code returned for type ExportBookingsNotFound
const ExportBookingsNotFoundCode int = 404

/*ExportBookingsNotFound The specified resource was not found

swagger:response exportBookingsNotFound
*/
type ExportBookingsNotFound struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewExportBookingsNotFound creates ExportBookingsNotFound with default headers values
func NewExportBookingsNotFound() *ExportBookingsNotFound {

	return &ExportBookingsNotFound{}
}

// WithPayload adds the payload to the export bookings not found response
func (o *ExportBookingsNotFound) WithPayload(payload *models.Error) *ExportBookingsNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the export bookings not found response
func (o *ExportBookingsNotFound) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *ExportBookingsNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// ExportBookingsInternalServerErrorCode is the HTTP code returned for type ExportBookingsInternalServerError
const ExportBookingsInternalServerErrorCode int = 500

/*ExportBookingsInternalServerError Internal Error

swagger:response exportBookingsInternalServerError
*/
type ExportBookingsInternalServerError struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewExportBookingsInternalServerError creates ExportBookingsInternalServerError with default headers values
func NewExportBookingsInternalServerError() *ExportBookingsInternalServerError {

	return &ExportBookingsInternalServerError{}
}

// WithPayload adds the payload to the export bookings internal server error response
func (o *ExportBookingsInternalServerError) WithPayload(payload *models.Error) *ExportBookingsInternalServerError {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the export bookings internal server error response
func (o *ExportBookingsInternalServerError) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *ExportBookingsInternalServerError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(500)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
