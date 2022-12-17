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

// GetPolicyStatusForUserReader is a Reader for the GetPolicyStatusForUser structure.
type GetPolicyStatusForUserReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetPolicyStatusForUserReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetPolicyStatusForUserOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 401:
		result := NewGetPolicyStatusForUserUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewGetPolicyStatusForUserNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewGetPolicyStatusForUserInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewGetPolicyStatusForUserOK creates a GetPolicyStatusForUserOK with default headers values
func NewGetPolicyStatusForUserOK() *GetPolicyStatusForUserOK {
	return &GetPolicyStatusForUserOK{}
}

/*
GetPolicyStatusForUserOK describes a response with status code 200, with default header values.

OK
*/
type GetPolicyStatusForUserOK struct {
	Payload *models.PolicyStatus
}

// IsSuccess returns true when this get policy status for user o k response has a 2xx status code
func (o *GetPolicyStatusForUserOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this get policy status for user o k response has a 3xx status code
func (o *GetPolicyStatusForUserOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get policy status for user o k response has a 4xx status code
func (o *GetPolicyStatusForUserOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this get policy status for user o k response has a 5xx status code
func (o *GetPolicyStatusForUserOK) IsServerError() bool {
	return false
}

// IsCode returns true when this get policy status for user o k response a status code equal to that given
func (o *GetPolicyStatusForUserOK) IsCode(code int) bool {
	return code == 200
}

func (o *GetPolicyStatusForUserOK) Error() string {
	return fmt.Sprintf("[GET /users/{user_name}/policies/{policy_name}][%d] getPolicyStatusForUserOK  %+v", 200, o.Payload)
}

func (o *GetPolicyStatusForUserOK) String() string {
	return fmt.Sprintf("[GET /users/{user_name}/policies/{policy_name}][%d] getPolicyStatusForUserOK  %+v", 200, o.Payload)
}

func (o *GetPolicyStatusForUserOK) GetPayload() *models.PolicyStatus {
	return o.Payload
}

func (o *GetPolicyStatusForUserOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.PolicyStatus)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetPolicyStatusForUserUnauthorized creates a GetPolicyStatusForUserUnauthorized with default headers values
func NewGetPolicyStatusForUserUnauthorized() *GetPolicyStatusForUserUnauthorized {
	return &GetPolicyStatusForUserUnauthorized{}
}

/*
GetPolicyStatusForUserUnauthorized describes a response with status code 401, with default header values.

Unauthorized
*/
type GetPolicyStatusForUserUnauthorized struct {
	Payload *models.Error
}

// IsSuccess returns true when this get policy status for user unauthorized response has a 2xx status code
func (o *GetPolicyStatusForUserUnauthorized) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get policy status for user unauthorized response has a 3xx status code
func (o *GetPolicyStatusForUserUnauthorized) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get policy status for user unauthorized response has a 4xx status code
func (o *GetPolicyStatusForUserUnauthorized) IsClientError() bool {
	return true
}

// IsServerError returns true when this get policy status for user unauthorized response has a 5xx status code
func (o *GetPolicyStatusForUserUnauthorized) IsServerError() bool {
	return false
}

// IsCode returns true when this get policy status for user unauthorized response a status code equal to that given
func (o *GetPolicyStatusForUserUnauthorized) IsCode(code int) bool {
	return code == 401
}

func (o *GetPolicyStatusForUserUnauthorized) Error() string {
	return fmt.Sprintf("[GET /users/{user_name}/policies/{policy_name}][%d] getPolicyStatusForUserUnauthorized  %+v", 401, o.Payload)
}

func (o *GetPolicyStatusForUserUnauthorized) String() string {
	return fmt.Sprintf("[GET /users/{user_name}/policies/{policy_name}][%d] getPolicyStatusForUserUnauthorized  %+v", 401, o.Payload)
}

func (o *GetPolicyStatusForUserUnauthorized) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetPolicyStatusForUserUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetPolicyStatusForUserNotFound creates a GetPolicyStatusForUserNotFound with default headers values
func NewGetPolicyStatusForUserNotFound() *GetPolicyStatusForUserNotFound {
	return &GetPolicyStatusForUserNotFound{}
}

/*
GetPolicyStatusForUserNotFound describes a response with status code 404, with default header values.

The specified resource was not found
*/
type GetPolicyStatusForUserNotFound struct {
	Payload *models.Error
}

// IsSuccess returns true when this get policy status for user not found response has a 2xx status code
func (o *GetPolicyStatusForUserNotFound) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get policy status for user not found response has a 3xx status code
func (o *GetPolicyStatusForUserNotFound) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get policy status for user not found response has a 4xx status code
func (o *GetPolicyStatusForUserNotFound) IsClientError() bool {
	return true
}

// IsServerError returns true when this get policy status for user not found response has a 5xx status code
func (o *GetPolicyStatusForUserNotFound) IsServerError() bool {
	return false
}

// IsCode returns true when this get policy status for user not found response a status code equal to that given
func (o *GetPolicyStatusForUserNotFound) IsCode(code int) bool {
	return code == 404
}

func (o *GetPolicyStatusForUserNotFound) Error() string {
	return fmt.Sprintf("[GET /users/{user_name}/policies/{policy_name}][%d] getPolicyStatusForUserNotFound  %+v", 404, o.Payload)
}

func (o *GetPolicyStatusForUserNotFound) String() string {
	return fmt.Sprintf("[GET /users/{user_name}/policies/{policy_name}][%d] getPolicyStatusForUserNotFound  %+v", 404, o.Payload)
}

func (o *GetPolicyStatusForUserNotFound) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetPolicyStatusForUserNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetPolicyStatusForUserInternalServerError creates a GetPolicyStatusForUserInternalServerError with default headers values
func NewGetPolicyStatusForUserInternalServerError() *GetPolicyStatusForUserInternalServerError {
	return &GetPolicyStatusForUserInternalServerError{}
}

/*
GetPolicyStatusForUserInternalServerError describes a response with status code 500, with default header values.

Internal Error
*/
type GetPolicyStatusForUserInternalServerError struct {
	Payload *models.Error
}

// IsSuccess returns true when this get policy status for user internal server error response has a 2xx status code
func (o *GetPolicyStatusForUserInternalServerError) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get policy status for user internal server error response has a 3xx status code
func (o *GetPolicyStatusForUserInternalServerError) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get policy status for user internal server error response has a 4xx status code
func (o *GetPolicyStatusForUserInternalServerError) IsClientError() bool {
	return false
}

// IsServerError returns true when this get policy status for user internal server error response has a 5xx status code
func (o *GetPolicyStatusForUserInternalServerError) IsServerError() bool {
	return true
}

// IsCode returns true when this get policy status for user internal server error response a status code equal to that given
func (o *GetPolicyStatusForUserInternalServerError) IsCode(code int) bool {
	return code == 500
}

func (o *GetPolicyStatusForUserInternalServerError) Error() string {
	return fmt.Sprintf("[GET /users/{user_name}/policies/{policy_name}][%d] getPolicyStatusForUserInternalServerError  %+v", 500, o.Payload)
}

func (o *GetPolicyStatusForUserInternalServerError) String() string {
	return fmt.Sprintf("[GET /users/{user_name}/policies/{policy_name}][%d] getPolicyStatusForUserInternalServerError  %+v", 500, o.Payload)
}

func (o *GetPolicyStatusForUserInternalServerError) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetPolicyStatusForUserInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
