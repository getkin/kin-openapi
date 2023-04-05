package openapi3filter_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

const validatorSpec = `
openapi: 3.0.0
info:
  title: 'Validator'
  version: '0.0.0'
paths:
  /test:
    post:
      operationId: newTest
      description: create a new test
      parameters:
        - in: query
          name: version
          schema:
            type: string
          required: true
      requestBody:
        required: true
        content:
          application/json:
            schema: { $ref: '#/components/schemas/TestContents' }
      responses:
        '201':
          description: 'created test'
          content:
            application/json:
              schema: { $ref: '#/components/schemas/TestResource' }
        '400': { $ref: '#/components/responses/ErrorResponse' }
        '500': { $ref: '#/components/responses/ErrorResponse' }
  /test/{id}:
    get:
      operationId: getTest
      description: get a test
      parameters:
        - in: path
          name: id
          schema:
            type: string
          required: true
        - in: query
          name: version
          schema:
            type: string
          required: true
      responses:
        '200':
          description: 'respond with test resource'
          content:
            application/json:
              schema: { $ref: '#/components/schemas/TestResource' }
        '400': { $ref: '#/components/responses/ErrorResponse' }
        '404': { $ref: '#/components/responses/ErrorResponse' }
        '500': { $ref: '#/components/responses/ErrorResponse' }
components:
  schemas:
    TestContents:
      type: object
      properties:
        name:
          type: string
        expected:
          type: number
        actual:
          type: number
      required: [name, expected, actual]
      additionalProperties: false
    TestResource:
      type: object
      properties:
        id:
          type: string
        contents:
          { $ref: '#/components/schemas/TestContents' }
      required: [id, contents]
      additionalProperties: false
    Error:
      type: object
      properties:
        code:
          type: string
        message:
          type: string
      required: [code, message]
      additionalProperties: false
  responses:
    ErrorResponse:
      description: 'an error occurred'
      content:
        application/json:
          schema: { $ref: '#/components/schemas/Error' }
`

type validatorTestHandler struct {
	contentType       string
	getBody, postBody string
	errBody           string
	errStatusCode     int
}

const validatorOkResponse = `{"id": "42", "contents": {"name": "foo", "expected": 9, "actual": 10}}`

func (h validatorTestHandler) withDefaults() validatorTestHandler {
	if h.contentType == "" {
		h.contentType = "application/json"
	}
	if h.getBody == "" {
		h.getBody = validatorOkResponse
	}
	if h.postBody == "" {
		h.postBody = validatorOkResponse
	}
	if h.errBody == "" {
		h.errBody = `{"code":"bad","message":"bad things"}`
	}
	return h
}

var testUrlRE = regexp.MustCompile(`^/test(/\d+)?$`)

func (h *validatorTestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", h.contentType)
	if h.errStatusCode != 0 {
		w.WriteHeader(h.errStatusCode)
		w.Write([]byte(h.errBody))
		return
	}
	if !testUrlRE.MatchString(r.URL.Path) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(h.errBody))
		return
	}
	switch r.Method {
	case "GET":
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(h.getBody))
	case "POST":
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(h.postBody))
	default:
		http.Error(w, h.errBody, http.StatusMethodNotAllowed)
	}
}

func TestValidator(t *testing.T) {
	doc, err := openapi3.NewLoader().LoadFromData([]byte(validatorSpec))
	require.NoError(t, err, "failed to load test fixture spec")

	ctx := context.Background()
	err = doc.Validate(ctx)
	require.NoError(t, err, "invalid test fixture spec")

	type testRequest struct {
		method, path, body, contentType string
	}
	type testResponse struct {
		statusCode int
		body       string
	}
	tests := []struct {
		name     string
		handler  validatorTestHandler
		options  []openapi3filter.ValidatorOption
		request  testRequest
		response testResponse
		strict   bool
	}{{
		name:    "valid GET",
		handler: validatorTestHandler{}.withDefaults(),
		request: testRequest{
			method: "GET",
			path:   "/test/42?version=1",
		},
		response: testResponse{
			200, validatorOkResponse,
		},
		strict: true,
	}, {
		name:    "valid POST",
		handler: validatorTestHandler{}.withDefaults(),
		request: testRequest{
			method:      "POST",
			path:        "/test?version=1",
			body:        `{"name": "foo", "expected": 9, "actual": 10}`,
			contentType: "application/json",
		},
		response: testResponse{
			201, validatorOkResponse,
		},
		strict: true,
	}, {
		name:    "not found; no GET operation for /test",
		handler: validatorTestHandler{}.withDefaults(),
		request: testRequest{
			method: "GET",
			path:   "/test?version=1",
		},
		response: testResponse{
			404, "not found\n",
		},
		strict: true,
	}, {
		name:    "not found; no POST operation for /test/42",
		handler: validatorTestHandler{}.withDefaults(),
		request: testRequest{
			method: "POST",
			path:   "/test/42?version=1",
		},
		response: testResponse{
			404, "not found\n",
		},
		strict: true,
	}, {
		name:    "invalid request; missing version",
		handler: validatorTestHandler{}.withDefaults(),
		request: testRequest{
			method: "GET",
			path:   "/test/42",
		},
		response: testResponse{
			400, "bad request\n",
		},
		strict: true,
	}, {
		name:    "invalid POST request; wrong property type",
		handler: validatorTestHandler{}.withDefaults(),
		request: testRequest{
			method:      "POST",
			path:        "/test?version=1",
			body:        `{"name": "foo", "expected": "nine", "actual": "ten"}`,
			contentType: "application/json",
		},
		response: testResponse{
			400, "bad request\n",
		},
		strict: true,
	}, {
		name:    "invalid POST request; missing property",
		handler: validatorTestHandler{}.withDefaults(),
		request: testRequest{
			method:      "POST",
			path:        "/test?version=1",
			body:        `{"name": "foo", "expected": 9}`,
			contentType: "application/json",
		},
		response: testResponse{
			400, "bad request\n",
		},
		strict: true,
	}, {
		name:    "invalid POST request; extra property",
		handler: validatorTestHandler{}.withDefaults(),
		request: testRequest{
			method:      "POST",
			path:        "/test?version=1",
			body:        `{"name": "foo", "expected": 9, "actual": 10, "ideal": 8}`,
			contentType: "application/json",
		},
		response: testResponse{
			400, "bad request\n",
		},
		strict: true,
	}, {
		name: "valid response; 404 error",
		handler: validatorTestHandler{
			contentType:   "application/json",
			errBody:       `{"code": "404", "message": "not found"}`,
			errStatusCode: 404,
		}.withDefaults(),
		request: testRequest{
			method: "GET",
			path:   "/test/42?version=1",
		},
		response: testResponse{
			404, `{"code": "404", "message": "not found"}`,
		},
		strict: true,
	}, {
		name: "invalid response; invalid error",
		handler: validatorTestHandler{
			errBody:       `"not found"`,
			errStatusCode: 404,
		}.withDefaults(),
		request: testRequest{
			method: "GET",
			path:   "/test/42?version=1",
		},
		response: testResponse{
			500, "server error\n",
		},
		strict: true,
	}, {
		name: "invalid POST response; not strict",
		handler: validatorTestHandler{
			postBody: `{"id": "42", "contents": {"name": "foo", "expected": 9, "actual": 10}, "extra": true}`,
		}.withDefaults(),
		request: testRequest{
			method:      "POST",
			path:        "/test?version=1",
			body:        `{"name": "foo", "expected": 9, "actual": 10}`,
			contentType: "application/json",
		},
		response: testResponse{
			statusCode: 201,
			body:       `{"id": "42", "contents": {"name": "foo", "expected": 9, "actual": 10}, "extra": true}`,
		},
		strict: false,
	}, {
		name: "POST response status code not in spec (return 200, spec only has 201)",
		handler: validatorTestHandler{
			postBody:      `{"id": "42", "contents": {"name": "foo", "expected": 9, "actual": 10}, "extra": true}`,
			errStatusCode: 200,
			errBody:       `{"id": "42", "contents": {"name": "foo", "expected": 9, "actual": 10}, "extra": true}`,
		}.withDefaults(),
		options: []openapi3filter.ValidatorOption{openapi3filter.ValidationOptions(openapi3filter.Options{
			IncludeResponseStatus: true,
		})},
		request: testRequest{
			method:      "POST",
			path:        "/test?version=1",
			body:        `{"name": "foo", "expected": 9, "actual": 10}`,
			contentType: "application/json",
		},
		response: testResponse{
			statusCode: 200,
			body:       `{"id": "42", "contents": {"name": "foo", "expected": 9, "actual": 10}, "extra": true}`,
		},
		strict: false,
	}}
	for i, test := range tests {
		t.Logf("test#%d: %s", i, test.name)
		t.Run(test.name, func(t *testing.T) {
			// Set up a test HTTP server
			var h http.Handler
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				h.ServeHTTP(w, r)
			}))
			defer s.Close()

			// Update the OpenAPI servers section with the test server URL This is
			// needed by the router which matches request routes for OpenAPI
			// validation.
			doc.Servers = []*openapi3.Server{{URL: s.URL}}
			err = doc.Validate(ctx)
			require.NoError(t, err, "failed to validate with test server")

			// Create the router and validator
			router, err := gorillamux.NewRouter(doc)
			require.NoError(t, err, "failed to create router")

			// Now wrap the test handler with the validator middlware
			v := openapi3filter.NewValidator(router, append(test.options, openapi3filter.Strict(test.strict))...)
			h = v.Middleware(&test.handler)

			// Test: make a client request
			var requestBody io.Reader
			if test.request.body != "" {
				requestBody = bytes.NewBufferString(test.request.body)
			}
			req, err := http.NewRequest(test.request.method, s.URL+test.request.path, requestBody)
			require.NoError(t, err, "failed to create request")

			if test.request.contentType != "" {
				req.Header.Set("Content-Type", test.request.contentType)
			}
			resp, err := s.Client().Do(req)
			require.NoError(t, err, "request failed")
			defer resp.Body.Close()
			require.Equalf(t, test.response.statusCode, resp.StatusCode,
				"response code expect %d got %d", test.response.statusCode, resp.StatusCode)

			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err, "failed to read response body")
			require.Equalf(t, test.response.body, string(body),
				"response body expect %q got %q", test.response.body, string(body))
		})
	}
}

func ExampleValidator() {
	// OpenAPI specification for a simple service that squares integers, with
	// some limitations.
	doc, err := openapi3.NewLoader().LoadFromData([]byte(`
openapi: 3.0.0
info:
  title: 'Validator - square example'
  version: '0.0.0'
paths:
  /square/{x}:
    get:
      description: square an integer
      parameters:
        - name: x
          in: path
          schema:
            type: integer
          required: true
      responses:
        '200':
          description: squared integer response
          content:
            "application/json":
              schema:
                type: object
                properties:
                  result:
                    type: integer
                    minimum: 0
                    maximum: 1000000
                required: [result]
                additionalProperties: false`[1:]))
	if err != nil {
		panic(err)
	}

	// Square service handler sanity checks inputs, but just crashes on invalid
	// requests.
	squareHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		xParam := path.Base(r.URL.Path)
		x, err := strconv.ParseInt(xParam, 10, 64)
		if err != nil {
			panic(err)
		}
		w.Header().Set("Content-Type", "application/json")
		result := map[string]interface{}{"result": x * x}
		if x == 42 {
			// An easter egg. Unfortunately, the spec does not allow additional properties...
			result["comment"] = "the answer to the ulitimate question of life, the universe, and everything"
		}
		if err = json.NewEncoder(w).Encode(&result); err != nil {
			panic(err)
		}
	})

	// Start an http server.
	var mainHandler http.Handler
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Why are we wrapping the main server handler with a closure here?
		// Validation matches request Host: to server URLs in the spec. With an
		// httptest.Server, the URL is dynamic and we have to create it first!
		// In a real configured service, this is less likely to be needed.
		mainHandler.ServeHTTP(w, r)
	}))
	defer srv.Close()

	// Patch the OpenAPI spec to match the httptest.Server.URL. Only needed
	// because the server URL is dynamic here.
	doc.Servers = []*openapi3.Server{{URL: srv.URL}}
	if err := doc.Validate(context.Background()); err != nil { // Assert our OpenAPI is valid!
		panic(err)
	}
	// This router is used by the validator to match requests with the OpenAPI
	// spec. It does not place restrictions on how the wrapped handler routes
	// requests; use of gorilla/mux is just a validator implementation detail.
	router, err := gorillamux.NewRouter(doc)
	if err != nil {
		panic(err)
	}
	// Strict validation will respond HTTP 500 if the service tries to emit a
	// response that does not conform to the OpenAPI spec. Very useful for
	// testing a service against its spec in development and CI. In production,
	// availability may be more important than strictness.
	v := openapi3filter.NewValidator(router, openapi3filter.Strict(true),
		openapi3filter.OnErr(func(w http.ResponseWriter, status int, code openapi3filter.ErrCode, err error) {
			// Customize validation error responses to use JSON
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  status,
				"message": http.StatusText(status),
			})
		}))
	// Now we can finally set the main server handler.
	mainHandler = v.Middleware(squareHandler)

	printResp := func(resp *http.Response, err error) {
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		fmt.Println(resp.StatusCode, strings.TrimSpace(string(contents)))
	}
	// Valid requests to our sum service
	printResp(srv.Client().Get(srv.URL + "/square/2"))
	printResp(srv.Client().Get(srv.URL + "/square/789"))
	// 404 Not found requests - method or path not found
	printResp(srv.Client().Post(srv.URL+"/square/2", "application/json", bytes.NewBufferString(`{"result": 5}`)))
	printResp(srv.Client().Get(srv.URL + "/sum/2"))
	printResp(srv.Client().Get(srv.URL + "/square/circle/4")) // Handler would process this; validation rejects it
	printResp(srv.Client().Get(srv.URL + "/square"))
	// 400 Bad requests - note they never reach the wrapped square handler (which would panic)
	printResp(srv.Client().Get(srv.URL + "/square/five"))
	// 500 Invalid responses
	printResp(srv.Client().Get(srv.URL + "/square/42"))    // Our "easter egg" added a property which is not allowed
	printResp(srv.Client().Get(srv.URL + "/square/65536")) // Answer overflows the maximum allowed value (1000000)
	// Output:
	// 200 {"result":4}
	// 200 {"result":622521}
	// 404 {"message":"Not Found","status":404}
	// 404 {"message":"Not Found","status":404}
	// 404 {"message":"Not Found","status":404}
	// 404 {"message":"Not Found","status":404}
	// 400 {"message":"Bad Request","status":400}
	// 500 {"message":"Internal Server Error","status":500}
	// 500 {"message":"Internal Server Error","status":500}
}
