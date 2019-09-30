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
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/stretchr/testify/require"
)

type ExampleRequest struct {
	Method      string
	URL         string
	ContentType string
	Body        interface{}
}

type ExampleResponse struct {
	Status      int
	ContentType string
	Body        interface{}
}

type ExampleSecurityScheme struct {
	Name   string
	Scheme *openapi3.SecurityScheme
}

func TestFilter(t *testing.T) {
	// Declare a schema for an object with name and id properties
	complexArgSchema := openapi3.NewObjectSchema().
		WithProperty("name", openapi3.NewStringSchema()).
		WithProperty("id", openapi3.NewStringSchema().WithMaxLength(2))
	complexArgSchema.Required = []string{"name", "id"}

	// Declare router
	swagger := &openapi3.Swagger{
		Servers: openapi3.Servers{
			{
				URL: "http://example.com/api/",
			},
		},
		Paths: openapi3.Paths{
			"/prefix/{pathArg}/suffix": &openapi3.PathItem{
				Post: &openapi3.Operation{
					Parameters: openapi3.Parameters{
						{
							Value: &openapi3.Parameter{
								In:     "path",
								Name:   "pathArg",
								Schema: openapi3.NewStringSchema().WithMaxLength(2).NewRef(),
							},
						},
						{
							Value: &openapi3.Parameter{
								In:     "query",
								Name:   "queryArg",
								Schema: openapi3.NewStringSchema().WithMaxLength(2).NewRef(),
							},
						},
						{
							Value: &openapi3.Parameter{
								In:      "query",
								Name:    "contentArg",
								Content: openapi3.NewContentWithJSONSchema(complexArgSchema),
							},
						},
						{
							Value: &openapi3.Parameter{
								In:   "query",
								Name: "contentArg2",
								Content: openapi3.Content{
									"application/something_funny": openapi3.NewMediaType().WithSchema(complexArgSchema),
								},
							},
						},
					},
				},
			},
		},
	}

	router := openapi3filter.NewRouter().WithSwagger(swagger)
	expectWithDecoder := func(req ExampleRequest, resp ExampleResponse, decoder openapi3filter.ContentParameterDecoder) error {
		t.Logf("Request: %s %s", req.Method, req.URL)
		httpReq, _ := http.NewRequest(req.Method, req.URL, marshalReader(req.Body))
		httpReq.Header.Set("Content-Type", req.ContentType)

		// Find route
		route, pathParams, err := router.FindRoute(httpReq.Method, httpReq.URL)
		require.NoError(t, err)

		// Validate request
		requestValidationInput := &openapi3filter.RequestValidationInput{
			Request:      httpReq,
			PathParams:   pathParams,
			Route:        route,
			ParamDecoder: decoder,
		}
		if err := openapi3filter.ValidateRequest(context.TODO(), requestValidationInput); err != nil {
			return err
		}
		t.Logf("Response: %d", resp.Status)
		responseValidationInput := &openapi3filter.ResponseValidationInput{
			RequestValidationInput: requestValidationInput,
			Status:                 resp.Status,
			Header: http.Header{
				"Content-Type": []string{
					resp.ContentType,
				},
			},
		}
		if resp.Body != nil {
			data, err := json.Marshal(resp.Body)
			require.NoError(t, err)
			responseValidationInput.SetBodyBytes(data)
		}
		err = openapi3filter.ValidateResponse(context.TODO(), responseValidationInput)
		require.NoError(t, err)
		return err
	}
	expect := func(req ExampleRequest, resp ExampleResponse) error {
		return expectWithDecoder(req, resp, nil)
	}

	var err error
	var req ExampleRequest
	var resp ExampleResponse

	// Test paths
	req = ExampleRequest{
		Method: "POST",
		URL:    "http://example.com/api/prefix/v/suffix",
	}
	resp = ExampleResponse{
		Status: 200,
	}
	err = expect(req, resp)
	require.NoError(t, err)

	// Test query parameter openapi3filter
	req = ExampleRequest{
		Method: "POST",
		URL:    "http://example.com/api/prefix/EXCEEDS_MAX_LENGTH/suffix",
	}
	err = expect(req, resp)
	require.IsType(t, &openapi3filter.RequestError{}, err)

	// Test query parameter openapi3filter
	req = ExampleRequest{
		Method: "POST",
		URL:    "http://example.com/api/prefix/v/suffix?queryArg=a",
	}
	err = expect(req, resp)
	require.NoError(t, err)

	req = ExampleRequest{
		Method: "POST",
		URL:    "http://example.com/api/prefix/v/suffix?queryArg=EXCEEDS_MAX_LENGTH",
	}
	err = expect(req, resp)
	require.IsType(t, &openapi3filter.RequestError{}, err)

	// Test query parameter openapi3filter
	req = ExampleRequest{
		Method: "POST",
		URL:    "http://example.com/api/prefix/v/suffix?queryArg=a",
	}
	err = expect(req, resp)
	require.NoError(t, err)

	req = ExampleRequest{
		Method: "POST",
		URL:    "http://example.com/api/prefix/v/suffix?queryArg=EXCEEDS_MAX_LENGTH",
	}
	err = expect(req, resp)
	require.IsType(t, &openapi3filter.RequestError{}, err)

	req = ExampleRequest{
		Method: "POST",
		URL:    "http://example.com/api/prefix/v/suffix",
	}
	resp = ExampleResponse{
		Status: 200,
	}
	err = expect(req, resp)
	// require.IsType(t, &openapi3filter.ResponseError{}, err)
	require.NoError(t, err)

	// Check that content validation works. This should pass, as ID is short
	// enough.
	req = ExampleRequest{
		Method: "POST",
		URL:    "http://example.com/api/prefix/v/suffix?contentArg={\"name\":\"bob\", \"id\":\"a\"}",
	}
	err = expect(req, resp)
	require.NoError(t, err)

	// Now it should fail due the ID being too long
	req = ExampleRequest{
		Method: "POST",
		URL:    "http://example.com/api/prefix/v/suffix?contentArg={\"name\":\"bob\", \"id\":\"EXCEEDS_MAX_LENGTH\"}",
	}
	err = expect(req, resp)
	require.IsType(t, &openapi3filter.RequestError{}, err)

	// Now, repeat the above two test cases using a custom parameter decoder.
	customDecoder := func(param *openapi3.Parameter, values []string) (interface{}, *openapi3.Schema, error) {
		var value interface{}
		err := json.Unmarshal([]byte(values[0]), &value)
		schema := param.Content.Get("application/something_funny").Schema.Value
		return value, schema, err
	}

	req = ExampleRequest{
		Method: "POST",
		URL:    "http://example.com/api/prefix/v/suffix?contentArg2={\"name\":\"bob\", \"id\":\"a\"}",
	}
	err = expectWithDecoder(req, resp, customDecoder)
	require.NoError(t, err)

	// Now it should fail due the ID being too long
	req = ExampleRequest{
		Method: "POST",
		URL:    "http://example.com/api/prefix/v/suffix?contentArg2={\"name\":\"bob\", \"id\":\"EXCEEDS_MAX_LENGTH\"}",
	}
	err = expectWithDecoder(req, resp, customDecoder)
	require.IsType(t, &openapi3filter.RequestError{}, err)
}

func marshalReader(value interface{}) io.ReadCloser {
	if value == nil {
		return nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return ioutil.NopCloser(bytes.NewReader(data))
}

func TestValidateRequestBody(t *testing.T) {
	requiredReqBody := openapi3.NewRequestBody().
		WithContent(openapi3.NewContentWithJSONSchema(openapi3.NewStringSchema())).
		WithRequired(true)

	plainTextContent := openapi3.NewContent()
	plainTextContent["text/plain"] = openapi3.NewMediaType().WithSchema(openapi3.NewStringSchema())

	testCases := []struct {
		name    string
		body    *openapi3.RequestBody
		mime    string
		data    io.Reader
		wantErr error
	}{
		{
			name: "non required empty",
			body: openapi3.NewRequestBody().
				WithContent(openapi3.NewContentWithJSONSchema(openapi3.NewStringSchema())),
		},
		{
			name: "non required not empty",
			body: openapi3.NewRequestBody().
				WithContent(openapi3.NewContentWithJSONSchema(openapi3.NewStringSchema())),
			mime: "application/json",
			data: toJSON("foo"),
		},
		{
			name:    "required empty",
			body:    requiredReqBody,
			wantErr: &openapi3filter.RequestError{RequestBody: requiredReqBody, Err: openapi3filter.ErrInvalidRequired},
		},
		{
			name: "required not empty",
			body: requiredReqBody,
			mime: "application/json",
			data: toJSON("foo"),
		},
		{
			name: "not JSON data",
			body: openapi3.NewRequestBody().WithContent(plainTextContent).WithRequired(true),
			mime: "text/plain",
			data: strings.NewReader("foo"),
		},
		{
			name: "not declared content",
			body: openapi3.NewRequestBody().WithRequired(true),
			mime: "application/json",
			data: toJSON("foo"),
		},
		{
			name: "not declared schema",
			body: openapi3.NewRequestBody().WithJSONSchemaRef(nil),
			mime: "application/json",
			data: toJSON("foo"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/test", tc.data)
			if tc.mime != "" {
				req.Header.Set(http.CanonicalHeaderKey("Content-Type"), tc.mime)
			}
			inp := &openapi3filter.RequestValidationInput{Request: req}
			err := openapi3filter.ValidateRequestBody(context.Background(), inp, tc.body)

			if tc.wantErr == nil {
				require.NoError(t, err)
				return
			}

			require.True(t, matchReqBodyError(tc.wantErr, err), "got error:\n%s\nwant error\n%s", err, tc.wantErr)
		})
	}
}

func matchReqBodyError(want, got error) bool {
	if want == got {
		return true
	}
	wErr, ok := want.(*openapi3filter.RequestError)
	if !ok {
		return false
	}
	gErr, ok := got.(*openapi3filter.RequestError)
	if !ok {
		return false
	}
	if !reflect.DeepEqual(wErr.RequestBody, gErr.RequestBody) {
		return false
	}
	if wErr.Err != nil {
		return matchReqBodyError(wErr.Err, gErr.Err)
	}
	return false
}

func toJSON(v interface{}) io.Reader {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return bytes.NewReader(data)
}

// TestOperationOrSwaggerSecurity asserts that the swagger's SecurityRequirements are used if no SecurityRequirements are provided for an operation.
func TestOperationOrSwaggerSecurity(t *testing.T) {
	// Create the security schemes
	securitySchemes := []ExampleSecurityScheme{
		{
			Name: "apikey",
			Scheme: &openapi3.SecurityScheme{
				Type: "apiKey",
				Name: "apikey",
				In:   "cookie",
			},
		},
		{
			Name: "http-basic",
			Scheme: &openapi3.SecurityScheme{
				Type:   "http",
				Scheme: "basic",
			},
		},
	}

	// Create the test cases
	tc := []struct {
		name            string
		schemes         *[]ExampleSecurityScheme
		expectedSchemes *[]ExampleSecurityScheme
	}{
		{
			name:    "/inherited-security",
			schemes: nil,
			expectedSchemes: &[]ExampleSecurityScheme{
				securitySchemes[1],
			},
		},
		{
			name:            "/overwrite-without-security",
			schemes:         &[]ExampleSecurityScheme{},
			expectedSchemes: &[]ExampleSecurityScheme{},
		},
		{
			name: "/overwrite-with-security",
			schemes: &[]ExampleSecurityScheme{
				securitySchemes[0],
			},
			expectedSchemes: &[]ExampleSecurityScheme{
				securitySchemes[0],
			},
		},
	}

	// Create the swagger
	swagger := &openapi3.Swagger{
		Paths: map[string]*openapi3.PathItem{},
		Security: openapi3.SecurityRequirements{
			{
				securitySchemes[1].Name: {},
			},
		},
		Components: openapi3.Components{
			SecuritySchemes: map[string]*openapi3.SecuritySchemeRef{},
		},
	}

	// Add the security schemes to the components
	for _, scheme := range securitySchemes {
		swagger.Components.SecuritySchemes[scheme.Name] = &openapi3.SecuritySchemeRef{
			Value: scheme.Scheme,
		}
	}

	// Add the paths from the test cases to the swagger's paths
	for _, tc := range tc {
		var securityRequirements *openapi3.SecurityRequirements = nil
		if tc.schemes != nil {
			tempS := make(openapi3.SecurityRequirements, 0)
			for _, scheme := range *tc.schemes {
				tempS = append(
					tempS,
					openapi3.SecurityRequirement{
						scheme.Name: {},
					},
				)
			}
			securityRequirements = &tempS
		}
		swagger.Paths[tc.name] = &openapi3.PathItem{
			Get: &openapi3.Operation{
				Security: securityRequirements,
			},
		}
	}

	// Declare the router
	router := openapi3filter.NewRouter().WithSwagger(swagger)

	// Test each case
	for _, path := range tc {
		// Make a map of the schemes and whether they're
		var schemesValidated *map[*openapi3.SecurityScheme]bool = nil
		if path.expectedSchemes != nil {
			temp := make(map[*openapi3.SecurityScheme]bool)
			schemesValidated = &temp
			for _, scheme := range *path.expectedSchemes {
				(*schemesValidated)[scheme.Scheme] = false
			}
		}

		// Create the request
		emptyBody := bytes.NewReader(make([]byte, 0))
		pathUrl, err := url.Parse(path.name)
		require.NoError(t, err)
		route, _, err := router.FindRoute(http.MethodGet, pathUrl)
		require.NoError(t, err)
		req := openapi3filter.RequestValidationInput{
			Request: httptest.NewRequest(http.MethodGet, path.name, emptyBody),
			Route:   route,
			Options: &openapi3filter.Options{
				AuthenticationFunc: func(c context.Context, input *openapi3filter.AuthenticationInput) error {
					if schemesValidated != nil {
						if validated, ok := (*schemesValidated)[input.SecurityScheme]; ok {
							if validated {
								t.Fatalf("The path \"%s\" had the schemes %v named \"%s\" validated more than once",
									path.name, input.SecurityScheme, input.SecuritySchemeName)
							}
							(*schemesValidated)[input.SecurityScheme] = true
							return nil
						}
					}

					t.Fatalf("The path \"%s\" had the schemes %v named \"%s\"",
						path.name, input.SecurityScheme, input.SecuritySchemeName)

					return nil
				},
			},
		}

		// Validate the request
		err = openapi3filter.ValidateRequest(nil, &req)
		require.NoError(t, err)

		for securityRequirement, validated := range *schemesValidated {
			if !validated {
				t.Fatalf("The security requirement %v was exepected to be validated but wasn't",
					securityRequirement)
			}
		}
	}
}

// TestAlternateRequirementMet asserts that ValidateSecurityRequirements succeeds if any SecurityRequirement is met and otherwise doesn't.
func TestAnySecurityRequirementMet(t *testing.T) {
	// Create of a map of scheme names and whether they are valid
	schemes := map[string]bool{
		"a": true,
		"b": true,
		"c": false,
		"d": false,
	}

	// Create the test cases
	tc := []struct {
		name    string
		schemes []string
		error   bool
	}{
		{
			name:    "/ok1",
			schemes: []string{"a", "b"},
			error:   false,
		},
		{
			name:    "/ok2",
			schemes: []string{"a", "c"},
			error:   false,
		},
		{
			name:    "/error",
			schemes: []string{"c", "d"},
			error:   true,
		},
	}

	// Create the swagger
	swagger := openapi3.Swagger{
		Paths: map[string]*openapi3.PathItem{},
		Components: openapi3.Components{
			SecuritySchemes: map[string]*openapi3.SecuritySchemeRef{},
		},
	}

	// Add the security schemes to the swagger's components
	for schemeName := range schemes {
		swagger.Components.SecuritySchemes[schemeName] = &openapi3.SecuritySchemeRef{
			Value: &openapi3.SecurityScheme{
				Type:   "http",
				Scheme: "basic",
			},
		}
	}

	// Add the paths to the swagger
	for _, tc := range tc {
		// Create the security requirements from the test cases's schemes
		securityRequirements := make(openapi3.SecurityRequirements, len(tc.schemes))
		for i, scheme := range tc.schemes {
			securityRequirements[i] = openapi3.SecurityRequirement{
				scheme: {},
			}
		}

		// Create the path with the security requirements
		swagger.Paths[tc.name] = &openapi3.PathItem{
			Get: &openapi3.Operation{
				Security: &securityRequirements,
			},
		}
	}

	// Create the router
	router := openapi3filter.NewRouter().WithSwagger(&swagger)

	// Create the authentication function
	authFunc := makeAuthFunc(schemes)

	for _, tc := range tc {
		// Create the request input for the path
		tcUrl, err := url.Parse(tc.name)
		require.NoError(t, err)
		route, _, err := router.FindRoute(http.MethodGet, tcUrl)
		require.NoError(t, err)
		req := openapi3filter.RequestValidationInput{
			Route: route,
			Options: &openapi3filter.Options{
				AuthenticationFunc: authFunc,
			},
		}

		// Validate the security requirements
		err = openapi3filter.ValidateSecurityRequirements(nil, &req, *route.Operation.Security)

		// If there should have been an error
		if tc.error {
			require.Errorf(t, err, "an error is expected for path \"%s\"", tc.name)
		} else {
			require.NoErrorf(t, err, "an error wasn't expected for path \"%s\"", tc.name)
		}
	}
}

// TestAllSchemesMet asserts that ValidateSecurityRequirement succeeds if all the SecuritySchemes of a SecurityRequirement are met and otherwise doesn't.
func TestAllSchemesMet(t *testing.T) {
	// Create of a map of scheme names and whether they are met
	schemes := map[string]bool{
		"a": true,
		"b": true,
		"c": false,
	}

	// Create the test cases
	tc := []struct {
		name  string
		error bool
	}{
		{
			name:  "/ok",
			error: false,
		},
		{
			name:  "/error",
			error: true,
		},
	}

	// Create the swagger
	swagger := openapi3.Swagger{
		Paths: map[string]*openapi3.PathItem{},
		Components: openapi3.Components{
			SecuritySchemes: map[string]*openapi3.SecuritySchemeRef{},
		},
	}

	// Add the security schemes to the swagger's components
	for schemeName := range schemes {
		swagger.Components.SecuritySchemes[schemeName] = &openapi3.SecuritySchemeRef{
			Value: &openapi3.SecurityScheme{
				Type:   "http",
				Scheme: "basic",
			},
		}
	}

	// Add the paths to the swagger
	for _, tc := range tc {
		// Create the security requirement for the path
		securityRequirement := openapi3.SecurityRequirement{}
		for scheme, valid := range schemes {
			// If the scheme is valid or the test case is meant to return an error
			if valid || tc.error {
				// Add the scheme to the security requirement
				securityRequirement[scheme] = []string{}
			}
		}

		swagger.Paths[tc.name] = &openapi3.PathItem{
			Get: &openapi3.Operation{
				Security: &openapi3.SecurityRequirements{
					securityRequirement,
				},
			},
		}
	}

	// Create the router from the swagger
	router := openapi3filter.NewRouter().WithSwagger(&swagger)

	// Create the authentication function
	authFunc := makeAuthFunc(schemes)

	for _, tc := range tc {
		// Create the request input for the path
		tcUrl, err := url.Parse(tc.name)
		require.NoError(t, err)
		route, _, err := router.FindRoute(http.MethodGet, tcUrl)
		require.NoError(t, err)
		req := openapi3filter.RequestValidationInput{
			Route: route,
			Options: &openapi3filter.Options{
				AuthenticationFunc: authFunc,
			},
		}

		// Validate the security requirements
		err = openapi3filter.ValidateSecurityRequirements(nil, &req, *route.Operation.Security)

		// If there should have been an error
		if tc.error {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

// makeAuthFunc creates an authentication function that accepts the given valid schemes.
// If an invalid or unknown scheme is encountered, an error is returned by the returned function.
// Otherwise the return value of the returned function is nil.
func makeAuthFunc(schemes map[string]bool) func(c context.Context, input *openapi3filter.AuthenticationInput) error {
	return func(c context.Context, input *openapi3filter.AuthenticationInput) error {
		// If the scheme is valid and present in the schemes
		valid, present := schemes[input.SecuritySchemeName]
		if valid && present {
			return nil
		}

		// If the scheme is present in che schemes
		if present {
			// Return an unmet scheme error
			return fmt.Errorf("security scheme for \"%s\" wasn't met", input.SecuritySchemeName)
		} else {
			// Return an unknown scheme error
			return fmt.Errorf("security scheme for \"%s\" is unknown", input.SecuritySchemeName)
		}
	}
}
