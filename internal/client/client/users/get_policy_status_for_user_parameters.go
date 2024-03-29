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

// NewGetPolicyStatusForUserParams creates a new GetPolicyStatusForUserParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewGetPolicyStatusForUserParams() *GetPolicyStatusForUserParams {
	return &GetPolicyStatusForUserParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewGetPolicyStatusForUserParamsWithTimeout creates a new GetPolicyStatusForUserParams object
// with the ability to set a timeout on a request.
func NewGetPolicyStatusForUserParamsWithTimeout(timeout time.Duration) *GetPolicyStatusForUserParams {
	return &GetPolicyStatusForUserParams{
		timeout: timeout,
	}
}

// NewGetPolicyStatusForUserParamsWithContext creates a new GetPolicyStatusForUserParams object
// with the ability to set a context for a request.
func NewGetPolicyStatusForUserParamsWithContext(ctx context.Context) *GetPolicyStatusForUserParams {
	return &GetPolicyStatusForUserParams{
		Context: ctx,
	}
}

// NewGetPolicyStatusForUserParamsWithHTTPClient creates a new GetPolicyStatusForUserParams object
// with the ability to set a custom HTTPClient for a request.
func NewGetPolicyStatusForUserParamsWithHTTPClient(client *http.Client) *GetPolicyStatusForUserParams {
	return &GetPolicyStatusForUserParams{
		HTTPClient: client,
	}
}

/* GetPolicyStatusForUserParams contains all the parameters to send to the API endpoint
   for the get policy status for user operation.

   Typically these are written to a http.Request.
*/
type GetPolicyStatusForUserParams struct {

	// PolicyName.
	PolicyName string

	// UserName.
	UserName string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the get policy status for user params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *GetPolicyStatusForUserParams) WithDefaults() *GetPolicyStatusForUserParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the get policy status for user params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *GetPolicyStatusForUserParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the get policy status for user params
func (o *GetPolicyStatusForUserParams) WithTimeout(timeout time.Duration) *GetPolicyStatusForUserParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the get policy status for user params
func (o *GetPolicyStatusForUserParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the get policy status for user params
func (o *GetPolicyStatusForUserParams) WithContext(ctx context.Context) *GetPolicyStatusForUserParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the get policy status for user params
func (o *GetPolicyStatusForUserParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the get policy status for user params
func (o *GetPolicyStatusForUserParams) WithHTTPClient(client *http.Client) *GetPolicyStatusForUserParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the get policy status for user params
func (o *GetPolicyStatusForUserParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithPolicyName adds the policyName to the get policy status for user params
func (o *GetPolicyStatusForUserParams) WithPolicyName(policyName string) *GetPolicyStatusForUserParams {
	o.SetPolicyName(policyName)
	return o
}

// SetPolicyName adds the policyName to the get policy status for user params
func (o *GetPolicyStatusForUserParams) SetPolicyName(policyName string) {
	o.PolicyName = policyName
}

// WithUserName adds the userName to the get policy status for user params
func (o *GetPolicyStatusForUserParams) WithUserName(userName string) *GetPolicyStatusForUserParams {
	o.SetUserName(userName)
	return o
}

// SetUserName adds the userName to the get policy status for user params
func (o *GetPolicyStatusForUserParams) SetUserName(userName string) {
	o.UserName = userName
}

// WriteToRequest writes these params to a swagger request
func (o *GetPolicyStatusForUserParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// path param policy_name
	if err := r.SetPathParam("policy_name", o.PolicyName); err != nil {
		return err
	}

	// path param user_name
	if err := r.SetPathParam("user_name", o.UserName); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
