// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/practable/book/internal/ac/models"
)

// DenyReader is a Reader for the Deny structure.
type DenyReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *DenyReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 204:
		result := NewDenyNoContent()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewDenyBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 401:
		result := NewDenyUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewDenyNoContent creates a DenyNoContent with default headers values
func NewDenyNoContent() *DenyNoContent {
	return &DenyNoContent{}
}

/* DenyNoContent describes a response with status code 204, with default header values.

The bid was denied successfully.
*/
type DenyNoContent struct {
}

// IsSuccess returns true when this deny no content response has a 2xx status code
func (o *DenyNoContent) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this deny no content response has a 3xx status code
func (o *DenyNoContent) IsRedirect() bool {
	return false
}

// IsClientError returns true when this deny no content response has a 4xx status code
func (o *DenyNoContent) IsClientError() bool {
	return false
}

// IsServerError returns true when this deny no content response has a 5xx status code
func (o *DenyNoContent) IsServerError() bool {
	return false
}

// IsCode returns true when this deny no content response a status code equal to that given
func (o *DenyNoContent) IsCode(code int) bool {
	return code == 204
}

func (o *DenyNoContent) Error() string {
	return fmt.Sprintf("[POST /bids/deny][%d] denyNoContent ", 204)
}

func (o *DenyNoContent) String() string {
	return fmt.Sprintf("[POST /bids/deny][%d] denyNoContent ", 204)
}

func (o *DenyNoContent) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewDenyBadRequest creates a DenyBadRequest with default headers values
func NewDenyBadRequest() *DenyBadRequest {
	return &DenyBadRequest{}
}

/* DenyBadRequest describes a response with status code 400, with default header values.

BadRequest
*/
type DenyBadRequest struct {
	Payload *models.Error
}

// IsSuccess returns true when this deny bad request response has a 2xx status code
func (o *DenyBadRequest) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this deny bad request response has a 3xx status code
func (o *DenyBadRequest) IsRedirect() bool {
	return false
}

// IsClientError returns true when this deny bad request response has a 4xx status code
func (o *DenyBadRequest) IsClientError() bool {
	return true
}

// IsServerError returns true when this deny bad request response has a 5xx status code
func (o *DenyBadRequest) IsServerError() bool {
	return false
}

// IsCode returns true when this deny bad request response a status code equal to that given
func (o *DenyBadRequest) IsCode(code int) bool {
	return code == 400
}

func (o *DenyBadRequest) Error() string {
	return fmt.Sprintf("[POST /bids/deny][%d] denyBadRequest  %+v", 400, o.Payload)
}

func (o *DenyBadRequest) String() string {
	return fmt.Sprintf("[POST /bids/deny][%d] denyBadRequest  %+v", 400, o.Payload)
}

func (o *DenyBadRequest) GetPayload() *models.Error {
	return o.Payload
}

func (o *DenyBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewDenyUnauthorized creates a DenyUnauthorized with default headers values
func NewDenyUnauthorized() *DenyUnauthorized {
	return &DenyUnauthorized{}
}

/* DenyUnauthorized describes a response with status code 401, with default header values.

Unauthorized
*/
type DenyUnauthorized struct {
	Payload *models.Error
}

// IsSuccess returns true when this deny unauthorized response has a 2xx status code
func (o *DenyUnauthorized) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this deny unauthorized response has a 3xx status code
func (o *DenyUnauthorized) IsRedirect() bool {
	return false
}

// IsClientError returns true when this deny unauthorized response has a 4xx status code
func (o *DenyUnauthorized) IsClientError() bool {
	return true
}

// IsServerError returns true when this deny unauthorized response has a 5xx status code
func (o *DenyUnauthorized) IsServerError() bool {
	return false
}

// IsCode returns true when this deny unauthorized response a status code equal to that given
func (o *DenyUnauthorized) IsCode(code int) bool {
	return code == 401
}

func (o *DenyUnauthorized) Error() string {
	return fmt.Sprintf("[POST /bids/deny][%d] denyUnauthorized  %+v", 401, o.Payload)
}

func (o *DenyUnauthorized) String() string {
	return fmt.Sprintf("[POST /bids/deny][%d] denyUnauthorized  %+v", 401, o.Payload)
}

func (o *DenyUnauthorized) GetPayload() *models.Error {
	return o.Payload
}

func (o *DenyUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
