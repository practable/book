// Code generated by go-swagger; DO NOT EDIT.

package admin

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
)

// New creates a new admin API client.
func New(transport runtime.ClientTransport, formats strfmt.Registry) ClientService {
	return &Client{transport: transport, formats: formats}
}

/*
Client for admin API
*/
type Client struct {
	transport runtime.ClientTransport
	formats   strfmt.Registry
}

// ClientOption is the option for Client methods
type ClientOption func(*runtime.ClientOperation)

// ClientService is the interface for Client methods
type ClientService interface {
	CheckManifest(params *CheckManifestParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*CheckManifestOK, *CheckManifestNoContent, error)

	ExportBookings(params *ExportBookingsParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*ExportBookingsOK, error)

	ExportManifest(params *ExportManifestParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*ExportManifestOK, error)

	ExportOldBookings(params *ExportOldBookingsParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*ExportOldBookingsOK, error)

	ExportUsers(params *ExportUsersParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*ExportUsersOK, error)

	GetSlotIsAvailable(params *GetSlotIsAvailableParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*GetSlotIsAvailableOK, error)

	ReplaceBookings(params *ReplaceBookingsParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*ReplaceBookingsOK, error)

	ReplaceManifest(params *ReplaceManifestParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*ReplaceManifestOK, error)

	ReplaceOldBookings(params *ReplaceOldBookingsParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*ReplaceOldBookingsOK, error)

	SetSlotIsAvailable(params *SetSlotIsAvailableParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*SetSlotIsAvailableNoContent, error)

	GetStoreStatusAdmin(params *GetStoreStatusAdminParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*GetStoreStatusAdminOK, error)

	SetLock(params *SetLockParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*SetLockOK, error)

	SetTransport(transport runtime.ClientTransport)
}

/*
  CheckManifest checks a manifest

  Check a manifest is valid. Returns 204 if valid or, 200 with a list of error(s).
*/
func (a *Client) CheckManifest(params *CheckManifestParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*CheckManifestOK, *CheckManifestNoContent, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewCheckManifestParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "CheckManifest",
		Method:             "GET",
		PathPattern:        "/admin/manifest/check",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &CheckManifestReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, nil, err
	}
	switch value := result.(type) {
	case *CheckManifestOK:
		return value, nil, nil
	case *CheckManifestNoContent:
		return nil, value, nil
	}
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for admin: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
  ExportBookings exports a copy of all current bookings

  Exports a copy of the current bookings, with sufficient information to allow editing and replacement. If successful produces JSON-formatted bookings list.
*/
func (a *Client) ExportBookings(params *ExportBookingsParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*ExportBookingsOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewExportBookingsParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "ExportBookings",
		Method:             "GET",
		PathPattern:        "/admin/bookings",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json", "text/plain"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &ExportBookingsReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*ExportBookingsOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for ExportBookings: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
  ExportManifest exports the manifest

  Export the manifest (resources, slots, policies, descriptions etc). Does not include bookings or users
*/
func (a *Client) ExportManifest(params *ExportManifestParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*ExportManifestOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewExportManifestParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "ExportManifest",
		Method:             "GET",
		PathPattern:        "/admin/manifest",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json", "text/plain"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &ExportManifestReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*ExportManifestOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for ExportManifest: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
  ExportOldBookings exports a copy of all old bookings

  Exports a copy of the old bookings, with sufficient information to allow editing and replacement. If successful produces JSON-formatted bookings list.
*/
func (a *Client) ExportOldBookings(params *ExportOldBookingsParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*ExportOldBookingsOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewExportOldBookingsParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "ExportOldBookings",
		Method:             "GET",
		PathPattern:        "/admin/oldbookings",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json", "text/plain"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &ExportOldBookingsReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*ExportOldBookingsOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for ExportOldBookings: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
  ExportUsers exports users

  Export bookings and usage data for each user
*/
func (a *Client) ExportUsers(params *ExportUsersParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*ExportUsersOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewExportUsersParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "ExportUsers",
		Method:             "GET",
		PathPattern:        "/admin/users",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json", "text/plain"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &ExportUsersReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*ExportUsersOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for ExportUsers: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
  GetSlotIsAvailable gets the availability of the slot

  Gets the availability of the underlying resource for the slot, including a status message. Indicates when equipment is offline temprorarily, e.g. due to failing an automated test.
*/
func (a *Client) GetSlotIsAvailable(params *GetSlotIsAvailableParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*GetSlotIsAvailableOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetSlotIsAvailableParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "GetSlotIsAvailable",
		Method:             "GET",
		PathPattern:        "/admin/slots/{slot_name}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json", "text/plain"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetSlotIsAvailableReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetSlotIsAvailableOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for GetSlotIsAvailable: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
  ReplaceBookings replaces current bookings

  Deletes all current bookings, refunds usage to users, and then replaces with current bookings. Existing users are retained, new users are created as required to match bookings.
*/
func (a *Client) ReplaceBookings(params *ReplaceBookingsParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*ReplaceBookingsOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewReplaceBookingsParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "ReplaceBookings",
		Method:             "PUT",
		PathPattern:        "/admin/bookings",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &ReplaceBookingsReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*ReplaceBookingsOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for ReplaceBookings: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
  ReplaceManifest replaces the manifest

  Delete the existing manifest and replace it with a new one. All items have specified names so bookings do not need updating (except perhaps you should if booked resources have been removed)
*/
func (a *Client) ReplaceManifest(params *ReplaceManifestParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*ReplaceManifestOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewReplaceManifestParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "ReplaceManifest",
		Method:             "PUT",
		PathPattern:        "/admin/manifest",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &ReplaceManifestReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*ReplaceManifestOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for ReplaceManifest: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
  ReplaceOldBookings replaces old bookings

  Deletes all old bookings, and all users, then replaces both according to the bookings in the request, i.e. users and their usage are created as required to match bookings.
*/
func (a *Client) ReplaceOldBookings(params *ReplaceOldBookingsParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*ReplaceOldBookingsOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewReplaceOldBookingsParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "ReplaceOldBookings",
		Method:             "PUT",
		PathPattern:        "/admin/oldbookings",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &ReplaceOldBookingsReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*ReplaceOldBookingsOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for ReplaceOldBookings: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
  SetSlotIsAvailable sets the availability of the slot

  Sets the availability of the underlying resource for the slot, including a status message. Used to prevent users accessing equipment that should not be used, e.g. after failing an automated test, or make it available again after fixing it.
*/
func (a *Client) SetSlotIsAvailable(params *SetSlotIsAvailableParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*SetSlotIsAvailableNoContent, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewSetSlotIsAvailableParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "SetSlotIsAvailable",
		Method:             "PUT",
		PathPattern:        "/admin/slots/{slot_name}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json", "text/plain"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &SetSlotIsAvailableReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*SetSlotIsAvailableNoContent)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for SetSlotIsAvailable: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
  GetStoreStatusAdmin gets current store status

  Gets a count of the number of elements in the store, e.g. Bookings, Descriptions etc to facilitate a necessary but not sufficient check that replace manifest and replace bookings have produced the correct results.
*/
func (a *Client) GetStoreStatusAdmin(params *GetStoreStatusAdminParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*GetStoreStatusAdminOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetStoreStatusAdminParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "getStoreStatusAdmin",
		Method:             "GET",
		PathPattern:        "/admin/status",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json", "text/plain"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetStoreStatusAdminReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetStoreStatusAdminOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for getStoreStatusAdmin: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
  SetLock sets release booking lock

  Set whether the booking system is locked for users
*/
func (a *Client) SetLock(params *SetLockParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*SetLockOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewSetLockParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "setLock",
		Method:             "PUT",
		PathPattern:        "/admin/status",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json", "text/plain"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &SetLockReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*SetLockOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for setLock: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

// SetTransport changes the transport on the client
func (a *Client) SetTransport(transport runtime.ClientTransport) {
	a.transport = transport
}
