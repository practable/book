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

// NewReplaceManifestParams creates a new ReplaceManifestParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewReplaceManifestParams() *ReplaceManifestParams {
	return &ReplaceManifestParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewReplaceManifestParamsWithTimeout creates a new ReplaceManifestParams object
// with the ability to set a timeout on a request.
func NewReplaceManifestParamsWithTimeout(timeout time.Duration) *ReplaceManifestParams {
	return &ReplaceManifestParams{
		timeout: timeout,
	}
}

// NewReplaceManifestParamsWithContext creates a new ReplaceManifestParams object
// with the ability to set a context for a request.
func NewReplaceManifestParamsWithContext(ctx context.Context) *ReplaceManifestParams {
	return &ReplaceManifestParams{
		Context: ctx,
	}
}

// NewReplaceManifestParamsWithHTTPClient creates a new ReplaceManifestParams object
// with the ability to set a custom HTTPClient for a request.
func NewReplaceManifestParamsWithHTTPClient(client *http.Client) *ReplaceManifestParams {
	return &ReplaceManifestParams{
		HTTPClient: client,
	}
}

/* ReplaceManifestParams contains all the parameters to send to the API endpoint
   for the replace manifest operation.

   Typically these are written to a http.Request.
*/
type ReplaceManifestParams struct {

	// Manifest.
	Manifest string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the replace manifest params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ReplaceManifestParams) WithDefaults() *ReplaceManifestParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the replace manifest params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ReplaceManifestParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the replace manifest params
func (o *ReplaceManifestParams) WithTimeout(timeout time.Duration) *ReplaceManifestParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the replace manifest params
func (o *ReplaceManifestParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the replace manifest params
func (o *ReplaceManifestParams) WithContext(ctx context.Context) *ReplaceManifestParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the replace manifest params
func (o *ReplaceManifestParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the replace manifest params
func (o *ReplaceManifestParams) WithHTTPClient(client *http.Client) *ReplaceManifestParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the replace manifest params
func (o *ReplaceManifestParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithManifest adds the manifest to the replace manifest params
func (o *ReplaceManifestParams) WithManifest(manifest string) *ReplaceManifestParams {
	o.SetManifest(manifest)
	return o
}

// SetManifest adds the manifest to the replace manifest params
func (o *ReplaceManifestParams) SetManifest(manifest string) {
	o.Manifest = manifest
}

// WriteToRequest writes these params to a swagger request
func (o *ReplaceManifestParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error
	if err := r.SetBodyParam(o.Manifest); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
