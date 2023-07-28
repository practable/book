// Code generated by go-swagger; DO NOT EDIT.

package admin

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

// NewGetResourcesParams creates a new GetResourcesParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewGetResourcesParams() *GetResourcesParams {
	return &GetResourcesParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewGetResourcesParamsWithTimeout creates a new GetResourcesParams object
// with the ability to set a timeout on a request.
func NewGetResourcesParamsWithTimeout(timeout time.Duration) *GetResourcesParams {
	return &GetResourcesParams{
		timeout: timeout,
	}
}

// NewGetResourcesParamsWithContext creates a new GetResourcesParams object
// with the ability to set a context for a request.
func NewGetResourcesParamsWithContext(ctx context.Context) *GetResourcesParams {
	return &GetResourcesParams{
		Context: ctx,
	}
}

// NewGetResourcesParamsWithHTTPClient creates a new GetResourcesParams object
// with the ability to set a custom HTTPClient for a request.
func NewGetResourcesParamsWithHTTPClient(client *http.Client) *GetResourcesParams {
	return &GetResourcesParams{
		HTTPClient: client,
	}
}

/* GetResourcesParams contains all the parameters to send to the API endpoint
   for the get resources operation.

   Typically these are written to a http.Request.
*/
type GetResourcesParams struct {
	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the get resources params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *GetResourcesParams) WithDefaults() *GetResourcesParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the get resources params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *GetResourcesParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the get resources params
func (o *GetResourcesParams) WithTimeout(timeout time.Duration) *GetResourcesParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the get resources params
func (o *GetResourcesParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the get resources params
func (o *GetResourcesParams) WithContext(ctx context.Context) *GetResourcesParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the get resources params
func (o *GetResourcesParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the get resources params
func (o *GetResourcesParams) WithHTTPClient(client *http.Client) *GetResourcesParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the get resources params
func (o *GetResourcesParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WriteToRequest writes these params to a swagger request
func (o *GetResourcesParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
