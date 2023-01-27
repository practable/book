// Code generated by go-swagger; DO NOT EDIT.

package users

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/practable/book/internal/client/models"
)

// CancelBookingReader is a Reader for the CancelBooking structure.
type CancelBookingReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *CancelBookingReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 401:
		result := NewCancelBookingUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewCancelBookingNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewCancelBookingInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewCancelBookingUnauthorized creates a CancelBookingUnauthorized with default headers values
func NewCancelBookingUnauthorized() *CancelBookingUnauthorized {
	return &CancelBookingUnauthorized{}
}

/* CancelBookingUnauthorized describes a response with status code 401, with default header values.

Unauthorized
*/
type CancelBookingUnauthorized struct {
	Payload *models.Error
}

// IsSuccess returns true when this cancel booking unauthorized response has a 2xx status code
func (o *CancelBookingUnauthorized) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this cancel booking unauthorized response has a 3xx status code
func (o *CancelBookingUnauthorized) IsRedirect() bool {
	return false
}

// IsClientError returns true when this cancel booking unauthorized response has a 4xx status code
func (o *CancelBookingUnauthorized) IsClientError() bool {
	return true
}

// IsServerError returns true when this cancel booking unauthorized response has a 5xx status code
func (o *CancelBookingUnauthorized) IsServerError() bool {
	return false
}

// IsCode returns true when this cancel booking unauthorized response a status code equal to that given
func (o *CancelBookingUnauthorized) IsCode(code int) bool {
	return code == 401
}

func (o *CancelBookingUnauthorized) Error() string {
	return fmt.Sprintf("[DELETE /users/{user_name}/bookings/{booking_name}][%d] cancelBookingUnauthorized  %+v", 401, o.Payload)
}

func (o *CancelBookingUnauthorized) String() string {
	return fmt.Sprintf("[DELETE /users/{user_name}/bookings/{booking_name}][%d] cancelBookingUnauthorized  %+v", 401, o.Payload)
}

func (o *CancelBookingUnauthorized) GetPayload() *models.Error {
	return o.Payload
}

func (o *CancelBookingUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewCancelBookingNotFound creates a CancelBookingNotFound with default headers values
func NewCancelBookingNotFound() *CancelBookingNotFound {
	return &CancelBookingNotFound{}
}

/* CancelBookingNotFound describes a response with status code 404, with default header values.

The specified resource was not found
*/
type CancelBookingNotFound struct {
	Payload *models.Error
}

// IsSuccess returns true when this cancel booking not found response has a 2xx status code
func (o *CancelBookingNotFound) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this cancel booking not found response has a 3xx status code
func (o *CancelBookingNotFound) IsRedirect() bool {
	return false
}

// IsClientError returns true when this cancel booking not found response has a 4xx status code
func (o *CancelBookingNotFound) IsClientError() bool {
	return true
}

// IsServerError returns true when this cancel booking not found response has a 5xx status code
func (o *CancelBookingNotFound) IsServerError() bool {
	return false
}

// IsCode returns true when this cancel booking not found response a status code equal to that given
func (o *CancelBookingNotFound) IsCode(code int) bool {
	return code == 404
}

func (o *CancelBookingNotFound) Error() string {
	return fmt.Sprintf("[DELETE /users/{user_name}/bookings/{booking_name}][%d] cancelBookingNotFound  %+v", 404, o.Payload)
}

func (o *CancelBookingNotFound) String() string {
	return fmt.Sprintf("[DELETE /users/{user_name}/bookings/{booking_name}][%d] cancelBookingNotFound  %+v", 404, o.Payload)
}

func (o *CancelBookingNotFound) GetPayload() *models.Error {
	return o.Payload
}

func (o *CancelBookingNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewCancelBookingInternalServerError creates a CancelBookingInternalServerError with default headers values
func NewCancelBookingInternalServerError() *CancelBookingInternalServerError {
	return &CancelBookingInternalServerError{}
}

/* CancelBookingInternalServerError describes a response with status code 500, with default header values.

Internal Error
*/
type CancelBookingInternalServerError struct {
	Payload *models.Error
}

// IsSuccess returns true when this cancel booking internal server error response has a 2xx status code
func (o *CancelBookingInternalServerError) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this cancel booking internal server error response has a 3xx status code
func (o *CancelBookingInternalServerError) IsRedirect() bool {
	return false
}

// IsClientError returns true when this cancel booking internal server error response has a 4xx status code
func (o *CancelBookingInternalServerError) IsClientError() bool {
	return false
}

// IsServerError returns true when this cancel booking internal server error response has a 5xx status code
func (o *CancelBookingInternalServerError) IsServerError() bool {
	return true
}

// IsCode returns true when this cancel booking internal server error response a status code equal to that given
func (o *CancelBookingInternalServerError) IsCode(code int) bool {
	return code == 500
}

func (o *CancelBookingInternalServerError) Error() string {
	return fmt.Sprintf("[DELETE /users/{user_name}/bookings/{booking_name}][%d] cancelBookingInternalServerError  %+v", 500, o.Payload)
}

func (o *CancelBookingInternalServerError) String() string {
	return fmt.Sprintf("[DELETE /users/{user_name}/bookings/{booking_name}][%d] cancelBookingInternalServerError  %+v", 500, o.Payload)
}

func (o *CancelBookingInternalServerError) GetPayload() *models.Error {
	return o.Payload
}

func (o *CancelBookingInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
