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

	"github.com/practable/book/internal/client/models"
)

// NewReplaceOldBookingsParams creates a new ReplaceOldBookingsParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewReplaceOldBookingsParams() *ReplaceOldBookingsParams {
	return &ReplaceOldBookingsParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewReplaceOldBookingsParamsWithTimeout creates a new ReplaceOldBookingsParams object
// with the ability to set a timeout on a request.
func NewReplaceOldBookingsParamsWithTimeout(timeout time.Duration) *ReplaceOldBookingsParams {
	return &ReplaceOldBookingsParams{
		timeout: timeout,
	}
}

// NewReplaceOldBookingsParamsWithContext creates a new ReplaceOldBookingsParams object
// with the ability to set a context for a request.
func NewReplaceOldBookingsParamsWithContext(ctx context.Context) *ReplaceOldBookingsParams {
	return &ReplaceOldBookingsParams{
		Context: ctx,
	}
}

// NewReplaceOldBookingsParamsWithHTTPClient creates a new ReplaceOldBookingsParams object
// with the ability to set a custom HTTPClient for a request.
func NewReplaceOldBookingsParamsWithHTTPClient(client *http.Client) *ReplaceOldBookingsParams {
	return &ReplaceOldBookingsParams{
		HTTPClient: client,
	}
}

/* ReplaceOldBookingsParams contains all the parameters to send to the API endpoint
   for the replace old bookings operation.

   Typically these are written to a http.Request.
*/
type ReplaceOldBookingsParams struct {

	// Bookings.
	Bookings models.Bookings

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the replace old bookings params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ReplaceOldBookingsParams) WithDefaults() *ReplaceOldBookingsParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the replace old bookings params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ReplaceOldBookingsParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the replace old bookings params
func (o *ReplaceOldBookingsParams) WithTimeout(timeout time.Duration) *ReplaceOldBookingsParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the replace old bookings params
func (o *ReplaceOldBookingsParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the replace old bookings params
func (o *ReplaceOldBookingsParams) WithContext(ctx context.Context) *ReplaceOldBookingsParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the replace old bookings params
func (o *ReplaceOldBookingsParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the replace old bookings params
func (o *ReplaceOldBookingsParams) WithHTTPClient(client *http.Client) *ReplaceOldBookingsParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the replace old bookings params
func (o *ReplaceOldBookingsParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithBookings adds the bookings to the replace old bookings params
func (o *ReplaceOldBookingsParams) WithBookings(bookings models.Bookings) *ReplaceOldBookingsParams {
	o.SetBookings(bookings)
	return o
}

// SetBookings adds the bookings to the replace old bookings params
func (o *ReplaceOldBookingsParams) SetBookings(bookings models.Bookings) {
	o.Bookings = bookings
}

// WriteToRequest writes these params to a swagger request
func (o *ReplaceOldBookingsParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

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
