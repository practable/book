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

// NewMakeBookingParams creates a new MakeBookingParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewMakeBookingParams() *MakeBookingParams {
	return &MakeBookingParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewMakeBookingParamsWithTimeout creates a new MakeBookingParams object
// with the ability to set a timeout on a request.
func NewMakeBookingParamsWithTimeout(timeout time.Duration) *MakeBookingParams {
	return &MakeBookingParams{
		timeout: timeout,
	}
}

// NewMakeBookingParamsWithContext creates a new MakeBookingParams object
// with the ability to set a context for a request.
func NewMakeBookingParamsWithContext(ctx context.Context) *MakeBookingParams {
	return &MakeBookingParams{
		Context: ctx,
	}
}

// NewMakeBookingParamsWithHTTPClient creates a new MakeBookingParams object
// with the ability to set a custom HTTPClient for a request.
func NewMakeBookingParamsWithHTTPClient(client *http.Client) *MakeBookingParams {
	return &MakeBookingParams{
		HTTPClient: client,
	}
}

/* MakeBookingParams contains all the parameters to send to the API endpoint
   for the make booking operation.

   Typically these are written to a http.Request.
*/
type MakeBookingParams struct {

	// From.
	//
	// Format: date-time
	From strfmt.DateTime

	// PolicyName.
	PolicyName string

	// SlotName.
	SlotName string

	// To.
	//
	// Format: date-time
	To strfmt.DateTime

	// UserName.
	UserName string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the make booking params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *MakeBookingParams) WithDefaults() *MakeBookingParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the make booking params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *MakeBookingParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the make booking params
func (o *MakeBookingParams) WithTimeout(timeout time.Duration) *MakeBookingParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the make booking params
func (o *MakeBookingParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the make booking params
func (o *MakeBookingParams) WithContext(ctx context.Context) *MakeBookingParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the make booking params
func (o *MakeBookingParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the make booking params
func (o *MakeBookingParams) WithHTTPClient(client *http.Client) *MakeBookingParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the make booking params
func (o *MakeBookingParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithFrom adds the from to the make booking params
func (o *MakeBookingParams) WithFrom(from strfmt.DateTime) *MakeBookingParams {
	o.SetFrom(from)
	return o
}

// SetFrom adds the from to the make booking params
func (o *MakeBookingParams) SetFrom(from strfmt.DateTime) {
	o.From = from
}

// WithPolicyName adds the policyName to the make booking params
func (o *MakeBookingParams) WithPolicyName(policyName string) *MakeBookingParams {
	o.SetPolicyName(policyName)
	return o
}

// SetPolicyName adds the policyName to the make booking params
func (o *MakeBookingParams) SetPolicyName(policyName string) {
	o.PolicyName = policyName
}

// WithSlotName adds the slotName to the make booking params
func (o *MakeBookingParams) WithSlotName(slotName string) *MakeBookingParams {
	o.SetSlotName(slotName)
	return o
}

// SetSlotName adds the slotName to the make booking params
func (o *MakeBookingParams) SetSlotName(slotName string) {
	o.SlotName = slotName
}

// WithTo adds the to to the make booking params
func (o *MakeBookingParams) WithTo(to strfmt.DateTime) *MakeBookingParams {
	o.SetTo(to)
	return o
}

// SetTo adds the to to the make booking params
func (o *MakeBookingParams) SetTo(to strfmt.DateTime) {
	o.To = to
}

// WithUserName adds the userName to the make booking params
func (o *MakeBookingParams) WithUserName(userName string) *MakeBookingParams {
	o.SetUserName(userName)
	return o
}

// SetUserName adds the userName to the make booking params
func (o *MakeBookingParams) SetUserName(userName string) {
	o.UserName = userName
}

// WriteToRequest writes these params to a swagger request
func (o *MakeBookingParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// query param from
	qrFrom := o.From
	qFrom := qrFrom.String()
	if qFrom != "" {

		if err := r.SetQueryParam("from", qFrom); err != nil {
			return err
		}
	}

	// path param policy_name
	if err := r.SetPathParam("policy_name", o.PolicyName); err != nil {
		return err
	}

	// path param slot_name
	if err := r.SetPathParam("slot_name", o.SlotName); err != nil {
		return err
	}

	// query param to
	qrTo := o.To
	qTo := qrTo.String()
	if qTo != "" {

		if err := r.SetQueryParam("to", qTo); err != nil {
			return err
		}
	}

	// query param user_name
	qrUserName := o.UserName
	qUserName := qrUserName
	if qUserName != "" {

		if err := r.SetQueryParam("user_name", qUserName); err != nil {
			return err
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
