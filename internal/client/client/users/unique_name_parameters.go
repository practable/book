// Code generated by go-swagger; DO NOT EDIT.

package users

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

// NewUniqueNameParams creates a new UniqueNameParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewUniqueNameParams() *UniqueNameParams {
	return &UniqueNameParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewUniqueNameParamsWithTimeout creates a new UniqueNameParams object
// with the ability to set a timeout on a request.
func NewUniqueNameParamsWithTimeout(timeout time.Duration) *UniqueNameParams {
	return &UniqueNameParams{
		timeout: timeout,
	}
}

// NewUniqueNameParamsWithContext creates a new UniqueNameParams object
// with the ability to set a context for a request.
func NewUniqueNameParamsWithContext(ctx context.Context) *UniqueNameParams {
	return &UniqueNameParams{
		Context: ctx,
	}
}

// NewUniqueNameParamsWithHTTPClient creates a new UniqueNameParams object
// with the ability to set a custom HTTPClient for a request.
func NewUniqueNameParamsWithHTTPClient(client *http.Client) *UniqueNameParams {
	return &UniqueNameParams{
		HTTPClient: client,
	}
}

/* UniqueNameParams contains all the parameters to send to the API endpoint
   for the unique name operation.

   Typically these are written to a http.Request.
*/
type UniqueNameParams struct {
	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the unique name params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *UniqueNameParams) WithDefaults() *UniqueNameParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the unique name params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *UniqueNameParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the unique name params
func (o *UniqueNameParams) WithTimeout(timeout time.Duration) *UniqueNameParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the unique name params
func (o *UniqueNameParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the unique name params
func (o *UniqueNameParams) WithContext(ctx context.Context) *UniqueNameParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the unique name params
func (o *UniqueNameParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the unique name params
func (o *UniqueNameParams) WithHTTPClient(client *http.Client) *UniqueNameParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the unique name params
func (o *UniqueNameParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WriteToRequest writes these params to a swagger request
func (o *UniqueNameParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}