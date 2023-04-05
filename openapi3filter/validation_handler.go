package openapi3filter

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers"
	legacyrouter "github.com/getkin/kin-openapi/routers/legacy"
)

// AuthenticationFunc allows for custom security requirement validation.
// A non-nil error fails authentication according to https://spec.openapis.org/oas/v3.1.0#security-requirement-object
// See ValidateSecurityRequirements
type AuthenticationFunc func(context.Context, *AuthenticationInput) error

// NoopAuthenticationFunc is an AuthenticationFunc
func NoopAuthenticationFunc(context.Context, *AuthenticationInput) error { return nil }

var _ AuthenticationFunc = NoopAuthenticationFunc

type ValidationHandler struct {
	Handler            http.Handler
	AuthenticationFunc AuthenticationFunc
	File               string
	ErrorEncoder       ErrorEncoder
	router             routers.Router
}

func (h *ValidationHandler) Load() error {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile(h.File)
	if err != nil {
		return err
	}
	if err := doc.Validate(loader.Context); err != nil {
		return err
	}
	if h.router, err = legacyrouter.NewRouter(doc); err != nil {
		return err
	}

	// set defaults
	if h.Handler == nil {
		h.Handler = http.DefaultServeMux
	}
	if h.AuthenticationFunc == nil {
		h.AuthenticationFunc = NoopAuthenticationFunc
	}
	if h.ErrorEncoder == nil {
		h.ErrorEncoder = DefaultErrorEncoder
	}

	return nil
}

func (h *ValidationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if handled := h.before(w, r); handled {
		return
	}
	// TODO: validateResponse
	h.Handler.ServeHTTP(w, r)
}

// Middleware implements gorilla/mux MiddlewareFunc
func (h *ValidationHandler) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if handled := h.before(w, r); handled {
			return
		}
		// TODO: validateResponse
		next.ServeHTTP(w, r)
	})
}

func (h *ValidationHandler) before(w http.ResponseWriter, r *http.Request) (handled bool) {
	if err := h.validateRequest(r); err != nil {
		h.ErrorEncoder(r.Context(), err, w)
		return true
	}
	return false
}

func (h *ValidationHandler) validateRequest(r *http.Request) error {
	// Find route
	route, pathParams, err := h.router.FindRoute(r)
	if err != nil {
		return err
	}

	options := &Options{
		AuthenticationFunc: h.AuthenticationFunc,
	}

	// Validate request
	requestValidationInput := &RequestValidationInput{
		Request:    r,
		PathParams: pathParams,
		Route:      route,
		Options:    options,
	}
	if err = ValidateRequest(r.Context(), requestValidationInput); err != nil {
		return err
	}

	return nil
}

// removeServers remove servers from the OpenAPI spec if we shouldn't validate them.
//
// It also rewrites all the paths to begin with the server path, so that the paths still work.
// This assumes that all servers share the same path (e.g., all have /v1), or return an error.
func (h *ValidationHandler) removeServers(swagger *openapi3.T) (*openapi3.T, error) {
	// collect API pathPrefix path prefixes
	prefixes := make(map[string]struct{}, 0) // a "set"
	for _, s := range swagger.Servers {
		u, err := url.Parse(s.URL)
		if err != nil {
			return nil, err
		}
		prefixes[u.Path] = struct{}{}
	}
	if len(prefixes) != 1 {
		return nil, fmt.Errorf("requires a single API pathPrefix path prefix: %v", prefixes)
	}
	var prefix string
	for k := range prefixes {
		prefix = k
	}

	// update the paths to start with the API pathPrefix path prefixes
	paths := make(openapi3.Paths, 0)
	for key, path := range swagger.Paths {
		paths[prefix+key] = path
	}
	swagger.Paths = paths

	// now remove the servers
	swagger.Servers = nil

	return swagger, nil
}
