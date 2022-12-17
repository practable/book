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

// GetStoreStatusUserReader is a Reader for the GetStoreStatusUser structure.
type GetStoreStatusUserReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetStoreStatusUserReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetStoreStatusUserOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 401:
		result := NewGetStoreStatusUserUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewGetStoreStatusUserNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewGetStoreStatusUserInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewGetStoreStatusUserOK creates a GetStoreStatusUserOK with default headers values
func NewGetStoreStatusUserOK() *GetStoreStatusUserOK {
	return &GetStoreStatusUserOK{}
}

/*
GetStoreStatusUserOK describes a response with status code 200, with default header values.

OK
*/
type GetStoreStatusUserOK struct {
	Payload *models.StoreStatusUser
}

// IsSuccess returns true when this get store status user o k response has a 2xx status code
func (o *GetStoreStatusUserOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this get store status user o k response has a 3xx status code
func (o *GetStoreStatusUserOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get store status user o k response has a 4xx status code
func (o *GetStoreStatusUserOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this get store status user o k response has a 5xx status code
func (o *GetStoreStatusUserOK) IsServerError() bool {
	return false
}

// IsCode returns true when this get store status user o k response a status code equal to that given
func (o *GetStoreStatusUserOK) IsCode(code int) bool {
	return code == 200
}

func (o *GetStoreStatusUserOK) Error() string {
	return fmt.Sprintf("[GET /users/status][%d] getStoreStatusUserOK  %+v", 200, o.Payload)
}

func (o *GetStoreStatusUserOK) String() string {
	return fmt.Sprintf("[GET /users/status][%d] getStoreStatusUserOK  %+v", 200, o.Payload)
}

func (o *GetStoreStatusUserOK) GetPayload() *models.StoreStatusUser {
	return o.Payload
}

func (o *GetStoreStatusUserOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.StoreStatusUser)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetStoreStatusUserUnauthorized creates a GetStoreStatusUserUnauthorized with default headers values
func NewGetStoreStatusUserUnauthorized() *GetStoreStatusUserUnauthorized {
	return &GetStoreStatusUserUnauthorized{}
}

/*
GetStoreStatusUserUnauthorized describes a response with status code 401, with default header values.

Unauthorized
*/
type GetStoreStatusUserUnauthorized struct {
	Payload *models.Error
}

// IsSuccess returns true when this get store status user unauthorized response has a 2xx status code
func (o *GetStoreStatusUserUnauthorized) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get store status user unauthorized response has a 3xx status code
func (o *GetStoreStatusUserUnauthorized) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get store status user unauthorized response has a 4xx status code
func (o *GetStoreStatusUserUnauthorized) IsClientError() bool {
	return true
}

// IsServerError returns true when this get store status user unauthorized response has a 5xx status code
func (o *GetStoreStatusUserUnauthorized) IsServerError() bool {
	return false
}

// IsCode returns true when this get store status user unauthorized response a status code equal to that given
func (o *GetStoreStatusUserUnauthorized) IsCode(code int) bool {
	return code == 401
}

func (o *GetStoreStatusUserUnauthorized) Error() string {
	return fmt.Sprintf("[GET /users/status][%d] getStoreStatusUserUnauthorized  %+v", 401, o.Payload)
}

func (o *GetStoreStatusUserUnauthorized) String() string {
	return fmt.Sprintf("[GET /users/status][%d] getStoreStatusUserUnauthorized  %+v", 401, o.Payload)
}

func (o *GetStoreStatusUserUnauthorized) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetStoreStatusUserUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetStoreStatusUserNotFound creates a GetStoreStatusUserNotFound with default headers values
func NewGetStoreStatusUserNotFound() *GetStoreStatusUserNotFound {
	return &GetStoreStatusUserNotFound{}
}

/*
GetStoreStatusUserNotFound describes a response with status code 404, with default header values.

The specified resource was not found
*/
type GetStoreStatusUserNotFound struct {
	Payload *models.Error
}

// IsSuccess returns true when this get store status user not found response has a 2xx status code
func (o *GetStoreStatusUserNotFound) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get store status user not found response has a 3xx status code
func (o *GetStoreStatusUserNotFound) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get store status user not found response has a 4xx status code
func (o *GetStoreStatusUserNotFound) IsClientError() bool {
	return true
}

// IsServerError returns true when this get store status user not found response has a 5xx status code
func (o *GetStoreStatusUserNotFound) IsServerError() bool {
	return false
}

// IsCode returns true when this get store status user not found response a status code equal to that given
func (o *GetStoreStatusUserNotFound) IsCode(code int) bool {
	return code == 404
}

func (o *GetStoreStatusUserNotFound) Error() string {
	return fmt.Sprintf("[GET /users/status][%d] getStoreStatusUserNotFound  %+v", 404, o.Payload)
}

func (o *GetStoreStatusUserNotFound) String() string {
	return fmt.Sprintf("[GET /users/status][%d] getStoreStatusUserNotFound  %+v", 404, o.Payload)
}

func (o *GetStoreStatusUserNotFound) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetStoreStatusUserNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetStoreStatusUserInternalServerError creates a GetStoreStatusUserInternalServerError with default headers values
func NewGetStoreStatusUserInternalServerError() *GetStoreStatusUserInternalServerError {
	return &GetStoreStatusUserInternalServerError{}
}

/*
GetStoreStatusUserInternalServerError describes a response with status code 500, with default header values.

Internal Error
*/
type GetStoreStatusUserInternalServerError struct {
	Payload *models.Error
}

// IsSuccess returns true when this get store status user internal server error response has a 2xx status code
func (o *GetStoreStatusUserInternalServerError) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get store status user internal server error response has a 3xx status code
func (o *GetStoreStatusUserInternalServerError) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get store status user internal server error response has a 4xx status code
func (o *GetStoreStatusUserInternalServerError) IsClientError() bool {
	return false
}

// IsServerError returns true when this get store status user internal server error response has a 5xx status code
func (o *GetStoreStatusUserInternalServerError) IsServerError() bool {
	return true
}

// IsCode returns true when this get store status user internal server error response a status code equal to that given
func (o *GetStoreStatusUserInternalServerError) IsCode(code int) bool {
	return code == 500
}

func (o *GetStoreStatusUserInternalServerError) Error() string {
	return fmt.Sprintf("[GET /users/status][%d] getStoreStatusUserInternalServerError  %+v", 500, o.Payload)
}

func (o *GetStoreStatusUserInternalServerError) String() string {
	return fmt.Sprintf("[GET /users/status][%d] getStoreStatusUserInternalServerError  %+v", 500, o.Payload)
}

func (o *GetStoreStatusUserInternalServerError) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetStoreStatusUserInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}