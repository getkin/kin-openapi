package openapi3filter

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers"
)

func newPetstoreRequest(t *testing.T, method, path string, body io.Reader) *http.Request {
	host := "petstore.swagger.io"
	pathPrefix := "v2"
	r, err := http.NewRequest(method, fmt.Sprintf("http://%s/%s%s", host, pathPrefix, path), body)
	require.NoError(t, err)
	r.Header.Set(headerCT, "application/json")
	r.Header.Set("Authorization", "Bearer magicstring")
	r.Host = host
	return r
}

type validationFields struct {
	Handler      http.Handler
	File         string
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
	badHost, err := http.NewRequest(http.MethodGet, "http://unknown-host.com/v2/pet", nil)
	require.NoError(t, err)
	badPath := newPetstoreRequest(t, http.MethodGet, "/watdis", nil)
	badMethod := newPetstoreRequest(t, http.MethodTrace, "/pet", nil)

	missingBody1 := newPetstoreRequest(t, http.MethodPost, "/pet", nil)
	missingBody2 := newPetstoreRequest(t, http.MethodPost, "/pet", bytes.NewBufferString(``))

	noContentType := newPetstoreRequest(t, http.MethodPost, "/pet", bytes.NewBufferString(`{}`))
	noContentType.Header.Del(headerCT)

	noContentTypeNeeded := newPetstoreRequest(t, http.MethodGet, "/pet/findByStatus?status=sold", nil)
	noContentTypeNeeded.Header.Del(headerCT)

	unknownContentType := newPetstoreRequest(t, http.MethodPost, "/pet", bytes.NewBufferString(`{}`))
	unknownContentType.Header.Set(headerCT, "application/xml")

	unsupportedContentType := newPetstoreRequest(t, http.MethodPost, "/pet", bytes.NewBufferString(`{}`))
	unsupportedContentType.Header.Set(headerCT, "text/plain")

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
			wantErrReason:   routers.ErrPathNotFound.Error(),
			wantErrResponse: &ValidationError{Status: http.StatusNotFound, Title: routers.ErrPathNotFound.Error()},
		},
		{
			name: "error - unknown path",
			args: validationArgs{
				r: badPath,
			},
			wantErrReason:   routers.ErrPathNotFound.Error(),
			wantErrResponse: &ValidationError{Status: http.StatusNotFound, Title: routers.ErrPathNotFound.Error()},
		},
		{
			name: "error - unknown method",
			args: validationArgs{
				r: badMethod,
			},
			wantErrReason: routers.ErrMethodNotAllowed.Error(),
			// TODO: By HTTP spec, this should have an Allow header with what is allowed
			// but kin-openapi doesn't provide us the requested method or path, so impossible to provide details
			wantErrResponse: &ValidationError{
				Status: http.StatusMethodNotAllowed,
				Title:  routers.ErrMethodNotAllowed.Error(),
			},
		},
		{
			name: "error - missing body on POST",
			args: validationArgs{
				r: missingBody1,
			},
			wantErrBody: "request body has an error: " + ErrInvalidRequired.Error(),
			wantErrResponse: &ValidationError{
				Status: http.StatusBadRequest,
				Title:  "request body has an error: " + ErrInvalidRequired.Error(),
			},
		},
		{
			name: "error - empty body on POST",
			args: validationArgs{
				r: missingBody2,
			},
			wantErrBody: "request body has an error: " + ErrInvalidRequired.Error(),
			wantErrResponse: &ValidationError{
				Status: http.StatusBadRequest,
				Title:  "request body has an error: " + ErrInvalidRequired.Error(),
			},
		},

		//
		// Content-Type
		//

		{
			name: "error - missing content-type on POST",
			args: validationArgs{
				r: noContentType,
			},
			wantErrReason: prefixInvalidCT + ` ""`,
			wantErrResponse: &ValidationError{
				Status: http.StatusUnsupportedMediaType,
				Title:  "header Content-Type is required",
			},
		},
		{
			name: "error - unknown content-type on POST",
			args: validationArgs{
				r: unknownContentType,
			},
			wantErrReason:      "failed to decode request body",
			wantErrParseKind:   KindUnsupportedFormat,
			wantErrParseReason: prefixUnsupportedCT + ` "application/xml"`,
			wantErrResponse: &ValidationError{
				Status: http.StatusUnsupportedMediaType,
				Title:  prefixUnsupportedCT + ` "application/xml"`,
			},
		},
		{
			name: "error - unsupported content-type on POST",
			args: validationArgs{
				r: unsupportedContentType,
			},
			wantErrReason: prefixInvalidCT + ` "text/plain"`,
			wantErrResponse: &ValidationError{
				Status: http.StatusUnsupportedMediaType,
				Title:  prefixUnsupportedCT + ` "text/plain"`,
			},
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
			name: "empty deepobject query parameter",
			args: validationArgs{
				r: newPetstoreRequest(t, http.MethodGet, "/pet/filter", nil),
			},
		},
		{
			name: "deepobject query parameter type",
			args: validationArgs{
				r: newPetstoreRequest(t, http.MethodGet, "/pet/filter?deepFilter[booleans][0]=true&deepFilter[integers][0]=1&deepFilter[strings][0]=foo%26o&deepFilter[numbers][0]=1", nil),
			},
		},
		{
			name: "error - incorrect deepobject query parameter type bool",
			args: validationArgs{
				r: newPetstoreRequest(t, http.MethodGet, "/pet/filter?deepFilter[booleans][0]=notbool", nil),
			},
			wantErrParam:   "deepFilter",
			wantErrBody:    "parameter \"deepFilter\" in query has an error: path booleans.0: value notbool: an invalid boolean: invalid syntax",
			wantErrParamIn: "query",
			wantErrResponse: &ValidationError{
				Status: http.StatusBadRequest,
				Title:  "parameter \"deepFilter\" in query is invalid: notbool is an invalid boolean",
			},
		},
		{
			name: "error - incorrect deepobject query parameter type integer",
			args: validationArgs{
				r: newPetstoreRequest(t, http.MethodGet, "/pet/filter?deepFilter[integers][0]=1.234", nil),
			},
			wantErrParam:   "deepFilter",
			wantErrBody:    "parameter \"deepFilter\" in query has an error: path integers.0: value 1.234: an invalid integer: invalid syntax",
			wantErrParamIn: "query",
			wantErrResponse: &ValidationError{
				Status: http.StatusBadRequest,
				Title:  "parameter \"deepFilter\" in query is invalid: 1.234 is an invalid integer",
			},
		},
		{
			name: "error - incorrect deepobject query parameter type integer",
			args: validationArgs{
				r: newPetstoreRequest(t, http.MethodGet, "/pet/filter?deepFilter[numbers][0]=aaa", nil),
			},
			wantErrParam:   "deepFilter",
			wantErrBody:    "parameter \"deepFilter\" in query has an error: path numbers.0: value aaa: an invalid number: invalid syntax",
			wantErrParamIn: "query",
			wantErrResponse: &ValidationError{
				Status: http.StatusBadRequest,
				Title:  "parameter \"deepFilter\" in query is invalid: aaa is an invalid number",
			},
		},
		{
			name: "error - missing required query string parameter",
			args: validationArgs{
				r: newPetstoreRequest(t, http.MethodGet, "/pet/findByStatus", nil),
			},
			wantErrParam:   "status",
			wantErrParamIn: "query",
			wantErrBody:    `parameter "status" in query has an error: value is required but missing`,
			wantErrReason:  "value is required but missing",
			wantErrResponse: &ValidationError{
				Status: http.StatusBadRequest,
				Title:  `parameter "status" in query is required`,
			},
		},
		{
			name: "error - wrong query string parameter type as integer",
			args: validationArgs{
				r: newPetstoreRequest(t, http.MethodGet, "/pet/findByIds?ids=1,notAnInt", nil),
			},
			wantErrParam:   "ids",
			wantErrParamIn: "query",
			// This is a nested ParseError. The outer error is a KindOther with no details.
			// So we'd need to look at the inner one which is a KindInvalidFormat. So just check the error body.
			wantErrBody: `parameter "ids" in query has an error: path 1: value notAnInt: an invalid integer: invalid syntax`,
			// TODO: Should we treat query params of the wrong type like a 404 instead of a 400?
			wantErrResponse: &ValidationError{
				Status: http.StatusBadRequest,
				Title:  `parameter "ids" in query is invalid: notAnInt is an invalid integer`,
			},
		},
		{
			name: "success - ignores unknown query string parameter",
			args: validationArgs{
				r: newPetstoreRequest(t, http.MethodGet, "/pet/findByStatus?status=available&wat=isdis", nil),
			},
		},
		{
			name: "error - non required query string has empty value",
			args: validationArgs{
				r: newPetstoreRequest(t, http.MethodGet, "/pets/?tags=", nil),
			},
			wantErrParam:   "tags",
			wantErrParamIn: "query",
			wantErrBody:    `parameter "tags" in query has an error: empty value is not allowed`,
			wantErrReason:  "empty value is not allowed",
			wantErrResponse: &ValidationError{
				Status: http.StatusBadRequest,
				Title:  `parameter "tags" in query is not allowed to be empty`,
			},
		},
		{
			name: "success - non required query string has empty value, but has AllowEmptyValue",
			args: validationArgs{
				r: newPetstoreRequest(t, http.MethodGet, "/pets/?status=", nil),
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
			wantErrSchemaReason: "value is not one of the allowed values [\"available\",\"pending\",\"sold\"]",
			wantErrSchemaPath:   "/0",
			wantErrSchemaValue:  "available,sold",
			wantErrResponse: &ValidationError{
				Status: http.StatusBadRequest,
				Title:  "value is not one of the allowed values [\"available\",\"pending\",\"sold\"]",
				Detail: "value available,sold at /0 must be one of: available, pending, sold; " +
					// TODO: do we really want to use this heuristic to guess
					//  that they're using the wrong serialization?
					"perhaps you intended '?status=available&status=sold'",
				Source: &ValidationErrorSource{Parameter: "status"},
			},
		},
		{
			name: "error - invalid enum value for query string parameter",
			args: validationArgs{
				r: newPetstoreRequest(t, http.MethodGet, "/pet/findByStatus?status=sold&status=watdis", nil),
			},
			wantErrParam:        "status",
			wantErrParamIn:      "query",
			wantErrSchemaReason: "value is not one of the allowed values [\"available\",\"pending\",\"sold\"]",
			wantErrSchemaPath:   "/1",
			wantErrSchemaValue:  "watdis",
			wantErrResponse: &ValidationError{
				Status: http.StatusBadRequest,
				Title:  "value is not one of the allowed values [\"available\",\"pending\",\"sold\"]",
				Detail: "value watdis at /1 must be one of: available, pending, sold",
				Source: &ValidationErrorSource{Parameter: "status"},
			},
		},
		{
			name: "error - invalid enum value, allowing commas (without 'perhaps you intended' recommendation)",
			args: validationArgs{
				// fish,with,commas isn't an enum value
				r: newPetstoreRequest(t, http.MethodGet, "/pet/findByKind?kind=dog|fish,with,commas", nil),
			},
			wantErrParam:        "kind",
			wantErrParamIn:      "query",
			wantErrSchemaReason: "value is not one of the allowed values [\"dog\",\"cat\",\"turtle\",\"bird,with,commas\"]",
			wantErrSchemaPath:   "/1",
			wantErrSchemaValue:  "fish,with,commas",
			wantErrResponse: &ValidationError{
				Status: http.StatusBadRequest,
				Title:  "value is not one of the allowed values [\"dog\",\"cat\",\"turtle\",\"bird,with,commas\"]",
				Detail: "value fish,with,commas at /1 must be one of: dog, cat, turtle, bird,with,commas",
				// No 'perhaps you intended' because its the right serialization format
				Source: &ValidationErrorSource{Parameter: "kind"},
			},
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
			wantErrSchemaReason: "value is not one of the allowed values [\"demo\",\"prod\"]",
			wantErrSchemaPath:   "/",
			wantErrSchemaValue:  "watdis",
			wantErrResponse: &ValidationError{
				Status: http.StatusBadRequest,
				Title:  "value is not one of the allowed values [\"demo\",\"prod\"]",
				Detail: "value watdis at / must be one of: demo, prod",
				Source: &ValidationErrorSource{Parameter: "x-environment"},
			},
		},

		//
		// Request bodies
		//

		{
			name: "error - invalid enum value for header object attribute",
			args: validationArgs{
				r: newPetstoreRequest(t, http.MethodPost, "/pet", bytes.NewBufferString(`{"status":"watdis"}`)),
			},
			wantErrReason:       "doesn't match schema #/components/schemas/PetWithRequired",
			wantErrSchemaReason: "value is not one of the allowed values [\"available\",\"pending\",\"sold\"]",
			wantErrSchemaValue:  "watdis",
			wantErrSchemaPath:   "/status",
			wantErrResponse: &ValidationError{
				Status: http.StatusUnprocessableEntity,
				Title:  "value is not one of the allowed values [\"available\",\"pending\",\"sold\"]",
				Detail: "value watdis at /status must be one of: available, pending, sold",
				Source: &ValidationErrorSource{Pointer: "/status"},
			},
		},
		{
			name: "error - missing required object attribute",
			args: validationArgs{
				r: newPetstoreRequest(t, http.MethodPost, "/pet", bytes.NewBufferString(`{"name":"Bahama"}`)),
			},
			wantErrReason:       "doesn't match schema #/components/schemas/PetWithRequired",
			wantErrSchemaReason: `property "photoUrls" is missing`,
			wantErrSchemaValue:  map[string]string{"name": "Bahama"},
			wantErrSchemaPath:   "/photoUrls",
			wantErrResponse: &ValidationError{
				Status: http.StatusUnprocessableEntity,
				Title:  `property "photoUrls" is missing`,
				Source: &ValidationErrorSource{Pointer: "/photoUrls"},
			},
		},
		{
			name: "error - missing required nested object attribute",
			args: validationArgs{
				r: newPetstoreRequest(t, http.MethodPost, "/pet",
					bytes.NewBufferString(`{"name":"Bahama","photoUrls":[],"category":{}}`)),
			},
			wantErrReason:       "doesn't match schema #/components/schemas/PetWithRequired",
			wantErrSchemaReason: `property "name" is missing`,
			wantErrSchemaValue:  map[string]string{},
			wantErrSchemaPath:   "/category/name",
			wantErrResponse: &ValidationError{
				Status: http.StatusUnprocessableEntity,
				Title:  `property "name" is missing`,
				Source: &ValidationErrorSource{Pointer: "/category/name"},
			},
		},
		{
			name: "error - missing required deeply nested object attribute",
			args: validationArgs{
				r: newPetstoreRequest(t, http.MethodPost, "/pet",
					bytes.NewBufferString(`{"name":"Bahama","photoUrls":[],"category":{"tags": [{}]}}`)),
			},
			wantErrReason:       "doesn't match schema #/components/schemas/PetWithRequired",
			wantErrSchemaReason: `property "name" is missing`,
			wantErrSchemaValue:  map[string]string{},
			wantErrSchemaPath:   "/category/tags/0/name",
			wantErrResponse: &ValidationError{
				Status: http.StatusUnprocessableEntity,
				Title:  `property "name" is missing`,
				Source: &ValidationErrorSource{Pointer: "/category/tags/0/name"},
			},
		},
		{
			name: "error - wrong attribute type",
			args: validationArgs{
				r: newPetstoreRequest(t, http.MethodPost, "/pet",
					bytes.NewBufferString(`{"name":"Bahama","photoUrls":"http://cat"}`)),
			},
			wantErrReason:       "doesn't match schema #/components/schemas/PetWithRequired",
			wantErrSchemaReason: "value must be an array",
			wantErrSchemaPath:   "/photoUrls",
			wantErrSchemaValue:  "http://cat",
			// TODO: this shouldn't say "or not be present", but this requires recursively resolving
			//  innerErr.JSONPointer() against e.RequestBody.Content["application/json"].Schema.Value (.Required, .Properties)
			wantErrResponse: &ValidationError{
				Status: http.StatusUnprocessableEntity,
				Title:  "value must be an array",
				Source: &ValidationErrorSource{Pointer: "/photoUrls"},
			},
		},
		{
			name: "error - missing required object attribute from allOf required overlay",
			args: validationArgs{
				r: newPetstoreRequest(t, http.MethodPost, "/pet2", bytes.NewBufferString(`{"name":"Bahama"}`)),
			},
			wantErrReason:             "doesn't match schema",
			wantErrSchemaPath:         "/",
			wantErrSchemaValue:        map[string]string{"name": "Bahama"},
			wantErrSchemaReason:       `doesn't match all schemas from "allOf"`,
			wantErrSchemaOriginReason: `property "photoUrls" is missing`,
			wantErrSchemaOriginValue:  map[string]string{"name": "Bahama"},
			wantErrSchemaOriginPath:   "/photoUrls",
			wantErrResponse: &ValidationError{
				Status: http.StatusUnprocessableEntity,
				Title:  `property "photoUrls" is missing`,
				Source: &ValidationErrorSource{Pointer: "/photoUrls"},
			},
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
			wantErrBody:    `parameter "petId" in path has an error: value is required but missing`,
			wantErrReason:  "value is required but missing",
			wantErrResponse: &ValidationError{
				Status: http.StatusBadRequest,
				Title:  `parameter "petId" in path is required`,
			},
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
			wantErrResponse: &ValidationError{
				Status: http.StatusNotFound,
				Title:  `resource not found with "petId" value: NotAnInt`,
			},
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

				if e, ok := err.(*routers.RouteError); ok {
					req.Equal(tt.wantErrReason, e.Error())
					return
				}

				e, ok := err.(*RequestError)
				req.True(ok, "not a RequestError: %T -- %#v", err, err)

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
			req.Equal(tt.wantErr, err != nil, "wantError: %v", tt.wantErr)

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
	if tt.fields.File == "" {
		tt.fields.File = "testdata/fixtures/petstore.json"
	}
	h := &ValidationHandler{
		Handler:      tt.fields.Handler,
		File:         tt.fields.File,
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
		File:         "testdata/fixtures/petstore.json",
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
		File:         "testdata/fixtures/petstore.json",
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

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
		require.Equal(t, "[422][][] value must be an array [source pointer=/photoUrls]", string(body))
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

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
		require.Equal(t, "[422][][] value must be an array [source pointer=/photoUrls]", string(body))
	})
}
