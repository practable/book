// Code generated by go-swagger; DO NOT EDIT.

package operations

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

// NewListAllowedParams creates a new ListAllowedParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewListAllowedParams() *ListAllowedParams {
	return &ListAllowedParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewListAllowedParamsWithTimeout creates a new ListAllowedParams object
// with the ability to set a timeout on a request.
func NewListAllowedParamsWithTimeout(timeout time.Duration) *ListAllowedParams {
	return &ListAllowedParams{
		timeout: timeout,
	}
}

// NewListAllowedParamsWithContext creates a new ListAllowedParams object
// with the ability to set a context for a request.
func NewListAllowedParamsWithContext(ctx context.Context) *ListAllowedParams {
	return &ListAllowedParams{
		Context: ctx,
	}
}

// NewListAllowedParamsWithHTTPClient creates a new ListAllowedParams object
// with the ability to set a custom HTTPClient for a request.
func NewListAllowedParamsWithHTTPClient(client *http.Client) *ListAllowedParams {
	return &ListAllowedParams{
		HTTPClient: client,
	}
}

/* ListAllowedParams contains all the parameters to send to the API endpoint
   for the list allowed operation.

   Typically these are written to a http.Request.
*/
type ListAllowedParams struct {
	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the list allowed params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ListAllowedParams) WithDefaults() *ListAllowedParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the list allowed params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ListAllowedParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the list allowed params
func (o *ListAllowedParams) WithTimeout(timeout time.Duration) *ListAllowedParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the list allowed params
func (o *ListAllowedParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the list allowed params
func (o *ListAllowedParams) WithContext(ctx context.Context) *ListAllowedParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the list allowed params
func (o *ListAllowedParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the list allowed params
func (o *ListAllowedParams) WithHTTPClient(client *http.Client) *ListAllowedParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the list allowed params
func (o *ListAllowedParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WriteToRequest writes these params to a swagger request
func (o *ListAllowedParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
