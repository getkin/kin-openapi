# kinapi
The library provides packages for dealing with OpenAPI specifications.

## Features
  * Targets OpenAPI version 3.
  * Transforms OpenAPI files from v2 to v3
  * Transforms OpenAPI files from v3 to v2

## Dependencies
  * Go 1.7. Works in pre-1.7 versions if you provide `context.Context` to the compiler.
  * Tests require [github.com/jban332/kincore](github.com/jban332/kincore)

## kinapi vs go-openapi
The [go-openapi](https://github.com/go-openapi) project provides a stable and well-tested implementation of OpenAPI version 2.

The differences are:
  * This library targets OpenAPI version 3.
  * _go-openapi_ uses embedded structs in JSON marshalling/unmarshalling, while we use `jsoninfo` package.

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

# Using JSON serialization in other projects
The package `jsoninfo` was written to deal with JSON references and extension properties.

It:
  * Marshals/unmarshal JSON references
  * Marshals/unmarshal JSON extension properties (`"x-someExtension"`)
  * Refuses to unmarshal unsupported properties.

Usage looks like:
```
type Example struct {
 // Allow extension properties ("x-someProperty")
 jsoninfo.ExtensionProps
 
 // Allow reference property ("$ref")
 jsoninfo.RefProps
 
 // Normal properties
 SomeField float64
}

func (example *Example) MarshalJSON() ([]byte, error) {
 return jsoninfo.MarshalStructFields(example)
}

func (example *Example) UnmarshalJSON(data []byte) error {
 return jsoninfo.UnmarshalStructFields(data, example)
}
