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
	"github.com/go-openapi/swag"
)

// NewGetAvailabilityParams creates a new GetAvailabilityParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewGetAvailabilityParams() *GetAvailabilityParams {
	return &GetAvailabilityParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewGetAvailabilityParamsWithTimeout creates a new GetAvailabilityParams object
// with the ability to set a timeout on a request.
func NewGetAvailabilityParamsWithTimeout(timeout time.Duration) *GetAvailabilityParams {
	return &GetAvailabilityParams{
		timeout: timeout,
	}
}

// NewGetAvailabilityParamsWithContext creates a new GetAvailabilityParams object
// with the ability to set a context for a request.
func NewGetAvailabilityParamsWithContext(ctx context.Context) *GetAvailabilityParams {
	return &GetAvailabilityParams{
		Context: ctx,
	}
}

// NewGetAvailabilityParamsWithHTTPClient creates a new GetAvailabilityParams object
// with the ability to set a custom HTTPClient for a request.
func NewGetAvailabilityParamsWithHTTPClient(client *http.Client) *GetAvailabilityParams {
	return &GetAvailabilityParams{
		HTTPClient: client,
	}
}

/* GetAvailabilityParams contains all the parameters to send to the API endpoint
   for the get availability operation.

   Typically these are written to a http.Request.
*/
type GetAvailabilityParams struct {

	// Limit.
	Limit *int64

	// Offset.
	Offset *int64

	// SlotName.
	SlotName string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the get availability params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *GetAvailabilityParams) WithDefaults() *GetAvailabilityParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the get availability params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *GetAvailabilityParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the get availability params
func (o *GetAvailabilityParams) WithTimeout(timeout time.Duration) *GetAvailabilityParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the get availability params
func (o *GetAvailabilityParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the get availability params
func (o *GetAvailabilityParams) WithContext(ctx context.Context) *GetAvailabilityParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the get availability params
func (o *GetAvailabilityParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the get availability params
func (o *GetAvailabilityParams) WithHTTPClient(client *http.Client) *GetAvailabilityParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the get availability params
func (o *GetAvailabilityParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithLimit adds the limit to the get availability params
func (o *GetAvailabilityParams) WithLimit(limit *int64) *GetAvailabilityParams {
	o.SetLimit(limit)
	return o
}

// SetLimit adds the limit to the get availability params
func (o *GetAvailabilityParams) SetLimit(limit *int64) {
	o.Limit = limit
}

// WithOffset adds the offset to the get availability params
func (o *GetAvailabilityParams) WithOffset(offset *int64) *GetAvailabilityParams {
	o.SetOffset(offset)
	return o
}

// SetOffset adds the offset to the get availability params
func (o *GetAvailabilityParams) SetOffset(offset *int64) {
	o.Offset = offset
}

// WithSlotName adds the slotName to the get availability params
func (o *GetAvailabilityParams) WithSlotName(slotName string) *GetAvailabilityParams {
	o.SetSlotName(slotName)
	return o
}

// SetSlotName adds the slotName to the get availability params
func (o *GetAvailabilityParams) SetSlotName(slotName string) {
	o.SlotName = slotName
}

// WriteToRequest writes these params to a swagger request
func (o *GetAvailabilityParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if o.Limit != nil {

		// query param limit
		var qrLimit int64

		if o.Limit != nil {
			qrLimit = *o.Limit
		}
		qLimit := swag.FormatInt64(qrLimit)
		if qLimit != "" {

			if err := r.SetQueryParam("limit", qLimit); err != nil {
				return err
			}
		}
	}

	if o.Offset != nil {

		// query param offset
		var qrOffset int64

		if o.Offset != nil {
			qrOffset = *o.Offset
		}
		qOffset := swag.FormatInt64(qrOffset)
		if qOffset != "" {

			if err := r.SetQueryParam("offset", qOffset); err != nil {
				return err
			}
		}
	}

	// path param slot_name
	if err := r.SetPathParam("slot_name", o.SlotName); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
