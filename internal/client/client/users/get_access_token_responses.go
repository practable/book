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

// GetAccessTokenReader is a Reader for the GetAccessToken structure.
type GetAccessTokenReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetAccessTokenReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetAccessTokenOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 401:
		result := NewGetAccessTokenUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewGetAccessTokenNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewGetAccessTokenInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewGetAccessTokenOK creates a GetAccessTokenOK with default headers values
func NewGetAccessTokenOK() *GetAccessTokenOK {
	return &GetAccessTokenOK{}
}

/*
GetAccessTokenOK describes a response with status code 200, with default header values.

OK
*/
type GetAccessTokenOK struct {
	Payload *models.AccessToken
}

// IsSuccess returns true when this get access token o k response has a 2xx status code
func (o *GetAccessTokenOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this get access token o k response has a 3xx status code
func (o *GetAccessTokenOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get access token o k response has a 4xx status code
func (o *GetAccessTokenOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this get access token o k response has a 5xx status code
func (o *GetAccessTokenOK) IsServerError() bool {
	return false
}

// IsCode returns true when this get access token o k response a status code equal to that given
func (o *GetAccessTokenOK) IsCode(code int) bool {
	return code == 200
}

func (o *GetAccessTokenOK) Error() string {
	return fmt.Sprintf("[POST /login/{user_name}][%d] getAccessTokenOK  %+v", 200, o.Payload)
}

func (o *GetAccessTokenOK) String() string {
	return fmt.Sprintf("[POST /login/{user_name}][%d] getAccessTokenOK  %+v", 200, o.Payload)
}

func (o *GetAccessTokenOK) GetPayload() *models.AccessToken {
	return o.Payload
}

func (o *GetAccessTokenOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.AccessToken)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetAccessTokenUnauthorized creates a GetAccessTokenUnauthorized with default headers values
func NewGetAccessTokenUnauthorized() *GetAccessTokenUnauthorized {
	return &GetAccessTokenUnauthorized{}
}

/*
GetAccessTokenUnauthorized describes a response with status code 401, with default header values.

Unauthorized
*/
type GetAccessTokenUnauthorized struct {
	Payload *models.Error
}

// IsSuccess returns true when this get access token unauthorized response has a 2xx status code
func (o *GetAccessTokenUnauthorized) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get access token unauthorized response has a 3xx status code
func (o *GetAccessTokenUnauthorized) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get access token unauthorized response has a 4xx status code
func (o *GetAccessTokenUnauthorized) IsClientError() bool {
	return true
}

// IsServerError returns true when this get access token unauthorized response has a 5xx status code
func (o *GetAccessTokenUnauthorized) IsServerError() bool {
	return false
}

// IsCode returns true when this get access token unauthorized response a status code equal to that given
func (o *GetAccessTokenUnauthorized) IsCode(code int) bool {
	return code == 401
}

func (o *GetAccessTokenUnauthorized) Error() string {
	return fmt.Sprintf("[POST /login/{user_name}][%d] getAccessTokenUnauthorized  %+v", 401, o.Payload)
}

func (o *GetAccessTokenUnauthorized) String() string {
	return fmt.Sprintf("[POST /login/{user_name}][%d] getAccessTokenUnauthorized  %+v", 401, o.Payload)
}

func (o *GetAccessTokenUnauthorized) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetAccessTokenUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetAccessTokenNotFound creates a GetAccessTokenNotFound with default headers values
func NewGetAccessTokenNotFound() *GetAccessTokenNotFound {
	return &GetAccessTokenNotFound{}
}

/*
GetAccessTokenNotFound describes a response with status code 404, with default header values.

The specified resource was not found
*/
type GetAccessTokenNotFound struct {
	Payload *models.Error
}

// IsSuccess returns true when this get access token not found response has a 2xx status code
func (o *GetAccessTokenNotFound) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get access token not found response has a 3xx status code
func (o *GetAccessTokenNotFound) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get access token not found response has a 4xx status code
func (o *GetAccessTokenNotFound) IsClientError() bool {
	return true
}

// IsServerError returns true when this get access token not found response has a 5xx status code
func (o *GetAccessTokenNotFound) IsServerError() bool {
	return false
}

// IsCode returns true when this get access token not found response a status code equal to that given
func (o *GetAccessTokenNotFound) IsCode(code int) bool {
	return code == 404
}

func (o *GetAccessTokenNotFound) Error() string {
	return fmt.Sprintf("[POST /login/{user_name}][%d] getAccessTokenNotFound  %+v", 404, o.Payload)
}

func (o *GetAccessTokenNotFound) String() string {
	return fmt.Sprintf("[POST /login/{user_name}][%d] getAccessTokenNotFound  %+v", 404, o.Payload)
}

func (o *GetAccessTokenNotFound) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetAccessTokenNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetAccessTokenInternalServerError creates a GetAccessTokenInternalServerError with default headers values
func NewGetAccessTokenInternalServerError() *GetAccessTokenInternalServerError {
	return &GetAccessTokenInternalServerError{}
}

/*
GetAccessTokenInternalServerError describes a response with status code 500, with default header values.

Internal Error
*/
type GetAccessTokenInternalServerError struct {
	Payload *models.Error
}

// IsSuccess returns true when this get access token internal server error response has a 2xx status code
func (o *GetAccessTokenInternalServerError) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get access token internal server error response has a 3xx status code
func (o *GetAccessTokenInternalServerError) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get access token internal server error response has a 4xx status code
func (o *GetAccessTokenInternalServerError) IsClientError() bool {
	return false
}

// IsServerError returns true when this get access token internal server error response has a 5xx status code
func (o *GetAccessTokenInternalServerError) IsServerError() bool {
	return true
}

// IsCode returns true when this get access token internal server error response a status code equal to that given
func (o *GetAccessTokenInternalServerError) IsCode(code int) bool {
	return code == 500
}

func (o *GetAccessTokenInternalServerError) Error() string {
	return fmt.Sprintf("[POST /login/{user_name}][%d] getAccessTokenInternalServerError  %+v", 500, o.Payload)
}

func (o *GetAccessTokenInternalServerError) String() string {
	return fmt.Sprintf("[POST /login/{user_name}][%d] getAccessTokenInternalServerError  %+v", 500, o.Payload)
}

func (o *GetAccessTokenInternalServerError) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetAccessTokenInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
