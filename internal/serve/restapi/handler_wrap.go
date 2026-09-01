package restapi

import "net/http"

// WrapHandler adds narrowly scoped HTTP handling around the generated API.
func (s *Server) WrapHandler(wrap func(http.Handler) http.Handler) {
	if wrap != nil {
		s.handler = wrap(s.handler)
	}
}
