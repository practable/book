// Code generated by go-swagger; DO NOT EDIT.

package admin

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/timdrysdale/interval/internal/client/models"
)

// ExportManifestReader is a Reader for the ExportManifest structure.
type ExportManifestReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *ExportManifestReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewExportManifestOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 401:
		result := NewExportManifestUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewExportManifestNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewExportManifestInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewExportManifestOK creates a ExportManifestOK with default headers values
func NewExportManifestOK() *ExportManifestOK {
	return &ExportManifestOK{}
}

/* ExportManifestOK describes a response with status code 200, with default header values.

OK
*/
type ExportManifestOK struct {
	Payload *models.Manifest
}

// IsSuccess returns true when this export manifest o k response has a 2xx status code
func (o *ExportManifestOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this export manifest o k response has a 3xx status code
func (o *ExportManifestOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this export manifest o k response has a 4xx status code
func (o *ExportManifestOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this export manifest o k response has a 5xx status code
func (o *ExportManifestOK) IsServerError() bool {
	return false
}

// IsCode returns true when this export manifest o k response a status code equal to that given
func (o *ExportManifestOK) IsCode(code int) bool {
	return code == 200
}

func (o *ExportManifestOK) Error() string {
	return fmt.Sprintf("[GET /admin/manifest][%d] exportManifestOK  %+v", 200, o.Payload)
}

func (o *ExportManifestOK) String() string {
	return fmt.Sprintf("[GET /admin/manifest][%d] exportManifestOK  %+v", 200, o.Payload)
}

func (o *ExportManifestOK) GetPayload() *models.Manifest {
	return o.Payload
}

func (o *ExportManifestOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Manifest)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewExportManifestUnauthorized creates a ExportManifestUnauthorized with default headers values
func NewExportManifestUnauthorized() *ExportManifestUnauthorized {
	return &ExportManifestUnauthorized{}
}

/* ExportManifestUnauthorized describes a response with status code 401, with default header values.

Unauthorized
*/
type ExportManifestUnauthorized struct {
	Payload *models.Error
}

// IsSuccess returns true when this export manifest unauthorized response has a 2xx status code
func (o *ExportManifestUnauthorized) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this export manifest unauthorized response has a 3xx status code
func (o *ExportManifestUnauthorized) IsRedirect() bool {
	return false
}

// IsClientError returns true when this export manifest unauthorized response has a 4xx status code
func (o *ExportManifestUnauthorized) IsClientError() bool {
	return true
}

// IsServerError returns true when this export manifest unauthorized response has a 5xx status code
func (o *ExportManifestUnauthorized) IsServerError() bool {
	return false
}

// IsCode returns true when this export manifest unauthorized response a status code equal to that given
func (o *ExportManifestUnauthorized) IsCode(code int) bool {
	return code == 401
}

func (o *ExportManifestUnauthorized) Error() string {
	return fmt.Sprintf("[GET /admin/manifest][%d] exportManifestUnauthorized  %+v", 401, o.Payload)
}

func (o *ExportManifestUnauthorized) String() string {
	return fmt.Sprintf("[GET /admin/manifest][%d] exportManifestUnauthorized  %+v", 401, o.Payload)
}

func (o *ExportManifestUnauthorized) GetPayload() *models.Error {
	return o.Payload
}

func (o *ExportManifestUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewExportManifestNotFound creates a ExportManifestNotFound with default headers values
func NewExportManifestNotFound() *ExportManifestNotFound {
	return &ExportManifestNotFound{}
}

/* ExportManifestNotFound describes a response with status code 404, with default header values.

The specified resource was not found
*/
type ExportManifestNotFound struct {
	Payload *models.Error
}

// IsSuccess returns true when this export manifest not found response has a 2xx status code
func (o *ExportManifestNotFound) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this export manifest not found response has a 3xx status code
func (o *ExportManifestNotFound) IsRedirect() bool {
	return false
}

// IsClientError returns true when this export manifest not found response has a 4xx status code
func (o *ExportManifestNotFound) IsClientError() bool {
	return true
}

// IsServerError returns true when this export manifest not found response has a 5xx status code
func (o *ExportManifestNotFound) IsServerError() bool {
	return false
}

// IsCode returns true when this export manifest not found response a status code equal to that given
func (o *ExportManifestNotFound) IsCode(code int) bool {
	return code == 404
}

func (o *ExportManifestNotFound) Error() string {
	return fmt.Sprintf("[GET /admin/manifest][%d] exportManifestNotFound  %+v", 404, o.Payload)
}

func (o *ExportManifestNotFound) String() string {
	return fmt.Sprintf("[GET /admin/manifest][%d] exportManifestNotFound  %+v", 404, o.Payload)
}

func (o *ExportManifestNotFound) GetPayload() *models.Error {
	return o.Payload
}

func (o *ExportManifestNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewExportManifestInternalServerError creates a ExportManifestInternalServerError with default headers values
func NewExportManifestInternalServerError() *ExportManifestInternalServerError {
	return &ExportManifestInternalServerError{}
}

/* ExportManifestInternalServerError describes a response with status code 500, with default header values.

Internal Error
*/
type ExportManifestInternalServerError struct {
	Payload *models.Error
}

// IsSuccess returns true when this export manifest internal server error response has a 2xx status code
func (o *ExportManifestInternalServerError) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this export manifest internal server error response has a 3xx status code
func (o *ExportManifestInternalServerError) IsRedirect() bool {
	return false
}

// IsClientError returns true when this export manifest internal server error response has a 4xx status code
func (o *ExportManifestInternalServerError) IsClientError() bool {
	return false
}

// IsServerError returns true when this export manifest internal server error response has a 5xx status code
func (o *ExportManifestInternalServerError) IsServerError() bool {
	return true
}

// IsCode returns true when this export manifest internal server error response a status code equal to that given
func (o *ExportManifestInternalServerError) IsCode(code int) bool {
	return code == 500
}

func (o *ExportManifestInternalServerError) Error() string {
	return fmt.Sprintf("[GET /admin/manifest][%d] exportManifestInternalServerError  %+v", 500, o.Payload)
}

func (o *ExportManifestInternalServerError) String() string {
	return fmt.Sprintf("[GET /admin/manifest][%d] exportManifestInternalServerError  %+v", 500, o.Payload)
}

func (o *ExportManifestInternalServerError) GetPayload() *models.Error {
	return o.Payload
}

func (o *ExportManifestInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
