package client

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/go-openapi/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type diagnosticTestResponse struct {
	code int
	body string
}

func (r diagnosticTestResponse) Code() int                  { return r.code }
func (r diagnosticTestResponse) Message() string            { return "Unprocessable Entity" }
func (r diagnosticTestResponse) GetHeader(string) string    { return "" }
func (r diagnosticTestResponse) GetHeaders(string) []string { return nil }
func (r diagnosticTestResponse) Body() io.ReadCloser        { return io.NopCloser(strings.NewReader(r.body)) }

func TestDiagnosticResponseReaderPreservesUnexpectedJSONError(t *testing.T) {
	response := diagnosticTestResponse{
		code: 422,
		body: `{"code":602,"message":"display_guides.1m.label in body is required"}`,
	}
	generatedFallback := runtime.ClientResponseReaderFunc(func(response runtime.ClientResponse, _ runtime.Consumer) (interface{}, error) {
		return nil, runtime.NewAPIError("[PUT /admin/manifest] ReplaceManifest", response, response.Code())
	})

	_, err := (&diagnosticResponseReader{next: generatedFallback}).ReadResponse(response, runtime.JSONConsumer())
	require.Error(t, err)
	assert.Contains(t, err.Error(), `"code":602`)
	assert.Contains(t, err.Error(), `"message":"display_guides.1m.label in body is required"`)
	assert.NotContains(t, err.Error(), ": {}")
}

func TestDiagnosticResponseReaderLeavesDeclaredErrorsUntouched(t *testing.T) {
	expected := errors.New("declared response")
	declared := runtime.ClientResponseReaderFunc(func(runtime.ClientResponse, runtime.Consumer) (interface{}, error) {
		return nil, expected
	})

	_, err := (&diagnosticResponseReader{next: declared}).ReadResponse(diagnosticTestResponse{code: 401}, runtime.JSONConsumer())
	assert.ErrorIs(t, err, expected)
}
