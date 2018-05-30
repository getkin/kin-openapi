package openapi3filter_test

import (
	"bytes"
	"encoding/json"
	"github.com/ronniedada/kin-openapi/openapi3"
	"github.com/ronniedada/kin-openapi/openapi3filter"
	"github.com/ronniedada/kin-test/jsontest"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
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

func TestFilter(t *testing.T) {
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
					},
				},
			},
		},
	}
	router := openapi3filter.NewRouter().WithSwagger(swagger)

	// Declare helper method
	var req ExampleRequest
	var resp ExampleResponse
	expect := func(req ExampleRequest, resp ExampleResponse) *jsontest.ValueTester {
		t.Logf("Request: %s %s", req.Method, req.URL)
		httpReq, _ := http.NewRequest(req.Method, req.URL, marshalReader(req.Body))
		httpReq.Header.Set("Content-Type", req.ContentType)

		// Find route
		route, pathParams, err := router.FindRoute(httpReq.Method, httpReq.URL)
		if err != nil {
			return jsontest.ExpectWithErr(t, nil, err)
		}

		// Validate request
		requestValidationInput := &openapi3filter.RequestValidationInput{
			Request:    httpReq,
			PathParams: pathParams,
			Route:      route,
		}
		err = openapi3filter.ValidateRequest(nil, requestValidationInput)
		if err != nil {
			return jsontest.ExpectWithErr(t, nil, err)
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
			if err != nil {
				panic(err)
			}
			responseValidationInput.SetBodyBytes(data)
		}
		err = openapi3filter.ValidateResponse(nil, responseValidationInput)
		return jsontest.ExpectWithErr(t, nil, err)
	}

	// Test paths
	req = ExampleRequest{
		Method: "POST",
		URL:    "http://example.com/api/prefix/v/suffix",
	}
	resp = ExampleResponse{
		Status: 200,
	}
	expect(req, resp).ErrOfType(nil)

	// Test query parameter openapi3filter
	req = ExampleRequest{
		Method: "POST",
		URL:    "http://example.com/api/prefix/EXCEEDS_MAX_LENGTH/suffix",
	}
	expect(req, resp).ErrOfType(&openapi3filter.RequestError{})

	// Test query parameter openapi3filter
	req = ExampleRequest{
		Method: "POST",
		URL:    "http://example.com/api/prefix/v/suffix?queryArg=a",
	}
	expect(req, resp).ErrOfType(nil)

	req = ExampleRequest{
		Method: "POST",
		URL:    "http://example.com/api/prefix/v/suffix?queryArg=EXCEEDS_MAX_LENGTH",
	}
	expect(req, resp).ErrOfType(&openapi3filter.RequestError{})

	// Test query parameter openapi3filter
	req = ExampleRequest{
		Method: "POST",
		URL:    "http://example.com/api/prefix/v/suffix?queryArg=a",
	}
	expect(req, resp).ErrOfType(nil)

	req = ExampleRequest{
		Method: "POST",
		URL:    "http://example.com/api/prefix/v/suffix?queryArg=EXCEEDS_MAX_LENGTH",
	}
	expect(req, resp).ErrOfType(&openapi3filter.RequestError{})

	req = ExampleRequest{
		Method: "POST",
		URL:    "http://example.com/api/prefix/v/suffix",
	}
	resp = ExampleResponse{
		Status: 200,
	}
	expect(req, resp).ErrOfType(&openapi3filter.ResponseError{})
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
