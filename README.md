[![Build Status](https://travis-ci.com/getkin/kin-openapi.svg?branch=master)](https://travis-ci.com/getkin/kin-openapi)
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
  * (Feel free to add your project by [creating an issue](https://github.com/getkin/kin-openapi/issues/new) or a pull request)

## Alternative projects
  * [go-openapi](https://github.com/go-openapi)
    * Supports OpenAPI version 2.
  * See [this list](https://github.com/OAI/OpenAPI-Specification/blob/OpenAPI.next/IMPLEMENTATIONS.md).

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