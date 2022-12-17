// Code generated by go-swagger; DO NOT EDIT.

package users

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/timdrysdale/interval/internal/client/models"
)

// GetAvailabilityReader is a Reader for the GetAvailability structure.
type GetAvailabilityReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetAvailabilityReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetAvailabilityOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 401:
		result := NewGetAvailabilityUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewGetAvailabilityNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewGetAvailabilityInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewGetAvailabilityOK creates a GetAvailabilityOK with default headers values
func NewGetAvailabilityOK() *GetAvailabilityOK {
	return &GetAvailabilityOK{}
}

/*
GetAvailabilityOK describes a response with status code 200, with default header values.

OK
*/
type GetAvailabilityOK struct {
	Payload models.Intervals
}

// IsSuccess returns true when this get availability o k response has a 2xx status code
func (o *GetAvailabilityOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this get availability o k response has a 3xx status code
func (o *GetAvailabilityOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get availability o k response has a 4xx status code
func (o *GetAvailabilityOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this get availability o k response has a 5xx status code
func (o *GetAvailabilityOK) IsServerError() bool {
	return false
}

// IsCode returns true when this get availability o k response a status code equal to that given
func (o *GetAvailabilityOK) IsCode(code int) bool {
	return code == 200
}

func (o *GetAvailabilityOK) Error() string {
	return fmt.Sprintf("[GET /policies/{policy_name}/slots/{slot_name}][%d] getAvailabilityOK  %+v", 200, o.Payload)
}

func (o *GetAvailabilityOK) String() string {
	return fmt.Sprintf("[GET /policies/{policy_name}/slots/{slot_name}][%d] getAvailabilityOK  %+v", 200, o.Payload)
}

func (o *GetAvailabilityOK) GetPayload() models.Intervals {
	return o.Payload
}

func (o *GetAvailabilityOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetAvailabilityUnauthorized creates a GetAvailabilityUnauthorized with default headers values
func NewGetAvailabilityUnauthorized() *GetAvailabilityUnauthorized {
	return &GetAvailabilityUnauthorized{}
}

/*
GetAvailabilityUnauthorized describes a response with status code 401, with default header values.

Unauthorized
*/
type GetAvailabilityUnauthorized struct {
	Payload *models.Error
}

// IsSuccess returns true when this get availability unauthorized response has a 2xx status code
func (o *GetAvailabilityUnauthorized) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get availability unauthorized response has a 3xx status code
func (o *GetAvailabilityUnauthorized) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get availability unauthorized response has a 4xx status code
func (o *GetAvailabilityUnauthorized) IsClientError() bool {
	return true
}

// IsServerError returns true when this get availability unauthorized response has a 5xx status code
func (o *GetAvailabilityUnauthorized) IsServerError() bool {
	return false
}

// IsCode returns true when this get availability unauthorized response a status code equal to that given
func (o *GetAvailabilityUnauthorized) IsCode(code int) bool {
	return code == 401
}

func (o *GetAvailabilityUnauthorized) Error() string {
	return fmt.Sprintf("[GET /policies/{policy_name}/slots/{slot_name}][%d] getAvailabilityUnauthorized  %+v", 401, o.Payload)
}

func (o *GetAvailabilityUnauthorized) String() string {
	return fmt.Sprintf("[GET /policies/{policy_name}/slots/{slot_name}][%d] getAvailabilityUnauthorized  %+v", 401, o.Payload)
}

func (o *GetAvailabilityUnauthorized) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetAvailabilityUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetAvailabilityNotFound creates a GetAvailabilityNotFound with default headers values
func NewGetAvailabilityNotFound() *GetAvailabilityNotFound {
	return &GetAvailabilityNotFound{}
}

/*
GetAvailabilityNotFound describes a response with status code 404, with default header values.

The specified resource was not found
*/
type GetAvailabilityNotFound struct {
	Payload *models.Error
}

// IsSuccess returns true when this get availability not found response has a 2xx status code
func (o *GetAvailabilityNotFound) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get availability not found response has a 3xx status code
func (o *GetAvailabilityNotFound) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get availability not found response has a 4xx status code
func (o *GetAvailabilityNotFound) IsClientError() bool {
	return true
}

// IsServerError returns true when this get availability not found response has a 5xx status code
func (o *GetAvailabilityNotFound) IsServerError() bool {
	return false
}

// IsCode returns true when this get availability not found response a status code equal to that given
func (o *GetAvailabilityNotFound) IsCode(code int) bool {
	return code == 404
}

func (o *GetAvailabilityNotFound) Error() string {
	return fmt.Sprintf("[GET /policies/{policy_name}/slots/{slot_name}][%d] getAvailabilityNotFound  %+v", 404, o.Payload)
}

func (o *GetAvailabilityNotFound) String() string {
	return fmt.Sprintf("[GET /policies/{policy_name}/slots/{slot_name}][%d] getAvailabilityNotFound  %+v", 404, o.Payload)
}

func (o *GetAvailabilityNotFound) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetAvailabilityNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetAvailabilityInternalServerError creates a GetAvailabilityInternalServerError with default headers values
func NewGetAvailabilityInternalServerError() *GetAvailabilityInternalServerError {
	return &GetAvailabilityInternalServerError{}
}

/*
GetAvailabilityInternalServerError describes a response with status code 500, with default header values.

Internal Error
*/
type GetAvailabilityInternalServerError struct {
	Payload *models.Error
}

// IsSuccess returns true when this get availability internal server error response has a 2xx status code
func (o *GetAvailabilityInternalServerError) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get availability internal server error response has a 3xx status code
func (o *GetAvailabilityInternalServerError) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get availability internal server error response has a 4xx status code
func (o *GetAvailabilityInternalServerError) IsClientError() bool {
	return false
}

// IsServerError returns true when this get availability internal server error response has a 5xx status code
func (o *GetAvailabilityInternalServerError) IsServerError() bool {
	return true
}

// IsCode returns true when this get availability internal server error response a status code equal to that given
func (o *GetAvailabilityInternalServerError) IsCode(code int) bool {
	return code == 500
}

func (o *GetAvailabilityInternalServerError) Error() string {
	return fmt.Sprintf("[GET /policies/{policy_name}/slots/{slot_name}][%d] getAvailabilityInternalServerError  %+v", 500, o.Payload)
}

func (o *GetAvailabilityInternalServerError) String() string {
	return fmt.Sprintf("[GET /policies/{policy_name}/slots/{slot_name}][%d] getAvailabilityInternalServerError  %+v", 500, o.Payload)
}

func (o *GetAvailabilityInternalServerError) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetAvailabilityInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}