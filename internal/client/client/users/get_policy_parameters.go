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

// NewGetPolicyParams creates a new GetPolicyParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewGetPolicyParams() *GetPolicyParams {
	return &GetPolicyParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewGetPolicyParamsWithTimeout creates a new GetPolicyParams object
// with the ability to set a timeout on a request.
func NewGetPolicyParamsWithTimeout(timeout time.Duration) *GetPolicyParams {
	return &GetPolicyParams{
		timeout: timeout,
	}
}

// NewGetPolicyParamsWithContext creates a new GetPolicyParams object
// with the ability to set a context for a request.
func NewGetPolicyParamsWithContext(ctx context.Context) *GetPolicyParams {
	return &GetPolicyParams{
		Context: ctx,
	}
}

// NewGetPolicyParamsWithHTTPClient creates a new GetPolicyParams object
// with the ability to set a custom HTTPClient for a request.
func NewGetPolicyParamsWithHTTPClient(client *http.Client) *GetPolicyParams {
	return &GetPolicyParams{
		HTTPClient: client,
	}
}

/* GetPolicyParams contains all the parameters to send to the API endpoint
   for the get policy operation.

   Typically these are written to a http.Request.
*/
type GetPolicyParams struct {

	// PolicyName.
	PolicyName string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the get policy params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *GetPolicyParams) WithDefaults() *GetPolicyParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the get policy params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *GetPolicyParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the get policy params
func (o *GetPolicyParams) WithTimeout(timeout time.Duration) *GetPolicyParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the get policy params
func (o *GetPolicyParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the get policy params
func (o *GetPolicyParams) WithContext(ctx context.Context) *GetPolicyParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the get policy params
func (o *GetPolicyParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the get policy params
func (o *GetPolicyParams) WithHTTPClient(client *http.Client) *GetPolicyParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the get policy params
func (o *GetPolicyParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithPolicyName adds the policyName to the get policy params
func (o *GetPolicyParams) WithPolicyName(policyName string) *GetPolicyParams {
	o.SetPolicyName(policyName)
	return o
}

// SetPolicyName adds the policyName to the get policy params
func (o *GetPolicyParams) SetPolicyName(policyName string) {
	o.PolicyName = policyName
}

// WriteToRequest writes these params to a swagger request
func (o *GetPolicyParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// path param policy_name
	if err := r.SetPathParam("policy_name", o.PolicyName); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
