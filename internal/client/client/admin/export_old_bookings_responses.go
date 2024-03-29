// Code generated by go-swagger; DO NOT EDIT.

package admin

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/practable/book/internal/client/models"
)

// ExportOldBookingsReader is a Reader for the ExportOldBookings structure.
type ExportOldBookingsReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *ExportOldBookingsReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewExportOldBookingsOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 401:
		result := NewExportOldBookingsUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewExportOldBookingsNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewExportOldBookingsInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewExportOldBookingsOK creates a ExportOldBookingsOK with default headers values
func NewExportOldBookingsOK() *ExportOldBookingsOK {
	return &ExportOldBookingsOK{}
}

/* ExportOldBookingsOK describes a response with status code 200, with default header values.

OK
*/
type ExportOldBookingsOK struct {
	Payload models.Bookings
}

// IsSuccess returns true when this export old bookings o k response has a 2xx status code
func (o *ExportOldBookingsOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this export old bookings o k response has a 3xx status code
func (o *ExportOldBookingsOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this export old bookings o k response has a 4xx status code
func (o *ExportOldBookingsOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this export old bookings o k response has a 5xx status code
func (o *ExportOldBookingsOK) IsServerError() bool {
	return false
}

// IsCode returns true when this export old bookings o k response a status code equal to that given
func (o *ExportOldBookingsOK) IsCode(code int) bool {
	return code == 200
}

func (o *ExportOldBookingsOK) Error() string {
	return fmt.Sprintf("[GET /admin/oldbookings][%d] exportOldBookingsOK  %+v", 200, o.Payload)
}

func (o *ExportOldBookingsOK) String() string {
	return fmt.Sprintf("[GET /admin/oldbookings][%d] exportOldBookingsOK  %+v", 200, o.Payload)
}

func (o *ExportOldBookingsOK) GetPayload() models.Bookings {
	return o.Payload
}

func (o *ExportOldBookingsOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewExportOldBookingsUnauthorized creates a ExportOldBookingsUnauthorized with default headers values
func NewExportOldBookingsUnauthorized() *ExportOldBookingsUnauthorized {
	return &ExportOldBookingsUnauthorized{}
}

/* ExportOldBookingsUnauthorized describes a response with status code 401, with default header values.

Unauthorized
*/
type ExportOldBookingsUnauthorized struct {
	Payload *models.Error
}

// IsSuccess returns true when this export old bookings unauthorized response has a 2xx status code
func (o *ExportOldBookingsUnauthorized) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this export old bookings unauthorized response has a 3xx status code
func (o *ExportOldBookingsUnauthorized) IsRedirect() bool {
	return false
}

// IsClientError returns true when this export old bookings unauthorized response has a 4xx status code
func (o *ExportOldBookingsUnauthorized) IsClientError() bool {
	return true
}

// IsServerError returns true when this export old bookings unauthorized response has a 5xx status code
func (o *ExportOldBookingsUnauthorized) IsServerError() bool {
	return false
}

// IsCode returns true when this export old bookings unauthorized response a status code equal to that given
func (o *ExportOldBookingsUnauthorized) IsCode(code int) bool {
	return code == 401
}

func (o *ExportOldBookingsUnauthorized) Error() string {
	return fmt.Sprintf("[GET /admin/oldbookings][%d] exportOldBookingsUnauthorized  %+v", 401, o.Payload)
}

func (o *ExportOldBookingsUnauthorized) String() string {
	return fmt.Sprintf("[GET /admin/oldbookings][%d] exportOldBookingsUnauthorized  %+v", 401, o.Payload)
}

func (o *ExportOldBookingsUnauthorized) GetPayload() *models.Error {
	return o.Payload
}

func (o *ExportOldBookingsUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewExportOldBookingsNotFound creates a ExportOldBookingsNotFound with default headers values
func NewExportOldBookingsNotFound() *ExportOldBookingsNotFound {
	return &ExportOldBookingsNotFound{}
}

/* ExportOldBookingsNotFound describes a response with status code 404, with default header values.

The specified resource was not found
*/
type ExportOldBookingsNotFound struct {
	Payload *models.Error
}

// IsSuccess returns true when this export old bookings not found response has a 2xx status code
func (o *ExportOldBookingsNotFound) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this export old bookings not found response has a 3xx status code
func (o *ExportOldBookingsNotFound) IsRedirect() bool {
	return false
}

// IsClientError returns true when this export old bookings not found response has a 4xx status code
func (o *ExportOldBookingsNotFound) IsClientError() bool {
	return true
}

// IsServerError returns true when this export old bookings not found response has a 5xx status code
func (o *ExportOldBookingsNotFound) IsServerError() bool {
	return false
}

// IsCode returns true when this export old bookings not found response a status code equal to that given
func (o *ExportOldBookingsNotFound) IsCode(code int) bool {
	return code == 404
}

func (o *ExportOldBookingsNotFound) Error() string {
	return fmt.Sprintf("[GET /admin/oldbookings][%d] exportOldBookingsNotFound  %+v", 404, o.Payload)
}

func (o *ExportOldBookingsNotFound) String() string {
	return fmt.Sprintf("[GET /admin/oldbookings][%d] exportOldBookingsNotFound  %+v", 404, o.Payload)
}

func (o *ExportOldBookingsNotFound) GetPayload() *models.Error {
	return o.Payload
}

func (o *ExportOldBookingsNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewExportOldBookingsInternalServerError creates a ExportOldBookingsInternalServerError with default headers values
func NewExportOldBookingsInternalServerError() *ExportOldBookingsInternalServerError {
	return &ExportOldBookingsInternalServerError{}
}

/* ExportOldBookingsInternalServerError describes a response with status code 500, with default header values.

Internal Error
*/
type ExportOldBookingsInternalServerError struct {
	Payload *models.Error
}

// IsSuccess returns true when this export old bookings internal server error response has a 2xx status code
func (o *ExportOldBookingsInternalServerError) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this export old bookings internal server error response has a 3xx status code
func (o *ExportOldBookingsInternalServerError) IsRedirect() bool {
	return false
}

// IsClientError returns true when this export old bookings internal server error response has a 4xx status code
func (o *ExportOldBookingsInternalServerError) IsClientError() bool {
	return false
}

// IsServerError returns true when this export old bookings internal server error response has a 5xx status code
func (o *ExportOldBookingsInternalServerError) IsServerError() bool {
	return true
}

// IsCode returns true when this export old bookings internal server error response a status code equal to that given
func (o *ExportOldBookingsInternalServerError) IsCode(code int) bool {
	return code == 500
}

func (o *ExportOldBookingsInternalServerError) Error() string {
	return fmt.Sprintf("[GET /admin/oldbookings][%d] exportOldBookingsInternalServerError  %+v", 500, o.Payload)
}

func (o *ExportOldBookingsInternalServerError) String() string {
	return fmt.Sprintf("[GET /admin/oldbookings][%d] exportOldBookingsInternalServerError  %+v", 500, o.Payload)
}

func (o *ExportOldBookingsInternalServerError) GetPayload() *models.Error {
	return o.Payload
}

func (o *ExportOldBookingsInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
