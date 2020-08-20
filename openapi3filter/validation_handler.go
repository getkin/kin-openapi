package openapi3filter

import (
	"context"
	"net/http"
)

type AuthenticationFunc func(context.Context, *AuthenticationInput) error

func NoopAuthenticationFunc(context.Context, *AuthenticationInput) error { return nil }

var _ AuthenticationFunc = NoopAuthenticationFunc

type ValidationHandler struct {
	Handler            http.Handler
	AuthenticationFunc AuthenticationFunc
	SwaggerFile        string
	ErrorEncoder       ErrorEncoder
	router             *Router
}

func (h *ValidationHandler) Load() error {
	h.router = NewRouter()

	if err := h.router.AddSwaggerFromFile(h.SwaggerFile); err != nil {
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
	if err = ValidateRequest(r.Context(), requestValidationInput); err != nil {
		return err
	}

	return nil
}
