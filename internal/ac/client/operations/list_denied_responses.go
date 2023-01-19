// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/timdrysdale/interval/internal/ac/models"
)

// ListDeniedReader is a Reader for the ListDenied structure.
type ListDeniedReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *ListDeniedReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewListDeniedOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 401:
		result := NewListDeniedUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewListDeniedOK creates a ListDeniedOK with default headers values
func NewListDeniedOK() *ListDeniedOK {
	return &ListDeniedOK{}
}

/* ListDeniedOK describes a response with status code 200, with default header values.

List of current denied bids
*/
type ListDeniedOK struct {
	Payload *models.BookingIDs
}

// IsSuccess returns true when this list denied o k response has a 2xx status code
func (o *ListDeniedOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this list denied o k response has a 3xx status code
func (o *ListDeniedOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this list denied o k response has a 4xx status code
func (o *ListDeniedOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this list denied o k response has a 5xx status code
func (o *ListDeniedOK) IsServerError() bool {
	return false
}

// IsCode returns true when this list denied o k response a status code equal to that given
func (o *ListDeniedOK) IsCode(code int) bool {
	return code == 200
}

func (o *ListDeniedOK) Error() string {
	return fmt.Sprintf("[GET /bids/deny][%d] listDeniedOK  %+v", 200, o.Payload)
}

func (o *ListDeniedOK) String() string {
	return fmt.Sprintf("[GET /bids/deny][%d] listDeniedOK  %+v", 200, o.Payload)
}

func (o *ListDeniedOK) GetPayload() *models.BookingIDs {
	return o.Payload
}

func (o *ListDeniedOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.BookingIDs)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewListDeniedUnauthorized creates a ListDeniedUnauthorized with default headers values
func NewListDeniedUnauthorized() *ListDeniedUnauthorized {
	return &ListDeniedUnauthorized{}
}

/* ListDeniedUnauthorized describes a response with status code 401, with default header values.

Unauthorized
*/
type ListDeniedUnauthorized struct {
	Payload *models.Error
}

// IsSuccess returns true when this list denied unauthorized response has a 2xx status code
func (o *ListDeniedUnauthorized) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this list denied unauthorized response has a 3xx status code
func (o *ListDeniedUnauthorized) IsRedirect() bool {
	return false
}

// IsClientError returns true when this list denied unauthorized response has a 4xx status code
func (o *ListDeniedUnauthorized) IsClientError() bool {
	return true
}

// IsServerError returns true when this list denied unauthorized response has a 5xx status code
func (o *ListDeniedUnauthorized) IsServerError() bool {
	return false
}

// IsCode returns true when this list denied unauthorized response a status code equal to that given
func (o *ListDeniedUnauthorized) IsCode(code int) bool {
	return code == 401
}

func (o *ListDeniedUnauthorized) Error() string {
	return fmt.Sprintf("[GET /bids/deny][%d] listDeniedUnauthorized  %+v", 401, o.Payload)
}

func (o *ListDeniedUnauthorized) String() string {
	return fmt.Sprintf("[GET /bids/deny][%d] listDeniedUnauthorized  %+v", 401, o.Payload)
}

func (o *ListDeniedUnauthorized) GetPayload() *models.Error {
	return o.Payload
}

func (o *ListDeniedUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}