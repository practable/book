// Code generated by go-swagger; DO NOT EDIT.

package admin

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/timdrysdale/interval/internal/client/models"
)

// ReplaceBookingsReader is a Reader for the ReplaceBookings structure.
type ReplaceBookingsReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *ReplaceBookingsReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewReplaceBookingsOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 401:
		result := NewReplaceBookingsUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewReplaceBookingsNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewReplaceBookingsInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewReplaceBookingsOK creates a ReplaceBookingsOK with default headers values
func NewReplaceBookingsOK() *ReplaceBookingsOK {
	return &ReplaceBookingsOK{}
}

/* ReplaceBookingsOK describes a response with status code 200, with default header values.

OK
*/
type ReplaceBookingsOK struct {
	Payload *models.StoreStatusAdmin
}

// IsSuccess returns true when this replace bookings o k response has a 2xx status code
func (o *ReplaceBookingsOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this replace bookings o k response has a 3xx status code
func (o *ReplaceBookingsOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this replace bookings o k response has a 4xx status code
func (o *ReplaceBookingsOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this replace bookings o k response has a 5xx status code
func (o *ReplaceBookingsOK) IsServerError() bool {
	return false
}

// IsCode returns true when this replace bookings o k response a status code equal to that given
func (o *ReplaceBookingsOK) IsCode(code int) bool {
	return code == 200
}

func (o *ReplaceBookingsOK) Error() string {
	return fmt.Sprintf("[PUT /admin/bookings][%d] replaceBookingsOK  %+v", 200, o.Payload)
}

func (o *ReplaceBookingsOK) String() string {
	return fmt.Sprintf("[PUT /admin/bookings][%d] replaceBookingsOK  %+v", 200, o.Payload)
}

func (o *ReplaceBookingsOK) GetPayload() *models.StoreStatusAdmin {
	return o.Payload
}

func (o *ReplaceBookingsOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.StoreStatusAdmin)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewReplaceBookingsUnauthorized creates a ReplaceBookingsUnauthorized with default headers values
func NewReplaceBookingsUnauthorized() *ReplaceBookingsUnauthorized {
	return &ReplaceBookingsUnauthorized{}
}

/* ReplaceBookingsUnauthorized describes a response with status code 401, with default header values.

Unauthorized
*/
type ReplaceBookingsUnauthorized struct {
	Payload *models.Error
}

// IsSuccess returns true when this replace bookings unauthorized response has a 2xx status code
func (o *ReplaceBookingsUnauthorized) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this replace bookings unauthorized response has a 3xx status code
func (o *ReplaceBookingsUnauthorized) IsRedirect() bool {
	return false
}

// IsClientError returns true when this replace bookings unauthorized response has a 4xx status code
func (o *ReplaceBookingsUnauthorized) IsClientError() bool {
	return true
}

// IsServerError returns true when this replace bookings unauthorized response has a 5xx status code
func (o *ReplaceBookingsUnauthorized) IsServerError() bool {
	return false
}

// IsCode returns true when this replace bookings unauthorized response a status code equal to that given
func (o *ReplaceBookingsUnauthorized) IsCode(code int) bool {
	return code == 401
}

func (o *ReplaceBookingsUnauthorized) Error() string {
	return fmt.Sprintf("[PUT /admin/bookings][%d] replaceBookingsUnauthorized  %+v", 401, o.Payload)
}

func (o *ReplaceBookingsUnauthorized) String() string {
	return fmt.Sprintf("[PUT /admin/bookings][%d] replaceBookingsUnauthorized  %+v", 401, o.Payload)
}

func (o *ReplaceBookingsUnauthorized) GetPayload() *models.Error {
	return o.Payload
}

func (o *ReplaceBookingsUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewReplaceBookingsNotFound creates a ReplaceBookingsNotFound with default headers values
func NewReplaceBookingsNotFound() *ReplaceBookingsNotFound {
	return &ReplaceBookingsNotFound{}
}

/* ReplaceBookingsNotFound describes a response with status code 404, with default header values.

The specified resource was not found
*/
type ReplaceBookingsNotFound struct {
	Payload *models.Error
}

// IsSuccess returns true when this replace bookings not found response has a 2xx status code
func (o *ReplaceBookingsNotFound) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this replace bookings not found response has a 3xx status code
func (o *ReplaceBookingsNotFound) IsRedirect() bool {
	return false
}

// IsClientError returns true when this replace bookings not found response has a 4xx status code
func (o *ReplaceBookingsNotFound) IsClientError() bool {
	return true
}

// IsServerError returns true when this replace bookings not found response has a 5xx status code
func (o *ReplaceBookingsNotFound) IsServerError() bool {
	return false
}

// IsCode returns true when this replace bookings not found response a status code equal to that given
func (o *ReplaceBookingsNotFound) IsCode(code int) bool {
	return code == 404
}

func (o *ReplaceBookingsNotFound) Error() string {
	return fmt.Sprintf("[PUT /admin/bookings][%d] replaceBookingsNotFound  %+v", 404, o.Payload)
}

func (o *ReplaceBookingsNotFound) String() string {
	return fmt.Sprintf("[PUT /admin/bookings][%d] replaceBookingsNotFound  %+v", 404, o.Payload)
}

func (o *ReplaceBookingsNotFound) GetPayload() *models.Error {
	return o.Payload
}

func (o *ReplaceBookingsNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewReplaceBookingsInternalServerError creates a ReplaceBookingsInternalServerError with default headers values
func NewReplaceBookingsInternalServerError() *ReplaceBookingsInternalServerError {
	return &ReplaceBookingsInternalServerError{}
}

/* ReplaceBookingsInternalServerError describes a response with status code 500, with default header values.

Internal Error
*/
type ReplaceBookingsInternalServerError struct {
	Payload *models.Error
}

// IsSuccess returns true when this replace bookings internal server error response has a 2xx status code
func (o *ReplaceBookingsInternalServerError) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this replace bookings internal server error response has a 3xx status code
func (o *ReplaceBookingsInternalServerError) IsRedirect() bool {
	return false
}

// IsClientError returns true when this replace bookings internal server error response has a 4xx status code
func (o *ReplaceBookingsInternalServerError) IsClientError() bool {
	return false
}

// IsServerError returns true when this replace bookings internal server error response has a 5xx status code
func (o *ReplaceBookingsInternalServerError) IsServerError() bool {
	return true
}

// IsCode returns true when this replace bookings internal server error response a status code equal to that given
func (o *ReplaceBookingsInternalServerError) IsCode(code int) bool {
	return code == 500
}

func (o *ReplaceBookingsInternalServerError) Error() string {
	return fmt.Sprintf("[PUT /admin/bookings][%d] replaceBookingsInternalServerError  %+v", 500, o.Payload)
}

func (o *ReplaceBookingsInternalServerError) String() string {
	return fmt.Sprintf("[PUT /admin/bookings][%d] replaceBookingsInternalServerError  %+v", 500, o.Payload)
}

func (o *ReplaceBookingsInternalServerError) GetPayload() *models.Error {
	return o.Payload
}

func (o *ReplaceBookingsInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
