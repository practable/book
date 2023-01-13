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

// NewGetStoreStatusUserParams creates a new GetStoreStatusUserParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewGetStoreStatusUserParams() *GetStoreStatusUserParams {
	return &GetStoreStatusUserParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewGetStoreStatusUserParamsWithTimeout creates a new GetStoreStatusUserParams object
// with the ability to set a timeout on a request.
func NewGetStoreStatusUserParamsWithTimeout(timeout time.Duration) *GetStoreStatusUserParams {
	return &GetStoreStatusUserParams{
		timeout: timeout,
	}
}

// NewGetStoreStatusUserParamsWithContext creates a new GetStoreStatusUserParams object
// with the ability to set a context for a request.
func NewGetStoreStatusUserParamsWithContext(ctx context.Context) *GetStoreStatusUserParams {
	return &GetStoreStatusUserParams{
		Context: ctx,
	}
}

// NewGetStoreStatusUserParamsWithHTTPClient creates a new GetStoreStatusUserParams object
// with the ability to set a custom HTTPClient for a request.
func NewGetStoreStatusUserParamsWithHTTPClient(client *http.Client) *GetStoreStatusUserParams {
	return &GetStoreStatusUserParams{
		HTTPClient: client,
	}
}

/* GetStoreStatusUserParams contains all the parameters to send to the API endpoint
   for the get store status user operation.

   Typically these are written to a http.Request.
*/
type GetStoreStatusUserParams struct {
	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the get store status user params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *GetStoreStatusUserParams) WithDefaults() *GetStoreStatusUserParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the get store status user params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *GetStoreStatusUserParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the get store status user params
func (o *GetStoreStatusUserParams) WithTimeout(timeout time.Duration) *GetStoreStatusUserParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the get store status user params
func (o *GetStoreStatusUserParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the get store status user params
func (o *GetStoreStatusUserParams) WithContext(ctx context.Context) *GetStoreStatusUserParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the get store status user params
func (o *GetStoreStatusUserParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the get store status user params
func (o *GetStoreStatusUserParams) WithHTTPClient(client *http.Client) *GetStoreStatusUserParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the get store status user params
func (o *GetStoreStatusUserParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WriteToRequest writes these params to a swagger request
func (o *GetStoreStatusUserParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
