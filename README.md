[![CI](https://github.com/getkin/kin-openapi/workflows/go/badge.svg)](https://github.com/getkin/kin-openapi/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/getkin/kin-openapi)](https://goreportcard.com/report/github.com/getkin/kin-openapi)
[![GoDoc](https://godoc.org/github.com/getkin/kin-openapi?status.svg)](https://godoc.org/github.com/getkin/kin-openapi)
[![Join Gitter Chat Channel -](https://badges.gitter.im/getkin/kin.svg)](https://gitter.im/getkin/kin?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

# Introduction
A [Go](https://golang.org) project for handling [OpenAPI](https://www.openapis.org/) files. We target the latest OpenAPI version (currently 3), but the project contains support for older OpenAPI versions too.

Licensed under the [MIT License](LICENSE).

## Contributors and users
The project has received pull requests from many people. Thanks to everyone!

Here's some projects that depend on _kin-openapi_:
  * [github.com/getkin/kin](https://github.com/getkin/kin) - "A configurable backend"
  * [github.com/danielgtaylor/apisprout](https://github.com/danielgtaylor/apisprout) - "Lightweight, blazing fast, cross-platform OpenAPI 3 mock server with validation"
  * [github.com/deepmap/oapi-codegen](https://github.com/deepmap/oapi-codegen) - Generate Go server boilerplate from an OpenAPI 3 spec
  * [github.com/dunglas/vulcain](https://github.com/dunglas/vulcain) - "Use HTTP/2 Server Push to create fast and idiomatic client-driven REST APIs"
  * [github.com/danielgtaylor/restish](https://github.com/danielgtaylor/restish) - "...a CLI for interacting with REST-ish HTTP APIs with some nice features built-in"
  * [github.com/goadesign/goa](https://github.com/goadesign/goa) - "Goa is a framework for building micro-services and APIs in Go using a unique design-first approach."
  * (Feel free to add your project by [creating an issue](https://github.com/getkin/kin-openapi/issues/new) or a pull request)

## Alternative projects
  * [go-openapi](https://github.com/go-openapi)
    * Supports OpenAPI version 2.
  * See [this list](https://github.com/OAI/OpenAPI-Specification/blob/master/IMPLEMENTATIONS.md).

# Structure
  * _openapi2_ ([godoc](https://godoc.org/github.com/getkin/kin-openapi/openapi2))
    * Support for OpenAPI 2 files, including serialization, deserialization, and validation.
  * _openapi2conv_ ([godoc](https://godoc.org/github.com/getkin/kin-openapi/openapi2conv))
    * Converts OpenAPI 2 files into OpenAPI 3 files.
  * _openapi3_ ([godoc](https://godoc.org/github.com/getkin/kin-openapi/openapi3))
    * Support for OpenAPI 3 files, including serialization, deserialization, and validation.
  * _openapi3filter_ ([godoc](https://godoc.org/github.com/getkin/kin-openapi/openapi3filter))
    * Validates HTTP requests and responses
  * _openapi3gen_ ([godoc](https://godoc.org/github.com/getkin/kin-openapi/openapi3gen))
    * Generates `*openapi3.Schema` values for Go types.
  * _pathpattern_ ([godoc](https://godoc.org/github.com/getkin/kin-openapi/pathpattern))
    * Matches strings with OpenAPI path patterns ("/path/{parameter}")

# Some recipes
## Loading OpenAPI document
Use `SwaggerLoader`, which resolves all JSON references:
```go
swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromFile("swagger.json")
```

## Getting OpenAPI operation that matches request
```go
func GetOperation(httpRequest *http.Request) (*openapi3.Operation, error) {
  // Load Swagger file
  router := openapi3filter.NewRouter().WithSwaggerFromFile("swagger.json")

  // Find route
  route, _, err := router.FindRoute("GET", req.URL)
  if err != nil {
    return nil, err
  }

  // Get OpenAPI 3 operation
  return route.Operation
}
```

## Validating HTTP requests/responses
```go
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3filter"
)

func main() {
	router := openapi3filter.NewRouter().WithSwaggerFromFile("swagger.json")
	ctx := context.Background()
	httpReq, _ := http.NewRequest(http.MethodGet, "/items", nil)

	// Find route
	route, pathParams, _ := router.FindRoute(httpReq.Method, httpReq.URL)

	// Validate request
	requestValidationInput := &openapi3filter.RequestValidationInput{
		Request:    httpReq,
		PathParams: pathParams,
		Route:      route,
	}
	if err := openapi3filter.ValidateRequest(ctx, requestValidationInput); err != nil {
		panic(err)
	}

	var (
		respStatus      = 200
		respContentType = "application/json"
		respBody        = bytes.NewBufferString(`{}`)
	)

	log.Println("Response:", respStatus)
	responseValidationInput := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: requestValidationInput,
		Status:                 respStatus,
		Header: http.Header{
			"Content-Type": []string{respContentType},
		},
	}
	if respBody != nil {
		data, _ := json.Marshal(respBody)
		responseValidationInput.SetBodyBytes(data)
	}

	// Validate response.
	if err := openapi3filter.ValidateResponse(ctx, responseValidationInput); err != nil {
		panic(err)
	}
}
```

## Custom content type for body of HTTP request/response

By default, the library parses a body of HTTP request and response
if it has one of the next content types: `"text/plain"` or `"application/json"`.
To support other content types you must register decoders for them:

```go
func main() {
	// ...

	// Register a body's decoder for content type "application/xml".
	openapi3filter.RegisterBodyDecoder("application/xml", xmlBodyDecoder)

	// Now you can validate HTTP request that contains a body with content type "application/xml".
	requestValidationInput := &openapi3filter.RequestValidationInput{
		Request:    httpReq,
		PathParams: pathParams,
		Route:      route,
	}
	if err := openapi3filter.ValidateRequest(ctx, requestValidationInput); err != nil {
		panic(err)
	}

	// ...

	// And you can validate HTTP response that contains a body with content type "application/xml".
	if err := openapi3filter.ValidateResponse(ctx, responseValidationInput); err != nil {
		panic(err)
	}
}

func xmlBodyDecoder(body []byte) (interface{}, error) {
	// Decode body to a primitive, []inteface{}, or map[string]interface{}.
}
```

## Custom function for check uniqueness of JSON array

By defaut, the library check unique items by below predefined function

```go
func isSliceOfUniqueItems(xs []interface{}) bool {
	s := len(xs)
	m := make(map[string]struct{}, s)
	for _, x := range xs {
		// The input slice is coverted from a JSON string, there shall
		// have no error when covert it back.
		key, _ := json.Marshal(&x)
		m[string(key)] = struct{}{}
	}
	return s == len(m)
}
```

In the predefined function using `json.Marshal` to generate a string can
be used as a map key which is to support check the uniqueness of an array
when the array items are JSON objects or JSON arraies. You can register
you own function according to your input data to get better performance:

```go
func main() {
	// ...

	// Register a customized function used to check uniqueness of array.
	openapi3.RegisterArrayUniqueItemsChecker(arrayUniqueItemsChecker)

	// ... other validate codes
}

func arrayUniqueItemsChecker(items []interface{}) bool {
	// Check the uniqueness of the input slice(array in JSON)
}
```
