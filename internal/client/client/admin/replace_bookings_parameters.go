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

	"github.com/timdrysdale/interval/internal/client/models"
)

// NewReplaceBookingsParams creates a new ReplaceBookingsParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewReplaceBookingsParams() *ReplaceBookingsParams {
	return &ReplaceBookingsParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewReplaceBookingsParamsWithTimeout creates a new ReplaceBookingsParams object
// with the ability to set a timeout on a request.
func NewReplaceBookingsParamsWithTimeout(timeout time.Duration) *ReplaceBookingsParams {
	return &ReplaceBookingsParams{
		timeout: timeout,
	}
}

// NewReplaceBookingsParamsWithContext creates a new ReplaceBookingsParams object
// with the ability to set a context for a request.
func NewReplaceBookingsParamsWithContext(ctx context.Context) *ReplaceBookingsParams {
	return &ReplaceBookingsParams{
		Context: ctx,
	}
}

// NewReplaceBookingsParamsWithHTTPClient creates a new ReplaceBookingsParams object
// with the ability to set a custom HTTPClient for a request.
func NewReplaceBookingsParamsWithHTTPClient(client *http.Client) *ReplaceBookingsParams {
	return &ReplaceBookingsParams{
		HTTPClient: client,
	}
}

/*
ReplaceBookingsParams contains all the parameters to send to the API endpoint

	for the replace bookings operation.

	Typically these are written to a http.Request.
*/
type ReplaceBookingsParams struct {

	// Bookings.
	Bookings models.Bookings

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the replace bookings params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ReplaceBookingsParams) WithDefaults() *ReplaceBookingsParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the replace bookings params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ReplaceBookingsParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the replace bookings params
func (o *ReplaceBookingsParams) WithTimeout(timeout time.Duration) *ReplaceBookingsParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the replace bookings params
func (o *ReplaceBookingsParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the replace bookings params
func (o *ReplaceBookingsParams) WithContext(ctx context.Context) *ReplaceBookingsParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the replace bookings params
func (o *ReplaceBookingsParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the replace bookings params
func (o *ReplaceBookingsParams) WithHTTPClient(client *http.Client) *ReplaceBookingsParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the replace bookings params
func (o *ReplaceBookingsParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithBookings adds the bookings to the replace bookings params
func (o *ReplaceBookingsParams) WithBookings(bookings models.Bookings) *ReplaceBookingsParams {
	o.SetBookings(bookings)
	return o
}

// SetBookings adds the bookings to the replace bookings params
func (o *ReplaceBookingsParams) SetBookings(bookings models.Bookings) {
	o.Bookings = bookings
}

// WriteToRequest writes these params to a swagger request
func (o *ReplaceBookingsParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error
	if o.Bookings != nil {
		if err := r.SetBodyParam(o.Bookings); err != nil {
			return err
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}