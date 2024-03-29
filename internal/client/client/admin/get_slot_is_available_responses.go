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

// GetSlotIsAvailableReader is a Reader for the GetSlotIsAvailable structure.
type GetSlotIsAvailableReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetSlotIsAvailableReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetSlotIsAvailableOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 401:
		result := NewGetSlotIsAvailableUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewGetSlotIsAvailableNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewGetSlotIsAvailableInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewGetSlotIsAvailableOK creates a GetSlotIsAvailableOK with default headers values
func NewGetSlotIsAvailableOK() *GetSlotIsAvailableOK {
	return &GetSlotIsAvailableOK{}
}

/* GetSlotIsAvailableOK describes a response with status code 200, with default header values.

OK
*/
type GetSlotIsAvailableOK struct {
	Payload *models.SlotStatus
}

// IsSuccess returns true when this get slot is available o k response has a 2xx status code
func (o *GetSlotIsAvailableOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this get slot is available o k response has a 3xx status code
func (o *GetSlotIsAvailableOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get slot is available o k response has a 4xx status code
func (o *GetSlotIsAvailableOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this get slot is available o k response has a 5xx status code
func (o *GetSlotIsAvailableOK) IsServerError() bool {
	return false
}

// IsCode returns true when this get slot is available o k response a status code equal to that given
func (o *GetSlotIsAvailableOK) IsCode(code int) bool {
	return code == 200
}

func (o *GetSlotIsAvailableOK) Error() string {
	return fmt.Sprintf("[GET /admin/slots/{slot_name}][%d] getSlotIsAvailableOK  %+v", 200, o.Payload)
}

func (o *GetSlotIsAvailableOK) String() string {
	return fmt.Sprintf("[GET /admin/slots/{slot_name}][%d] getSlotIsAvailableOK  %+v", 200, o.Payload)
}

func (o *GetSlotIsAvailableOK) GetPayload() *models.SlotStatus {
	return o.Payload
}

func (o *GetSlotIsAvailableOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.SlotStatus)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetSlotIsAvailableUnauthorized creates a GetSlotIsAvailableUnauthorized with default headers values
func NewGetSlotIsAvailableUnauthorized() *GetSlotIsAvailableUnauthorized {
	return &GetSlotIsAvailableUnauthorized{}
}

/* GetSlotIsAvailableUnauthorized describes a response with status code 401, with default header values.

Unauthorized
*/
type GetSlotIsAvailableUnauthorized struct {
	Payload *models.Error
}

// IsSuccess returns true when this get slot is available unauthorized response has a 2xx status code
func (o *GetSlotIsAvailableUnauthorized) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get slot is available unauthorized response has a 3xx status code
func (o *GetSlotIsAvailableUnauthorized) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get slot is available unauthorized response has a 4xx status code
func (o *GetSlotIsAvailableUnauthorized) IsClientError() bool {
	return true
}

// IsServerError returns true when this get slot is available unauthorized response has a 5xx status code
func (o *GetSlotIsAvailableUnauthorized) IsServerError() bool {
	return false
}

// IsCode returns true when this get slot is available unauthorized response a status code equal to that given
func (o *GetSlotIsAvailableUnauthorized) IsCode(code int) bool {
	return code == 401
}

func (o *GetSlotIsAvailableUnauthorized) Error() string {
	return fmt.Sprintf("[GET /admin/slots/{slot_name}][%d] getSlotIsAvailableUnauthorized  %+v", 401, o.Payload)
}

func (o *GetSlotIsAvailableUnauthorized) String() string {
	return fmt.Sprintf("[GET /admin/slots/{slot_name}][%d] getSlotIsAvailableUnauthorized  %+v", 401, o.Payload)
}

func (o *GetSlotIsAvailableUnauthorized) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetSlotIsAvailableUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetSlotIsAvailableNotFound creates a GetSlotIsAvailableNotFound with default headers values
func NewGetSlotIsAvailableNotFound() *GetSlotIsAvailableNotFound {
	return &GetSlotIsAvailableNotFound{}
}

/* GetSlotIsAvailableNotFound describes a response with status code 404, with default header values.

The specified resource was not found
*/
type GetSlotIsAvailableNotFound struct {
	Payload *models.Error
}

// IsSuccess returns true when this get slot is available not found response has a 2xx status code
func (o *GetSlotIsAvailableNotFound) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get slot is available not found response has a 3xx status code
func (o *GetSlotIsAvailableNotFound) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get slot is available not found response has a 4xx status code
func (o *GetSlotIsAvailableNotFound) IsClientError() bool {
	return true
}

// IsServerError returns true when this get slot is available not found response has a 5xx status code
func (o *GetSlotIsAvailableNotFound) IsServerError() bool {
	return false
}

// IsCode returns true when this get slot is available not found response a status code equal to that given
func (o *GetSlotIsAvailableNotFound) IsCode(code int) bool {
	return code == 404
}

func (o *GetSlotIsAvailableNotFound) Error() string {
	return fmt.Sprintf("[GET /admin/slots/{slot_name}][%d] getSlotIsAvailableNotFound  %+v", 404, o.Payload)
}

func (o *GetSlotIsAvailableNotFound) String() string {
	return fmt.Sprintf("[GET /admin/slots/{slot_name}][%d] getSlotIsAvailableNotFound  %+v", 404, o.Payload)
}

func (o *GetSlotIsAvailableNotFound) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetSlotIsAvailableNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetSlotIsAvailableInternalServerError creates a GetSlotIsAvailableInternalServerError with default headers values
func NewGetSlotIsAvailableInternalServerError() *GetSlotIsAvailableInternalServerError {
	return &GetSlotIsAvailableInternalServerError{}
}

/* GetSlotIsAvailableInternalServerError describes a response with status code 500, with default header values.

Internal Error
*/
type GetSlotIsAvailableInternalServerError struct {
	Payload *models.Error
}

// IsSuccess returns true when this get slot is available internal server error response has a 2xx status code
func (o *GetSlotIsAvailableInternalServerError) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get slot is available internal server error response has a 3xx status code
func (o *GetSlotIsAvailableInternalServerError) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get slot is available internal server error response has a 4xx status code
func (o *GetSlotIsAvailableInternalServerError) IsClientError() bool {
	return false
}

// IsServerError returns true when this get slot is available internal server error response has a 5xx status code
func (o *GetSlotIsAvailableInternalServerError) IsServerError() bool {
	return true
}

// IsCode returns true when this get slot is available internal server error response a status code equal to that given
func (o *GetSlotIsAvailableInternalServerError) IsCode(code int) bool {
	return code == 500
}

func (o *GetSlotIsAvailableInternalServerError) Error() string {
	return fmt.Sprintf("[GET /admin/slots/{slot_name}][%d] getSlotIsAvailableInternalServerError  %+v", 500, o.Payload)
}

func (o *GetSlotIsAvailableInternalServerError) String() string {
	return fmt.Sprintf("[GET /admin/slots/{slot_name}][%d] getSlotIsAvailableInternalServerError  %+v", 500, o.Payload)
}

func (o *GetSlotIsAvailableInternalServerError) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetSlotIsAvailableInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
