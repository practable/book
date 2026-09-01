package restapi

import "net/http"

// WrapHandler adds narrowly scoped HTTP handling around the generated API
// without making generated routing code responsible for protocol extensions.
func (s *Server) WrapHandler(wrap func(http.Handler) http.Handler) {
	if wrap != nil {
		s.handler = wrap(s.handler)
	}
}
