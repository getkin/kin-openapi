package openapi3filter

import (
	"context"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"net/http"
	"net/url"
)

type AuthenticationFunc func(context.Context, *AuthenticationInput) error

func NoopAuthenticationFunc(context.Context, *AuthenticationInput) error { return nil }

var _ AuthenticationFunc = NoopAuthenticationFunc

type ValidationHandler struct {
	Handler            http.Handler
	AuthenticationFunc AuthenticationFunc
	SwaggerFile        string
	ErrorEncoder       ErrorEncoder
	IgnoreServerErrors bool
	router             *Router
}

func (h *ValidationHandler) Load() error {
	h.router = NewRouter()

	err := h.LoadSwagger()
	if err != nil {
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

func (h *ValidationHandler) LoadSwagger() error {
	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromFile(h.SwaggerFile)
	if err != nil {
		return err
	}
	if h.IgnoreServerErrors {
		// remove servers from the OpenAPI spec if we shouldn't validate them
		swagger, err = h.removeServers(swagger)
		if err != nil {
			return err
		}
	}
	return h.router.AddSwagger(swagger)
}

func (h *ValidationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.validateRequest(r)
	if err != nil {
		h.ErrorEncoder(r.Context(), err, w)
		return
	}
	// TODO: validateResponse
	h.Handler.ServeHTTP(w, r)
}

func (h *ValidationHandler) validateRequest(r *http.Request) error {
	// Find route
	route, pathParams, err := h.router.FindRoute(r.Method, r.URL)
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
	err = ValidateRequest(r.Context(), requestValidationInput)
	if err != nil {
		return err
	}

	return nil
}

// removeServers remove servers from the OpenAPI spec if we shouldn't validate them.
//
// It also rewrites all the paths to begin with the server path, so that the paths still work.
// This assumes that all servers share the same path (e.g., all have /v1), or return an error.
func (h *ValidationHandler) removeServers(swagger *openapi3.Swagger) (*openapi3.Swagger, error) {
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
