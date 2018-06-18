[![Build Status](https://travis-ci.org/getkin/kin-openapi.svg?branch=master)](https://travis-ci.org/getkin/kin-openapi)
[![Go Report Card](https://goreportcard.com/badge/github.com/getkin/kin-openapi)](https://goreportcard.com/report/github.com/getkin/kin-openapi)
[![GoDoc](https://godoc.org/github.com/getkin/kin-openapi?status.svg)](https://godoc.org/github.com/getkin/kin-openapi)

# Overview
This library provides packages for dealing with OpenAPI specifications.

## Status
### Current
  * [X] Reads and writes [OpenAPI version 3.0 documents](https://github.com/OAI/OpenAPI-Specification/blob/OpenAPI.next/README.md)
  * [X] Reads and writes [OpenAPI version 2.0 documents](https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md) and converts 2.0 -> 3.0 and 3.0->2.0.
    * Does NOT support all features.
  * [X] Validates:
    * [X] That a Go value matches [OpenAPI 3.0 schema object](https://github.com/OAI/OpenAPI-Specification/blob/OpenAPI.next/versions/3.0.md#schemaObject)
    * [X] That HTTP request matches [OpenAPI operation object](https://github.com/OAI/OpenAPI-Specification/blob/OpenAPI.next/versions/3.0.md#operationObject)
    * [X] That HTTP response matches OpenAPI 3.0 operation object
  * [X] Generates JSON schemas for Go types.

### TODO
  * [ ] More tests

## Dependencies
  * Go 1.7 or greater.
  * No other dependencies!

## License
  * [MIT License](LICENSE)

## Alternatives
  * [go-openapi](https://github.com/go-openapi)
    * Provides a stable and well-tested implementation of OpenAPI version 2.
  * See [this list](https://github.com/OAI/OpenAPI-Specification/blob/OpenAPI.next/IMPLEMENTATIONS.md).

# Packages
  * `jsoninfo`
    * Provides information and functions for marshalling/unmarshalling JSON. The purpose is a clutter-free implementation of JSON references and OpenAPI extension properties.
  * `openapi2`
    * Parses/writes OpenAPI 2.
  * `openapi2conv`
    * Converts OpenAPI 2 specification into OpenAPI 3 specification.
  * `openapi3`
    * Parses/writes OpenAPI 3. Includes OpenAPI schema / JSON schema valdation.
  * `openapi3filter`
    * Validates that HTTP request and HTTP response match an OpenAPI specification file.
  * `openapi3gen`
    * Generates OpenAPI 3 schemas for Go types.
  * `pathpattern`
    * Support for OpenAPI style path patterns.


# Getting started
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
  route, _, err := router.FindRoute("GET", req.URL.String())
  if err!=nil {
    return nil, err
  }

  // Get OpenAPI 3 operation
  return route.Operation
}
```

## Validating HTTP requests/responses
```go
import (
  "net/http"

  "github.com/getkin/kin-openapi/openapi3"
  "github.com/getkin/kin-openapi/openapi3filter"
)

var router = openapi3filter.NewRouter().WithSwaggerFromFile("swagger.json")

func ValidateRequest(req *http.Request) {
  openapi3filter.ValidateRequest(nil, &openapi3filter.ValidateRequestInput {
    Request: req,
    Router:  router,
  })

  // Get response

  openapi3filter.ValidateResponse(nil, &openapi3filter.ValidateResponseInput {
    // ...
  })
}

```

## Having extension properties in your own structs
The package `jsoninfo` marshals/unmarshal JSON extension properties (`"x-someExtension"`)

Usage looks like:
```go
type Example struct {
  // Allow extension properties ("x-someProperty")
  openapi3.ExtensionProps

  // Normal properties
  SomeField float64
}

func (example *Example) MarshalJSON() ([]byte, error) {
  return jsoninfo.MarshalStrictStruct(example)
}

func (example *Example) UnmarshalJSON(data []byte) error {
  return jsoninfo.UnmarshalStrictStruct(data, example)
}
```
