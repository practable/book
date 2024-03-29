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

// MakeBookingReader is a Reader for the MakeBooking structure.
type MakeBookingReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *MakeBookingReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 204:
		result := NewMakeBookingNoContent()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 401:
		result := NewMakeBookingUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewMakeBookingNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 409:
		result := NewMakeBookingConflict()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewMakeBookingInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewMakeBookingNoContent creates a MakeBookingNoContent with default headers values
func NewMakeBookingNoContent() *MakeBookingNoContent {
	return &MakeBookingNoContent{}
}

/* MakeBookingNoContent describes a response with status code 204, with default header values.

OK - No Content
*/
type MakeBookingNoContent struct {
}

// IsSuccess returns true when this make booking no content response has a 2xx status code
func (o *MakeBookingNoContent) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this make booking no content response has a 3xx status code
func (o *MakeBookingNoContent) IsRedirect() bool {
	return false
}

// IsClientError returns true when this make booking no content response has a 4xx status code
func (o *MakeBookingNoContent) IsClientError() bool {
	return false
}

// IsServerError returns true when this make booking no content response has a 5xx status code
func (o *MakeBookingNoContent) IsServerError() bool {
	return false
}

// IsCode returns true when this make booking no content response a status code equal to that given
func (o *MakeBookingNoContent) IsCode(code int) bool {
	return code == 204
}

func (o *MakeBookingNoContent) Error() string {
	return fmt.Sprintf("[POST /slots/{slot_name}][%d] makeBookingNoContent ", 204)
}

func (o *MakeBookingNoContent) String() string {
	return fmt.Sprintf("[POST /slots/{slot_name}][%d] makeBookingNoContent ", 204)
}

func (o *MakeBookingNoContent) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewMakeBookingUnauthorized creates a MakeBookingUnauthorized with default headers values
func NewMakeBookingUnauthorized() *MakeBookingUnauthorized {
	return &MakeBookingUnauthorized{}
}

/* MakeBookingUnauthorized describes a response with status code 401, with default header values.

Unauthorized
*/
type MakeBookingUnauthorized struct {
	Payload *models.Error
}

// IsSuccess returns true when this make booking unauthorized response has a 2xx status code
func (o *MakeBookingUnauthorized) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this make booking unauthorized response has a 3xx status code
func (o *MakeBookingUnauthorized) IsRedirect() bool {
	return false
}

// IsClientError returns true when this make booking unauthorized response has a 4xx status code
func (o *MakeBookingUnauthorized) IsClientError() bool {
	return true
}

// IsServerError returns true when this make booking unauthorized response has a 5xx status code
func (o *MakeBookingUnauthorized) IsServerError() bool {
	return false
}

// IsCode returns true when this make booking unauthorized response a status code equal to that given
func (o *MakeBookingUnauthorized) IsCode(code int) bool {
	return code == 401
}

func (o *MakeBookingUnauthorized) Error() string {
	return fmt.Sprintf("[POST /slots/{slot_name}][%d] makeBookingUnauthorized  %+v", 401, o.Payload)
}

func (o *MakeBookingUnauthorized) String() string {
	return fmt.Sprintf("[POST /slots/{slot_name}][%d] makeBookingUnauthorized  %+v", 401, o.Payload)
}

func (o *MakeBookingUnauthorized) GetPayload() *models.Error {
	return o.Payload
}

func (o *MakeBookingUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewMakeBookingNotFound creates a MakeBookingNotFound with default headers values
func NewMakeBookingNotFound() *MakeBookingNotFound {
	return &MakeBookingNotFound{}
}

/* MakeBookingNotFound describes a response with status code 404, with default header values.

The specified resource was not found
*/
type MakeBookingNotFound struct {
	Payload *models.Error
}

// IsSuccess returns true when this make booking not found response has a 2xx status code
func (o *MakeBookingNotFound) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this make booking not found response has a 3xx status code
func (o *MakeBookingNotFound) IsRedirect() bool {
	return false
}

// IsClientError returns true when this make booking not found response has a 4xx status code
func (o *MakeBookingNotFound) IsClientError() bool {
	return true
}

// IsServerError returns true when this make booking not found response has a 5xx status code
func (o *MakeBookingNotFound) IsServerError() bool {
	return false
}

// IsCode returns true when this make booking not found response a status code equal to that given
func (o *MakeBookingNotFound) IsCode(code int) bool {
	return code == 404
}

func (o *MakeBookingNotFound) Error() string {
	return fmt.Sprintf("[POST /slots/{slot_name}][%d] makeBookingNotFound  %+v", 404, o.Payload)
}

func (o *MakeBookingNotFound) String() string {
	return fmt.Sprintf("[POST /slots/{slot_name}][%d] makeBookingNotFound  %+v", 404, o.Payload)
}

func (o *MakeBookingNotFound) GetPayload() *models.Error {
	return o.Payload
}

func (o *MakeBookingNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewMakeBookingConflict creates a MakeBookingConflict with default headers values
func NewMakeBookingConflict() *MakeBookingConflict {
	return &MakeBookingConflict{}
}

/* MakeBookingConflict describes a response with status code 409, with default header values.

Conflict - unavailable for the requested interval
*/
type MakeBookingConflict struct {
	Payload interface{}
}

// IsSuccess returns true when this make booking conflict response has a 2xx status code
func (o *MakeBookingConflict) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this make booking conflict response has a 3xx status code
func (o *MakeBookingConflict) IsRedirect() bool {
	return false
}

// IsClientError returns true when this make booking conflict response has a 4xx status code
func (o *MakeBookingConflict) IsClientError() bool {
	return true
}

// IsServerError returns true when this make booking conflict response has a 5xx status code
func (o *MakeBookingConflict) IsServerError() bool {
	return false
}

// IsCode returns true when this make booking conflict response a status code equal to that given
func (o *MakeBookingConflict) IsCode(code int) bool {
	return code == 409
}

func (o *MakeBookingConflict) Error() string {
	return fmt.Sprintf("[POST /slots/{slot_name}][%d] makeBookingConflict  %+v", 409, o.Payload)
}

func (o *MakeBookingConflict) String() string {
	return fmt.Sprintf("[POST /slots/{slot_name}][%d] makeBookingConflict  %+v", 409, o.Payload)
}

func (o *MakeBookingConflict) GetPayload() interface{} {
	return o.Payload
}

func (o *MakeBookingConflict) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewMakeBookingInternalServerError creates a MakeBookingInternalServerError with default headers values
func NewMakeBookingInternalServerError() *MakeBookingInternalServerError {
	return &MakeBookingInternalServerError{}
}

/* MakeBookingInternalServerError describes a response with status code 500, with default header values.

Internal Error
*/
type MakeBookingInternalServerError struct {
	Payload *models.Error
}

// IsSuccess returns true when this make booking internal server error response has a 2xx status code
func (o *MakeBookingInternalServerError) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this make booking internal server error response has a 3xx status code
func (o *MakeBookingInternalServerError) IsRedirect() bool {
	return false
}

// IsClientError returns true when this make booking internal server error response has a 4xx status code
func (o *MakeBookingInternalServerError) IsClientError() bool {
	return false
}

// IsServerError returns true when this make booking internal server error response has a 5xx status code
func (o *MakeBookingInternalServerError) IsServerError() bool {
	return true
}

// IsCode returns true when this make booking internal server error response a status code equal to that given
func (o *MakeBookingInternalServerError) IsCode(code int) bool {
	return code == 500
}

func (o *MakeBookingInternalServerError) Error() string {
	return fmt.Sprintf("[POST /slots/{slot_name}][%d] makeBookingInternalServerError  %+v", 500, o.Payload)
}

func (o *MakeBookingInternalServerError) String() string {
	return fmt.Sprintf("[POST /slots/{slot_name}][%d] makeBookingInternalServerError  %+v", 500, o.Payload)
}

func (o *MakeBookingInternalServerError) GetPayload() *models.Error {
	return o.Payload
}

func (o *MakeBookingInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
