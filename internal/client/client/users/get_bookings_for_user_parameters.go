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

// NewGetBookingsForUserParams creates a new GetBookingsForUserParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewGetBookingsForUserParams() *GetBookingsForUserParams {
	return &GetBookingsForUserParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewGetBookingsForUserParamsWithTimeout creates a new GetBookingsForUserParams object
// with the ability to set a timeout on a request.
func NewGetBookingsForUserParamsWithTimeout(timeout time.Duration) *GetBookingsForUserParams {
	return &GetBookingsForUserParams{
		timeout: timeout,
	}
}

// NewGetBookingsForUserParamsWithContext creates a new GetBookingsForUserParams object
// with the ability to set a context for a request.
func NewGetBookingsForUserParamsWithContext(ctx context.Context) *GetBookingsForUserParams {
	return &GetBookingsForUserParams{
		Context: ctx,
	}
}

// NewGetBookingsForUserParamsWithHTTPClient creates a new GetBookingsForUserParams object
// with the ability to set a custom HTTPClient for a request.
func NewGetBookingsForUserParamsWithHTTPClient(client *http.Client) *GetBookingsForUserParams {
	return &GetBookingsForUserParams{
		HTTPClient: client,
	}
}

/* GetBookingsForUserParams contains all the parameters to send to the API endpoint
   for the get bookings for user operation.

   Typically these are written to a http.Request.
*/
type GetBookingsForUserParams struct {

	// UserName.
	UserName string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the get bookings for user params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *GetBookingsForUserParams) WithDefaults() *GetBookingsForUserParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the get bookings for user params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *GetBookingsForUserParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the get bookings for user params
func (o *GetBookingsForUserParams) WithTimeout(timeout time.Duration) *GetBookingsForUserParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the get bookings for user params
func (o *GetBookingsForUserParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the get bookings for user params
func (o *GetBookingsForUserParams) WithContext(ctx context.Context) *GetBookingsForUserParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the get bookings for user params
func (o *GetBookingsForUserParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the get bookings for user params
func (o *GetBookingsForUserParams) WithHTTPClient(client *http.Client) *GetBookingsForUserParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the get bookings for user params
func (o *GetBookingsForUserParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithUserName adds the userName to the get bookings for user params
func (o *GetBookingsForUserParams) WithUserName(userName string) *GetBookingsForUserParams {
	o.SetUserName(userName)
	return o
}

// SetUserName adds the userName to the get bookings for user params
func (o *GetBookingsForUserParams) SetUserName(userName string) {
	o.UserName = userName
}

// WriteToRequest writes these params to a swagger request
func (o *GetBookingsForUserParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// path param user_name
	if err := r.SetPathParam("user_name", o.UserName); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
