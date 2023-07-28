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

// GetResourceIsAvailableReader is a Reader for the GetResourceIsAvailable structure.
type GetResourceIsAvailableReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetResourceIsAvailableReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetResourceIsAvailableOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 401:
		result := NewGetResourceIsAvailableUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewGetResourceIsAvailableNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewGetResourceIsAvailableInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewGetResourceIsAvailableOK creates a GetResourceIsAvailableOK with default headers values
func NewGetResourceIsAvailableOK() *GetResourceIsAvailableOK {
	return &GetResourceIsAvailableOK{}
}

/* GetResourceIsAvailableOK describes a response with status code 200, with default header values.

OK
*/
type GetResourceIsAvailableOK struct {
	Payload *models.ResourceStatus
}

// IsSuccess returns true when this get resource is available o k response has a 2xx status code
func (o *GetResourceIsAvailableOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this get resource is available o k response has a 3xx status code
func (o *GetResourceIsAvailableOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get resource is available o k response has a 4xx status code
func (o *GetResourceIsAvailableOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this get resource is available o k response has a 5xx status code
func (o *GetResourceIsAvailableOK) IsServerError() bool {
	return false
}

// IsCode returns true when this get resource is available o k response a status code equal to that given
func (o *GetResourceIsAvailableOK) IsCode(code int) bool {
	return code == 200
}

func (o *GetResourceIsAvailableOK) Error() string {
	return fmt.Sprintf("[GET /admin/resources/{resource_name}][%d] getResourceIsAvailableOK  %+v", 200, o.Payload)
}

func (o *GetResourceIsAvailableOK) String() string {
	return fmt.Sprintf("[GET /admin/resources/{resource_name}][%d] getResourceIsAvailableOK  %+v", 200, o.Payload)
}

func (o *GetResourceIsAvailableOK) GetPayload() *models.ResourceStatus {
	return o.Payload
}

func (o *GetResourceIsAvailableOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ResourceStatus)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetResourceIsAvailableUnauthorized creates a GetResourceIsAvailableUnauthorized with default headers values
func NewGetResourceIsAvailableUnauthorized() *GetResourceIsAvailableUnauthorized {
	return &GetResourceIsAvailableUnauthorized{}
}

/* GetResourceIsAvailableUnauthorized describes a response with status code 401, with default header values.

Unauthorized
*/
type GetResourceIsAvailableUnauthorized struct {
	Payload *models.Error
}

// IsSuccess returns true when this get resource is available unauthorized response has a 2xx status code
func (o *GetResourceIsAvailableUnauthorized) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get resource is available unauthorized response has a 3xx status code
func (o *GetResourceIsAvailableUnauthorized) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get resource is available unauthorized response has a 4xx status code
func (o *GetResourceIsAvailableUnauthorized) IsClientError() bool {
	return true
}

// IsServerError returns true when this get resource is available unauthorized response has a 5xx status code
func (o *GetResourceIsAvailableUnauthorized) IsServerError() bool {
	return false
}

// IsCode returns true when this get resource is available unauthorized response a status code equal to that given
func (o *GetResourceIsAvailableUnauthorized) IsCode(code int) bool {
	return code == 401
}

func (o *GetResourceIsAvailableUnauthorized) Error() string {
	return fmt.Sprintf("[GET /admin/resources/{resource_name}][%d] getResourceIsAvailableUnauthorized  %+v", 401, o.Payload)
}

func (o *GetResourceIsAvailableUnauthorized) String() string {
	return fmt.Sprintf("[GET /admin/resources/{resource_name}][%d] getResourceIsAvailableUnauthorized  %+v", 401, o.Payload)
}

func (o *GetResourceIsAvailableUnauthorized) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetResourceIsAvailableUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetResourceIsAvailableNotFound creates a GetResourceIsAvailableNotFound with default headers values
func NewGetResourceIsAvailableNotFound() *GetResourceIsAvailableNotFound {
	return &GetResourceIsAvailableNotFound{}
}

/* GetResourceIsAvailableNotFound describes a response with status code 404, with default header values.

The specified resource was not found
*/
type GetResourceIsAvailableNotFound struct {
	Payload *models.Error
}

// IsSuccess returns true when this get resource is available not found response has a 2xx status code
func (o *GetResourceIsAvailableNotFound) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get resource is available not found response has a 3xx status code
func (o *GetResourceIsAvailableNotFound) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get resource is available not found response has a 4xx status code
func (o *GetResourceIsAvailableNotFound) IsClientError() bool {
	return true
}

// IsServerError returns true when this get resource is available not found response has a 5xx status code
func (o *GetResourceIsAvailableNotFound) IsServerError() bool {
	return false
}

// IsCode returns true when this get resource is available not found response a status code equal to that given
func (o *GetResourceIsAvailableNotFound) IsCode(code int) bool {
	return code == 404
}

func (o *GetResourceIsAvailableNotFound) Error() string {
	return fmt.Sprintf("[GET /admin/resources/{resource_name}][%d] getResourceIsAvailableNotFound  %+v", 404, o.Payload)
}

func (o *GetResourceIsAvailableNotFound) String() string {
	return fmt.Sprintf("[GET /admin/resources/{resource_name}][%d] getResourceIsAvailableNotFound  %+v", 404, o.Payload)
}

func (o *GetResourceIsAvailableNotFound) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetResourceIsAvailableNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetResourceIsAvailableInternalServerError creates a GetResourceIsAvailableInternalServerError with default headers values
func NewGetResourceIsAvailableInternalServerError() *GetResourceIsAvailableInternalServerError {
	return &GetResourceIsAvailableInternalServerError{}
}

/* GetResourceIsAvailableInternalServerError describes a response with status code 500, with default header values.

Internal Error
*/
type GetResourceIsAvailableInternalServerError struct {
	Payload *models.Error
}

// IsSuccess returns true when this get resource is available internal server error response has a 2xx status code
func (o *GetResourceIsAvailableInternalServerError) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get resource is available internal server error response has a 3xx status code
func (o *GetResourceIsAvailableInternalServerError) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get resource is available internal server error response has a 4xx status code
func (o *GetResourceIsAvailableInternalServerError) IsClientError() bool {
	return false
}

// IsServerError returns true when this get resource is available internal server error response has a 5xx status code
func (o *GetResourceIsAvailableInternalServerError) IsServerError() bool {
	return true
}

// IsCode returns true when this get resource is available internal server error response a status code equal to that given
func (o *GetResourceIsAvailableInternalServerError) IsCode(code int) bool {
	return code == 500
}

func (o *GetResourceIsAvailableInternalServerError) Error() string {
	return fmt.Sprintf("[GET /admin/resources/{resource_name}][%d] getResourceIsAvailableInternalServerError  %+v", 500, o.Payload)
}

func (o *GetResourceIsAvailableInternalServerError) String() string {
	return fmt.Sprintf("[GET /admin/resources/{resource_name}][%d] getResourceIsAvailableInternalServerError  %+v", 500, o.Payload)
}

func (o *GetResourceIsAvailableInternalServerError) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetResourceIsAvailableInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
