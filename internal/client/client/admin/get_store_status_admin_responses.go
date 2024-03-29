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

// GetStoreStatusAdminReader is a Reader for the GetStoreStatusAdmin structure.
type GetStoreStatusAdminReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetStoreStatusAdminReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetStoreStatusAdminOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 401:
		result := NewGetStoreStatusAdminUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewGetStoreStatusAdminNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewGetStoreStatusAdminInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewGetStoreStatusAdminOK creates a GetStoreStatusAdminOK with default headers values
func NewGetStoreStatusAdminOK() *GetStoreStatusAdminOK {
	return &GetStoreStatusAdminOK{}
}

/* GetStoreStatusAdminOK describes a response with status code 200, with default header values.

OK
*/
type GetStoreStatusAdminOK struct {
	Payload *models.StoreStatusAdmin
}

// IsSuccess returns true when this get store status admin o k response has a 2xx status code
func (o *GetStoreStatusAdminOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this get store status admin o k response has a 3xx status code
func (o *GetStoreStatusAdminOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get store status admin o k response has a 4xx status code
func (o *GetStoreStatusAdminOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this get store status admin o k response has a 5xx status code
func (o *GetStoreStatusAdminOK) IsServerError() bool {
	return false
}

// IsCode returns true when this get store status admin o k response a status code equal to that given
func (o *GetStoreStatusAdminOK) IsCode(code int) bool {
	return code == 200
}

func (o *GetStoreStatusAdminOK) Error() string {
	return fmt.Sprintf("[GET /admin/status][%d] getStoreStatusAdminOK  %+v", 200, o.Payload)
}

func (o *GetStoreStatusAdminOK) String() string {
	return fmt.Sprintf("[GET /admin/status][%d] getStoreStatusAdminOK  %+v", 200, o.Payload)
}

func (o *GetStoreStatusAdminOK) GetPayload() *models.StoreStatusAdmin {
	return o.Payload
}

func (o *GetStoreStatusAdminOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.StoreStatusAdmin)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetStoreStatusAdminUnauthorized creates a GetStoreStatusAdminUnauthorized with default headers values
func NewGetStoreStatusAdminUnauthorized() *GetStoreStatusAdminUnauthorized {
	return &GetStoreStatusAdminUnauthorized{}
}

/* GetStoreStatusAdminUnauthorized describes a response with status code 401, with default header values.

Unauthorized
*/
type GetStoreStatusAdminUnauthorized struct {
	Payload *models.Error
}

// IsSuccess returns true when this get store status admin unauthorized response has a 2xx status code
func (o *GetStoreStatusAdminUnauthorized) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get store status admin unauthorized response has a 3xx status code
func (o *GetStoreStatusAdminUnauthorized) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get store status admin unauthorized response has a 4xx status code
func (o *GetStoreStatusAdminUnauthorized) IsClientError() bool {
	return true
}

// IsServerError returns true when this get store status admin unauthorized response has a 5xx status code
func (o *GetStoreStatusAdminUnauthorized) IsServerError() bool {
	return false
}

// IsCode returns true when this get store status admin unauthorized response a status code equal to that given
func (o *GetStoreStatusAdminUnauthorized) IsCode(code int) bool {
	return code == 401
}

func (o *GetStoreStatusAdminUnauthorized) Error() string {
	return fmt.Sprintf("[GET /admin/status][%d] getStoreStatusAdminUnauthorized  %+v", 401, o.Payload)
}

func (o *GetStoreStatusAdminUnauthorized) String() string {
	return fmt.Sprintf("[GET /admin/status][%d] getStoreStatusAdminUnauthorized  %+v", 401, o.Payload)
}

func (o *GetStoreStatusAdminUnauthorized) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetStoreStatusAdminUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetStoreStatusAdminNotFound creates a GetStoreStatusAdminNotFound with default headers values
func NewGetStoreStatusAdminNotFound() *GetStoreStatusAdminNotFound {
	return &GetStoreStatusAdminNotFound{}
}

/* GetStoreStatusAdminNotFound describes a response with status code 404, with default header values.

The specified resource was not found
*/
type GetStoreStatusAdminNotFound struct {
	Payload *models.Error
}

// IsSuccess returns true when this get store status admin not found response has a 2xx status code
func (o *GetStoreStatusAdminNotFound) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get store status admin not found response has a 3xx status code
func (o *GetStoreStatusAdminNotFound) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get store status admin not found response has a 4xx status code
func (o *GetStoreStatusAdminNotFound) IsClientError() bool {
	return true
}

// IsServerError returns true when this get store status admin not found response has a 5xx status code
func (o *GetStoreStatusAdminNotFound) IsServerError() bool {
	return false
}

// IsCode returns true when this get store status admin not found response a status code equal to that given
func (o *GetStoreStatusAdminNotFound) IsCode(code int) bool {
	return code == 404
}

func (o *GetStoreStatusAdminNotFound) Error() string {
	return fmt.Sprintf("[GET /admin/status][%d] getStoreStatusAdminNotFound  %+v", 404, o.Payload)
}

func (o *GetStoreStatusAdminNotFound) String() string {
	return fmt.Sprintf("[GET /admin/status][%d] getStoreStatusAdminNotFound  %+v", 404, o.Payload)
}

func (o *GetStoreStatusAdminNotFound) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetStoreStatusAdminNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetStoreStatusAdminInternalServerError creates a GetStoreStatusAdminInternalServerError with default headers values
func NewGetStoreStatusAdminInternalServerError() *GetStoreStatusAdminInternalServerError {
	return &GetStoreStatusAdminInternalServerError{}
}

/* GetStoreStatusAdminInternalServerError describes a response with status code 500, with default header values.

Internal Error
*/
type GetStoreStatusAdminInternalServerError struct {
	Payload *models.Error
}

// IsSuccess returns true when this get store status admin internal server error response has a 2xx status code
func (o *GetStoreStatusAdminInternalServerError) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get store status admin internal server error response has a 3xx status code
func (o *GetStoreStatusAdminInternalServerError) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get store status admin internal server error response has a 4xx status code
func (o *GetStoreStatusAdminInternalServerError) IsClientError() bool {
	return false
}

// IsServerError returns true when this get store status admin internal server error response has a 5xx status code
func (o *GetStoreStatusAdminInternalServerError) IsServerError() bool {
	return true
}

// IsCode returns true when this get store status admin internal server error response a status code equal to that given
func (o *GetStoreStatusAdminInternalServerError) IsCode(code int) bool {
	return code == 500
}

func (o *GetStoreStatusAdminInternalServerError) Error() string {
	return fmt.Sprintf("[GET /admin/status][%d] getStoreStatusAdminInternalServerError  %+v", 500, o.Payload)
}

func (o *GetStoreStatusAdminInternalServerError) String() string {
	return fmt.Sprintf("[GET /admin/status][%d] getStoreStatusAdminInternalServerError  %+v", 500, o.Payload)
}

func (o *GetStoreStatusAdminInternalServerError) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetStoreStatusAdminInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
