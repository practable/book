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

// GetPolicyReader is a Reader for the GetPolicy structure.
type GetPolicyReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetPolicyReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetPolicyOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 401:
		result := NewGetPolicyUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewGetPolicyNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewGetPolicyInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewGetPolicyOK creates a GetPolicyOK with default headers values
func NewGetPolicyOK() *GetPolicyOK {
	return &GetPolicyOK{}
}

/*
GetPolicyOK describes a response with status code 200, with default header values.

OK
*/
type GetPolicyOK struct {
	Payload *models.PolicyDescribed
}

// IsSuccess returns true when this get policy o k response has a 2xx status code
func (o *GetPolicyOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this get policy o k response has a 3xx status code
func (o *GetPolicyOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get policy o k response has a 4xx status code
func (o *GetPolicyOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this get policy o k response has a 5xx status code
func (o *GetPolicyOK) IsServerError() bool {
	return false
}

// IsCode returns true when this get policy o k response a status code equal to that given
func (o *GetPolicyOK) IsCode(code int) bool {
	return code == 200
}

func (o *GetPolicyOK) Error() string {
	return fmt.Sprintf("[GET /policies/{policy_name}][%d] getPolicyOK  %+v", 200, o.Payload)
}

func (o *GetPolicyOK) String() string {
	return fmt.Sprintf("[GET /policies/{policy_name}][%d] getPolicyOK  %+v", 200, o.Payload)
}

func (o *GetPolicyOK) GetPayload() *models.PolicyDescribed {
	return o.Payload
}

func (o *GetPolicyOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.PolicyDescribed)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetPolicyUnauthorized creates a GetPolicyUnauthorized with default headers values
func NewGetPolicyUnauthorized() *GetPolicyUnauthorized {
	return &GetPolicyUnauthorized{}
}

/*
GetPolicyUnauthorized describes a response with status code 401, with default header values.

Unauthorized
*/
type GetPolicyUnauthorized struct {
	Payload *models.Error
}

// IsSuccess returns true when this get policy unauthorized response has a 2xx status code
func (o *GetPolicyUnauthorized) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get policy unauthorized response has a 3xx status code
func (o *GetPolicyUnauthorized) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get policy unauthorized response has a 4xx status code
func (o *GetPolicyUnauthorized) IsClientError() bool {
	return true
}

// IsServerError returns true when this get policy unauthorized response has a 5xx status code
func (o *GetPolicyUnauthorized) IsServerError() bool {
	return false
}

// IsCode returns true when this get policy unauthorized response a status code equal to that given
func (o *GetPolicyUnauthorized) IsCode(code int) bool {
	return code == 401
}

func (o *GetPolicyUnauthorized) Error() string {
	return fmt.Sprintf("[GET /policies/{policy_name}][%d] getPolicyUnauthorized  %+v", 401, o.Payload)
}

func (o *GetPolicyUnauthorized) String() string {
	return fmt.Sprintf("[GET /policies/{policy_name}][%d] getPolicyUnauthorized  %+v", 401, o.Payload)
}

func (o *GetPolicyUnauthorized) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetPolicyUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetPolicyNotFound creates a GetPolicyNotFound with default headers values
func NewGetPolicyNotFound() *GetPolicyNotFound {
	return &GetPolicyNotFound{}
}

/*
GetPolicyNotFound describes a response with status code 404, with default header values.

The specified resource was not found
*/
type GetPolicyNotFound struct {
	Payload *models.Error
}

// IsSuccess returns true when this get policy not found response has a 2xx status code
func (o *GetPolicyNotFound) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get policy not found response has a 3xx status code
func (o *GetPolicyNotFound) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get policy not found response has a 4xx status code
func (o *GetPolicyNotFound) IsClientError() bool {
	return true
}

// IsServerError returns true when this get policy not found response has a 5xx status code
func (o *GetPolicyNotFound) IsServerError() bool {
	return false
}

// IsCode returns true when this get policy not found response a status code equal to that given
func (o *GetPolicyNotFound) IsCode(code int) bool {
	return code == 404
}

func (o *GetPolicyNotFound) Error() string {
	return fmt.Sprintf("[GET /policies/{policy_name}][%d] getPolicyNotFound  %+v", 404, o.Payload)
}

func (o *GetPolicyNotFound) String() string {
	return fmt.Sprintf("[GET /policies/{policy_name}][%d] getPolicyNotFound  %+v", 404, o.Payload)
}

func (o *GetPolicyNotFound) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetPolicyNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetPolicyInternalServerError creates a GetPolicyInternalServerError with default headers values
func NewGetPolicyInternalServerError() *GetPolicyInternalServerError {
	return &GetPolicyInternalServerError{}
}

/*
GetPolicyInternalServerError describes a response with status code 500, with default header values.

Internal Error
*/
type GetPolicyInternalServerError struct {
	Payload *models.Error
}

// IsSuccess returns true when this get policy internal server error response has a 2xx status code
func (o *GetPolicyInternalServerError) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get policy internal server error response has a 3xx status code
func (o *GetPolicyInternalServerError) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get policy internal server error response has a 4xx status code
func (o *GetPolicyInternalServerError) IsClientError() bool {
	return false
}

// IsServerError returns true when this get policy internal server error response has a 5xx status code
func (o *GetPolicyInternalServerError) IsServerError() bool {
	return true
}

// IsCode returns true when this get policy internal server error response a status code equal to that given
func (o *GetPolicyInternalServerError) IsCode(code int) bool {
	return code == 500
}

func (o *GetPolicyInternalServerError) Error() string {
	return fmt.Sprintf("[GET /policies/{policy_name}][%d] getPolicyInternalServerError  %+v", 500, o.Payload)
}

func (o *GetPolicyInternalServerError) String() string {
	return fmt.Sprintf("[GET /policies/{policy_name}][%d] getPolicyInternalServerError  %+v", 500, o.Payload)
}

func (o *GetPolicyInternalServerError) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetPolicyInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
