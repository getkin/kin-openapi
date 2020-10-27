package openapi3filter

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func newPetstoreRequest(t *testing.T, method, path string, body io.Reader) *http.Request {
	host := "petstore.swagger.io"
	pathPrefix := "v2"
	r, err := http.NewRequest(method, fmt.Sprintf("http://%s/%s%s", host, pathPrefix, path), body)
	require.NoError(t, err)
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", "Bearer magicstring")
	r.Host = host
	return r
}

type validationFields struct {
	Handler      http.Handler
	SwaggerFile  string
	ErrorEncoder ErrorEncoder
}
type validationArgs struct {
	r *http.Request
}
type validationTest struct {
	name                      string
	fields                    validationFields
	args                      validationArgs
	wantErr                   bool
	wantErrBody               string
	wantErrReason             string
	wantErrSchemaReason       string
	wantErrSchemaPath         string
	wantErrSchemaValue        interface{}
	wantErrSchemaOriginReason string
	wantErrSchemaOriginPath   string
	wantErrSchemaOriginValue  interface{}
	wantErrParam              string
	wantErrParamIn            string
	wantErrParseKind          ParseErrorKind
	wantErrParseValue         interface{}
	wantErrParseReason        string
	wantErrResponse           *ValidationError
}

func getValidationTests(t *testing.T) []*validationTest {
	badHost, _ := http.NewRequest(http.MethodGet, "http://unknown-host.com/v2/pet", nil)
	badPath := newPetstoreRequest(t, http.MethodGet, "/watdis", nil)
	badMethod := newPetstoreRequest(t, http.MethodTrace, "/pet", nil)

	missingBody1 := newPetstoreRequest(t, http.MethodPost, "/pet", nil)
	missingBody2 := newPetstoreRequest(t, http.MethodPost, "/pet", bytes.NewBufferString(``))

	noContentType := newPetstoreRequest(t, http.MethodPost, "/pet", bytes.NewBufferString(`{}`))
	noContentType.Header.Del("Content-Type")

	noContentTypeNeeded := newPetstoreRequest(t, http.MethodGet, "/pet/findByStatus?status=sold", nil)
	noContentTypeNeeded.Header.Del("Content-Type")

	unknownContentType := newPetstoreRequest(t, http.MethodPost, "/pet", bytes.NewBufferString(`{}`))
	unknownContentType.Header.Set("Content-Type", "application/xml")

	unsupportedContentType := newPetstoreRequest(t, http.MethodPost, "/pet", bytes.NewBufferString(`{}`))
	unsupportedContentType.Header.Set("Content-Type", "text/plain")

	unsupportedHeaderValue := newPetstoreRequest(t, http.MethodPost, "/pet", bytes.NewBufferString(`{}`))
	unsupportedHeaderValue.Header.Set("x-environment", "watdis")

	return []*validationTest{
		//
		// Basics
		//

		{
			name: "error - unknown host",
			args: validationArgs{
				r: badHost,
			},
			wantErrReason:   "Does not match any server",
			wantErrResponse: &ValidationError{Status: http.StatusNotFound, Title: "Does not match any server"},
		},
		{
			name: "error - unknown path",
			args: validationArgs{
				r: badPath,
			},
			wantErrReason:   "Path was not found",
			wantErrResponse: &ValidationError{Status: http.StatusNotFound, Title: "Path was not found"},
		},
		{
			name: "error - unknown method",
			args: validationArgs{
				r: badMethod,
			},
			wantErrReason: "Path doesn't support the HTTP method",
			// TODO: By HTTP spec, this should have an Allow header with what is allowed
			// but kin-openapi doesn't provide us the requested method or path, so impossible to provide details
			wantErrResponse: &ValidationError{Status: http.StatusMethodNotAllowed,
				Title: "Path doesn't support the HTTP method"},
		},
		{
			name: "error - missing body on POST",
			args: validationArgs{
				r: missingBody1,
			},
			wantErrBody: "Request body has an error: must have a value",
			wantErrResponse: &ValidationError{Status: http.StatusBadRequest,
				Title: "Request body has an error: must have a value"},
		},
		{
			name: "error - empty body on POST",
			args: validationArgs{
				r: missingBody2,
			},
			wantErrBody: "Request body has an error: must have a value",
			wantErrResponse: &ValidationError{Status: http.StatusBadRequest,
				Title: "Request body has an error: must have a value"},
		},

		//
		// Content-Type
		//

		{
			name: "error - missing content-type on POST",
			args: validationArgs{
				r: noContentType,
			},
			wantErrReason: "header 'Content-Type' has unexpected value: \"\"",
			wantErrResponse: &ValidationError{Status: http.StatusUnsupportedMediaType,
				Title: "header 'Content-Type' is required"},
		},
		{
			name: "error - unknown content-type on POST",
			args: validationArgs{
				r: unknownContentType,
			},
			wantErrReason:      "failed to decode request body",
			wantErrParseKind:   KindUnsupportedFormat,
			wantErrParseReason: "unsupported content type \"application/xml\"",
			wantErrResponse: &ValidationError{Status: http.StatusUnsupportedMediaType,
				Title: "unsupported content type \"application/xml\""},
		},
		{
			name: "error - unsupported content-type on POST",
			args: validationArgs{
				r: unsupportedContentType,
			},
			wantErrReason: "header 'Content-Type' has unexpected value: \"text/plain\"",
			wantErrResponse: &ValidationError{Status: http.StatusUnsupportedMediaType,
				Title: "unsupported content type \"text/plain\""},
		},
		{
			name: "success - no content-type header required on GET",
			args: validationArgs{
				r: noContentTypeNeeded,
			},
		},

		//
		// Query strings
		//

		{
			name: "error - missing required query string parameter",
			args: validationArgs{
				r: newPetstoreRequest(t, http.MethodGet, "/pet/findByStatus", nil),
			},
			wantErrParam:   "status",
			wantErrParamIn: "query",
			wantErrReason:  "must have a value",
			wantErrResponse: &ValidationError{Status: http.StatusBadRequest,
				Title: "Parameter 'status' in query is required"},
		},
		{
			name: "error - wrong query string parameter type",
			args: validationArgs{
				r: newPetstoreRequest(t, http.MethodGet, "/pet/findByIds?ids=1,notAnInt", nil),
			},
			wantErrParam:   "ids",
			wantErrParamIn: "query",
			// This is a nested ParseError. The outer error is a KindOther with no details.
			// So we'd need to look at the inner one which is a KindInvalidFormat. So just check the error body.
			wantErrBody: "Parameter 'ids' in query has an error: path 1: value notAnInt: an invalid integer: " +
				"strconv.ParseFloat: parsing \"notAnInt\": invalid syntax",
			// TODO: Should we treat query params of the wrong type like a 404 instead of a 400?
			wantErrResponse: &ValidationError{Status: http.StatusBadRequest,
				Title: "Parameter 'ids' in query is invalid: notAnInt is an invalid integer"},
		},
		{
			name: "success - ignores unknown query string parameter",
			args: validationArgs{
				r: newPetstoreRequest(t, http.MethodGet, "/pet/findByStatus?wat=isdis", nil),
			},
		},
		{
			name: "success - normal case, query strings",
			args: validationArgs{
				r: newPetstoreRequest(t, http.MethodGet, "/pet/findByStatus?status=available", nil),
			},
		},
		{
			name: "success - normal case, query strings, array",
			args: validationArgs{
				r: newPetstoreRequest(t, http.MethodGet, "/pet/findByStatus?status=available&status=sold", nil),
			},
		},
		{
			name: "error - invalid query string array serialization",
			args: validationArgs{
				r: newPetstoreRequest(t, http.MethodGet, "/pet/findByStatus?status=available,sold", nil),
			},
			wantErrParam:        "status",
			wantErrParamIn:      "query",
			wantErrSchemaReason: "JSON value is not one of the allowed values",
			wantErrSchemaPath:   "/0",
			wantErrSchemaValue:  "available,sold",
			wantErrResponse: &ValidationError{Status: http.StatusBadRequest,
				Title: "JSON value is not one of the allowed values",
				Detail: "Value 'available,sold' at /0 must be one of: available, pending, sold; " +
					// TODO: do we really want to use this heuristic to guess
					//  that they're using the wrong serialization?
					"perhaps you intended '?status=available&status=sold'",
				Source: &ValidationErrorSource{Parameter: "status"}},
		},
		{
			name: "error - invalid enum value for query string parameter",
			args: validationArgs{
				r: newPetstoreRequest(t, http.MethodGet, "/pet/findByStatus?status=sold&status=watdis", nil),
			},
			wantErrParam:        "status",
			wantErrParamIn:      "query",
			wantErrSchemaReason: "JSON value is not one of the allowed values",
			wantErrSchemaPath:   "/1",
			wantErrSchemaValue:  "watdis",
			wantErrResponse: &ValidationError{Status: http.StatusBadRequest,
				Title:  "JSON value is not one of the allowed values",
				Detail: "Value 'watdis' at /1 must be one of: available, pending, sold",
				Source: &ValidationErrorSource{Parameter: "status"}},
		},
		{
			name: "error - invalid enum value, allowing commas (without 'perhaps you intended' recommendation)",
			args: validationArgs{
				// fish,with,commas isn't an enum value
				r: newPetstoreRequest(t, http.MethodGet, "/pet/findByKind?kind=dog|fish,with,commas", nil),
			},
			wantErrParam:        "kind",
			wantErrParamIn:      "query",
			wantErrSchemaReason: "JSON value is not one of the allowed values",
			wantErrSchemaPath:   "/1",
			wantErrSchemaValue:  "fish,with,commas",
			wantErrResponse: &ValidationError{Status: http.StatusBadRequest,
				Title:  "JSON value is not one of the allowed values",
				Detail: "Value 'fish,with,commas' at /1 must be one of: dog, cat, turtle, bird,with,commas",
				// No 'perhaps you intended' because its the right serialization format
				Source: &ValidationErrorSource{Parameter: "kind"}},
		},
		{
			name: "success - valid enum value, allowing commas",
			args: validationArgs{
				r: newPetstoreRequest(t, http.MethodGet, "/pet/findByKind?kind=dog|bird,with,commas", nil),
			},
		},

		//
		// Request header params
		//
		{
			name: "error - invalid enum value for header string parameter",
			args: validationArgs{
				r: unsupportedHeaderValue,
			},
			wantErrParam:        "x-environment",
			wantErrParamIn:      "header",
			wantErrSchemaReason: "JSON value is not one of the allowed values",
			wantErrSchemaPath:   "/",
			wantErrSchemaValue:  "watdis",
			wantErrResponse: &ValidationError{Status: http.StatusBadRequest,
				Title:  "JSON value is not one of the allowed values",
				Detail: "Value 'watdis' at / must be one of: demo, prod",
				Source: &ValidationErrorSource{Parameter: "x-environment"}},
		},

		//
		// Request bodies
		//

		{
			name: "error - invalid enum value for header object attribute",
			args: validationArgs{
				r: newPetstoreRequest(t, http.MethodPost, "/pet", bytes.NewBufferString(`{"status":"watdis"}`)),
			},
			wantErrReason:       "doesn't match the schema",
			wantErrSchemaReason: "JSON value is not one of the allowed values",
			wantErrSchemaValue:  "watdis",
			wantErrSchemaPath:   "/status",
			wantErrResponse: &ValidationError{Status: http.StatusUnprocessableEntity,
				Title:  "JSON value is not one of the allowed values",
				Detail: "Value 'watdis' at /status must be one of: available, pending, sold",
				Source: &ValidationErrorSource{Pointer: "/status"}},
		},
		{
			name: "error - missing required object attribute",
			args: validationArgs{
				r: newPetstoreRequest(t, http.MethodPost, "/pet", bytes.NewBufferString(`{"name":"Bahama"}`)),
			},
			wantErrReason:       "doesn't match the schema",
			wantErrSchemaReason: "Property 'photoUrls' is missing",
			wantErrSchemaValue:  map[string]string{"name": "Bahama"},
			wantErrSchemaPath:   "/photoUrls",
			wantErrResponse: &ValidationError{Status: http.StatusUnprocessableEntity,
				Title:  "Property 'photoUrls' is missing",
				Source: &ValidationErrorSource{Pointer: "/photoUrls"}},
		},
		{
			name: "error - missing required nested object attribute",
			args: validationArgs{
				r: newPetstoreRequest(t, http.MethodPost, "/pet",
					bytes.NewBufferString(`{"name":"Bahama","photoUrls":[],"category":{}}`)),
			},
			wantErrReason:       "doesn't match the schema",
			wantErrSchemaReason: "Property 'name' is missing",
			wantErrSchemaValue:  map[string]string{},
			wantErrSchemaPath:   "/category/name",
			wantErrResponse: &ValidationError{Status: http.StatusUnprocessableEntity,
				Title:  "Property 'name' is missing",
				Source: &ValidationErrorSource{Pointer: "/category/name"}},
		},
		{
			name: "error - missing required deeply nested object attribute",
			args: validationArgs{
				r: newPetstoreRequest(t, http.MethodPost, "/pet",
					bytes.NewBufferString(`{"name":"Bahama","photoUrls":[],"category":{"tags": [{}]}}`)),
			},
			wantErrReason:       "doesn't match the schema",
			wantErrSchemaReason: "Property 'name' is missing",
			wantErrSchemaValue:  map[string]string{},
			wantErrSchemaPath:   "/category/tags/0/name",
			wantErrResponse: &ValidationError{Status: http.StatusUnprocessableEntity,
				Title:  "Property 'name' is missing",
				Source: &ValidationErrorSource{Pointer: "/category/tags/0/name"}},
		},
		{
			// TODO: Add support for validating readonly properties to upstream validator.
			name: "error - readonly object attribute",
			args: validationArgs{
				r: newPetstoreRequest(t, http.MethodPost, "/pet",
					bytes.NewBufferString(`{"id":213,"name":"Bahama","photoUrls":[]}}`)),
			},
			//wantErr: true,
		},
		{
			name: "error - wrong attribute type",
			args: validationArgs{
				r: newPetstoreRequest(t, http.MethodPost, "/pet",
					bytes.NewBufferString(`{"name":"Bahama","photoUrls":"http://cat"}`)),
			},
			wantErrReason:       "doesn't match the schema",
			wantErrSchemaReason: "Field must be set to array or not be present",
			wantErrSchemaPath:   "/photoUrls",
			wantErrSchemaValue:  "string",
			// TODO: this shouldn't say "or not be present", but this requires recursively resolving
			//  innerErr.JSONPointer() against e.RequestBody.Content["application/json"].Schema.Value (.Required, .Properties)
			wantErrResponse: &ValidationError{Status: http.StatusUnprocessableEntity,
				Title:  "Field must be set to array or not be present",
				Source: &ValidationErrorSource{Pointer: "/photoUrls"}},
		},
		{
			name: "error - missing required object attribute from allOf required overlay",
			args: validationArgs{
				r: newPetstoreRequest(t, http.MethodPost, "/pet2", bytes.NewBufferString(`{"name":"Bahama"}`)),
			},
			wantErrReason:             "doesn't match the schema",
			wantErrSchemaPath:         "/",
			wantErrSchemaValue:        map[string]string{"name": "Bahama"},
			wantErrSchemaOriginReason: "Property 'photoUrls' is missing",
			wantErrSchemaOriginValue:  map[string]string{"name": "Bahama"},
			wantErrSchemaOriginPath:   "/photoUrls",
			wantErrResponse: &ValidationError{Status: http.StatusUnprocessableEntity,
				Title:  "Property 'photoUrls' is missing",
				Source: &ValidationErrorSource{Pointer: "/photoUrls"}},
		},
		{
			name: "success - ignores unknown object attribute",
			args: validationArgs{
				r: newPetstoreRequest(t, http.MethodPost, "/pet",
					bytes.NewBufferString(`{"wat":"isdis","name":"Bahama","photoUrls":[]}`)),
			},
		},
		{
			name: "success - normal case, POST",
			args: validationArgs{
				r: newPetstoreRequest(t, http.MethodPost, "/pet",
					bytes.NewBufferString(`{"name":"Bahama","photoUrls":[]}`)),
			},
		},
		{
			name: "success - required properties are not required on PATCH if required overlaid using allOf elsewhere",
			args: validationArgs{
				r: newPetstoreRequest(t, http.MethodPatch, "/pet", bytes.NewBufferString(`{}`)),
			},
		},

		//
		// Path params
		//

		{
			name: "error - missing path param",
			args: validationArgs{
				r: newPetstoreRequest(t, http.MethodGet, "/pet/", nil),
			},
			wantErrParam:   "petId",
			wantErrParamIn: "path",
			wantErrReason:  "must have a value",
			wantErrResponse: &ValidationError{Status: http.StatusBadRequest,
				Title: "Parameter 'petId' in path is required"},
		},
		{
			name: "error - wrong path param type",
			args: validationArgs{
				r: newPetstoreRequest(t, http.MethodGet, "/pet/NotAnInt", nil),
			},
			wantErrParam:       "petId",
			wantErrParamIn:     "path",
			wantErrParseKind:   KindInvalidFormat,
			wantErrParseValue:  "NotAnInt",
			wantErrParseReason: "an invalid integer",
			wantErrResponse: &ValidationError{Status: http.StatusNotFound,
				Title: "Resource not found with 'petId' value: NotAnInt"},
		},
		{
			name: "success - normal case, with path params",
			args: validationArgs{
				r: newPetstoreRequest(t, http.MethodGet, "/pet/23", nil),
			},
		},
	}
}

func TestValidationHandler_validateRequest(t *testing.T) {
	tests := getValidationTests(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := require.New(t)

			h, err := buildValidationHandler(tt)
			req.NoError(err)

			err = h.validateRequest(tt.args.r)
			req.Equal(tt.wantErr, err != nil)

			if err != nil {
				if tt.wantErrBody != "" {
					req.Equal(tt.wantErrBody, err.Error())
				}

				if e, ok := err.(*RouteError); ok {
					req.Equal(tt.wantErrReason, e.Reason)
					return
				}

				e, ok := err.(*RequestError)
				req.True(ok, "error = %v, not a RequestError -- %#v", err, err)

				req.Equal(tt.wantErrReason, e.Reason)

				if e.Parameter != nil {
					req.Equal(tt.wantErrParam, e.Parameter.Name)
					req.Equal(tt.wantErrParamIn, e.Parameter.In)
				} else {
					req.False(tt.wantErrParam != "" || tt.wantErrParamIn != "",
						"error = %v, no Parameter -- %#v", e, e)
				}

				if innerErr, ok := e.Err.(*openapi3.SchemaError); ok {
					req.Equal(tt.wantErrSchemaReason, innerErr.Reason)
					pointer := toJSONPointer(innerErr.JSONPointer())
					req.Equal(tt.wantErrSchemaPath, pointer)
					req.Equal(fmt.Sprintf("%v", tt.wantErrSchemaValue), fmt.Sprintf("%v", innerErr.Value))

					if originErr, ok := innerErr.Origin.(*openapi3.SchemaError); ok {
						req.Equal(tt.wantErrSchemaOriginReason, originErr.Reason)
						pointer := toJSONPointer(originErr.JSONPointer())
						req.Equal(tt.wantErrSchemaOriginPath, pointer)
						req.Equal(fmt.Sprintf("%v", tt.wantErrSchemaOriginValue), fmt.Sprintf("%v", originErr.Value))
					}
				} else {
					req.False(tt.wantErrSchemaReason != "" || tt.wantErrSchemaPath != "",
						"error = %v, not a SchemaError -- %#v", e.Err, e.Err)
					req.False(tt.wantErrSchemaOriginReason != "" || tt.wantErrSchemaOriginPath != "",
						"error = %v, not a SchemaError with Origin -- %#v", e.Err, e.Err)
				}

				if innerErr, ok := e.Err.(*ParseError); ok {
					req.Equal(tt.wantErrParseKind, innerErr.Kind)
					req.Equal(tt.wantErrParseValue, innerErr.Value)
					req.Equal(tt.wantErrParseReason, innerErr.Reason)
				} else {
					req.False(tt.wantErrParseValue != nil || tt.wantErrParseReason != "",
						"error = %v, not a ParseError -- %#v", e.Err, e.Err)
				}
			}
		})
	}
}

func TestValidationErrorEncoder(t *testing.T) {
	tests := getValidationTests(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEncoder := &mockErrorEncoder{}
			encoder := &ValidationErrorEncoder{Encoder: mockEncoder.Encode}

			req := require.New(t)

			h, err := buildValidationHandler(tt)
			req.NoError(err)

			err = h.validateRequest(tt.args.r)
			req.Equal(tt.wantErr, err != nil)

			if err != nil {
				encoder.Encode(tt.args.r.Context(), err, httptest.NewRecorder())
				if tt.wantErrResponse != mockEncoder.Err {
					req.Equal(tt.wantErrResponse, mockEncoder.Err)
				}
			}
		})
	}
}

func buildValidationHandler(tt *validationTest) (*ValidationHandler, error) {
	if tt.fields.SwaggerFile == "" {
		tt.fields.SwaggerFile = "fixtures/petstore.json"
	}
	h := &ValidationHandler{
		Handler:      tt.fields.Handler,
		SwaggerFile:  tt.fields.SwaggerFile,
		ErrorEncoder: tt.fields.ErrorEncoder,
	}
	tt.wantErr = tt.wantErr ||
		(tt.wantErrBody != "") ||
		(tt.wantErrReason != "") ||
		(tt.wantErrSchemaReason != "") ||
		(tt.wantErrSchemaPath != "") ||
		(tt.wantErrParseValue != nil) ||
		(tt.wantErrParseReason != "")
	err := h.Load()
	return h, err
}

type testHandler struct {
	Called bool
}

func (h *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Called = true
}

type mockErrorEncoder struct {
	Called bool
	Ctx    context.Context
	Err    error
	W      http.ResponseWriter
}

func (e *mockErrorEncoder) Encode(ctx context.Context, err error, w http.ResponseWriter) {
	e.Called = true
	e.Ctx = ctx
	e.Err = err
	e.W = w
}

func runTest_ServeHTTP(t *testing.T, handler http.Handler, encoder ErrorEncoder, req *http.Request) *http.Response {
	h := &ValidationHandler{
		Handler:      handler,
		ErrorEncoder: encoder,
		SwaggerFile:  "fixtures/petstore.json",
	}
	err := h.Load()
	require.NoError(t, err)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w.Result()
}

func runTest_Middleware(t *testing.T, handler http.Handler, encoder ErrorEncoder, req *http.Request) *http.Response {
	h := &ValidationHandler{
		ErrorEncoder: encoder,
		SwaggerFile:  "fixtures/petstore.json",
	}
	err := h.Load()
	require.NoError(t, err)
	w := httptest.NewRecorder()
	h.Middleware(handler).ServeHTTP(w, req)
	return w.Result()
}

func TestValidationHandler_ServeHTTP(t *testing.T) {
	t.Run("errors on invalid requests", func(t *testing.T) {
		httpCtx := context.WithValue(context.Background(), "pig", "tails")
		r, err := http.NewRequest(http.MethodGet, "http://unknown-host.com/v2/pet", nil)
		require.NoError(t, err)
		r = r.WithContext(httpCtx)

		handler := &testHandler{}
		encoder := &mockErrorEncoder{}
		runTest_ServeHTTP(t, handler, encoder.Encode, r)

		require.False(t, handler.Called)
		require.True(t, encoder.Called)
		require.Equal(t, httpCtx, encoder.Ctx)
		require.NotNil(t, encoder.Err)
	})

	t.Run("passes valid requests through", func(t *testing.T) {
		r := newPetstoreRequest(t, http.MethodGet, "/pet/findByStatus?status=sold", nil)

		handler := &testHandler{}
		encoder := &mockErrorEncoder{}
		runTest_ServeHTTP(t, handler, encoder.Encode, r)

		require.True(t, handler.Called)
		require.False(t, encoder.Called)
	})

	t.Run("uses error encoder", func(t *testing.T) {
		r := newPetstoreRequest(t, http.MethodPost, "/pet", bytes.NewBufferString(`{"name":"Bahama","photoUrls":"http://cat"}`))

		handler := &testHandler{}
		encoder := &ValidationErrorEncoder{Encoder: (ErrorEncoder)(DefaultErrorEncoder)}
		resp := runTest_ServeHTTP(t, handler, encoder.Encode, r)

		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
		require.Equal(t, "[422][][] Field must be set to array or not be present [source pointer=/photoUrls]", string(body))
	})
}

func TestValidationHandler_Middleware(t *testing.T) {
	t.Run("errors on invalid requests", func(t *testing.T) {
		httpCtx := context.WithValue(context.Background(), "pig", "tails")
		r, err := http.NewRequest(http.MethodGet, "http://unknown-host.com/v2/pet", nil)
		require.NoError(t, err)
		r = r.WithContext(httpCtx)

		handler := &testHandler{}
		encoder := &mockErrorEncoder{}
		runTest_Middleware(t, handler, encoder.Encode, r)

		require.False(t, handler.Called)
		require.True(t, encoder.Called)
		require.Equal(t, httpCtx, encoder.Ctx)
		require.NotNil(t, encoder.Err)
	})

	t.Run("passes valid requests through", func(t *testing.T) {
		r := newPetstoreRequest(t, http.MethodGet, "/pet/findByStatus?status=sold", nil)

		handler := &testHandler{}
		encoder := &mockErrorEncoder{}
		runTest_Middleware(t, handler, encoder.Encode, r)

		require.True(t, handler.Called)
		require.False(t, encoder.Called)
	})

	t.Run("uses error encoder", func(t *testing.T) {
		r := newPetstoreRequest(t, http.MethodPost, "/pet", bytes.NewBufferString(`{"name":"Bahama","photoUrls":"http://cat"}`))

		handler := &testHandler{}
		encoder := &ValidationErrorEncoder{Encoder: (ErrorEncoder)(DefaultErrorEncoder)}
		resp := runTest_Middleware(t, handler, encoder.Encode, r)

		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
		require.Equal(t, "[422][][] Field must be set to array or not be present [source pointer=/photoUrls]", string(body))
	})
}
