package client

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

const maxUnexpectedResponseBody = 64 << 10

// diagnosticTransport preserves response bodies for status codes omitted from
// an operation's generated response switch. The stock go-swagger fallback
// stores the ClientResponse itself in APIError; JSON formatting that value
// produces {}, hiding useful server validation errors such as HTTP 422.
type diagnosticTransport struct {
	next runtime.ClientTransport
}

// NewDiagnosticHTTPClientWithConfig creates a generated Book client that also
// retains bodies from HTTP statuses omitted by an operation's response switch.
func NewDiagnosticHTTPClientWithConfig(formats strfmt.Registry, cfg *TransportConfig) *Client {
	if cfg == nil {
		cfg = DefaultTransportConfig()
	}
	transport := httptransport.New(cfg.Host, cfg.BasePath, cfg.Schemes)
	return New(withDiagnosticResponses(transport), formats)
}

func withDiagnosticResponses(transport runtime.ClientTransport) runtime.ClientTransport {
	if _, ok := transport.(*diagnosticTransport); ok {
		return transport
	}
	return &diagnosticTransport{next: transport}
}

func (t *diagnosticTransport) Submit(operation *runtime.ClientOperation) (interface{}, error) {
	copy := *operation
	copy.Reader = &diagnosticResponseReader{next: operation.Reader}
	return t.next.Submit(&copy)
}

type diagnosticResponseReader struct {
	next runtime.ClientResponseReader
}

func (r *diagnosticResponseReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	result, err := r.next.ReadResponse(response, consumer)
	apiError, ok := err.(*runtime.APIError)
	if !ok {
		return result, err
	}
	if _, isUnreadResponse := apiError.Response.(runtime.ClientResponse); !isUnreadResponse {
		return result, err
	}

	body, readErr := io.ReadAll(io.LimitReader(response.Body(), maxUnexpectedResponseBody+1))
	payload := interface{}(map[string]interface{}{
		"message": response.Message(),
	})
	if readErr != nil {
		payload = map[string]interface{}{"message": "could not read error response"}
	} else if len(body) > maxUnexpectedResponseBody {
		payload = map[string]interface{}{"message": "error response exceeded 64 KiB"}
	} else if len(strings.TrimSpace(string(body))) > 0 {
		if json.Unmarshal(body, &payload) != nil {
			payload = map[string]interface{}{"message": strings.TrimSpace(string(body))}
		}
	}

	return nil, runtime.NewAPIError(apiError.OperationName, payload, apiError.Code)
}
