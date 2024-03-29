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
	"github.com/go-openapi/swag"
)

// NewDenyParams creates a new DenyParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewDenyParams() *DenyParams {
	return &DenyParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewDenyParamsWithTimeout creates a new DenyParams object
// with the ability to set a timeout on a request.
func NewDenyParamsWithTimeout(timeout time.Duration) *DenyParams {
	return &DenyParams{
		timeout: timeout,
	}
}

// NewDenyParamsWithContext creates a new DenyParams object
// with the ability to set a context for a request.
func NewDenyParamsWithContext(ctx context.Context) *DenyParams {
	return &DenyParams{
		Context: ctx,
	}
}

// NewDenyParamsWithHTTPClient creates a new DenyParams object
// with the ability to set a custom HTTPClient for a request.
func NewDenyParamsWithHTTPClient(client *http.Client) *DenyParams {
	return &DenyParams{
		HTTPClient: client,
	}
}

/* DenyParams contains all the parameters to send to the API endpoint
   for the deny operation.

   Typically these are written to a http.Request.
*/
type DenyParams struct {

	// Bid.
	Bid string

	// Exp.
	Exp int64

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the deny params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *DenyParams) WithDefaults() *DenyParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the deny params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *DenyParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the deny params
func (o *DenyParams) WithTimeout(timeout time.Duration) *DenyParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the deny params
func (o *DenyParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the deny params
func (o *DenyParams) WithContext(ctx context.Context) *DenyParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the deny params
func (o *DenyParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the deny params
func (o *DenyParams) WithHTTPClient(client *http.Client) *DenyParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the deny params
func (o *DenyParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithBid adds the bid to the deny params
func (o *DenyParams) WithBid(bid string) *DenyParams {
	o.SetBid(bid)
	return o
}

// SetBid adds the bid to the deny params
func (o *DenyParams) SetBid(bid string) {
	o.Bid = bid
}

// WithExp adds the exp to the deny params
func (o *DenyParams) WithExp(exp int64) *DenyParams {
	o.SetExp(exp)
	return o
}

// SetExp adds the exp to the deny params
func (o *DenyParams) SetExp(exp int64) {
	o.Exp = exp
}

// WriteToRequest writes these params to a swagger request
func (o *DenyParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// query param bid
	qrBid := o.Bid
	qBid := qrBid
	if qBid != "" {

		if err := r.SetQueryParam("bid", qBid); err != nil {
			return err
		}
	}

	// query param exp
	qrExp := o.Exp
	qExp := swag.FormatInt64(qrExp)
	if qExp != "" {

		if err := r.SetQueryParam("exp", qExp); err != nil {
			return err
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
