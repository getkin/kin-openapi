# kinapi
The library provides packages for dealing with OpenAPI specifications.

## Features
  * Reads and writes [OpenAPI version 3.0 documents](https://github.com/OAI/OpenAPI-Specification/blob/OpenAPI.next/README.md)
  * Reads and writes [OpenAPI version 2.0 documents](https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md). Uses OpenAPI 3.0 elements where they are backwards compatible with version 2.0.
  * Transforms documents between OpenAPI versions:
    * 2.0 -> 3.0
    * 3.0 -> 2.0
  * Validates:
    * JSON schemas (`openapi3.Schema`)
    * HTTP requests
    * HTTP responses
  * Generates JSON schemas for Go types.

## Dependencies
  * Go 1.7. Works in pre-1.7 versions if you provide _context.Context_ to the compiler.
  * Tests require [github.com/jban332/kincore](github.com/jban332/kincore)

## Other OpenAPI  implementations
### go-openapi
The [go-openapi](https://github.com/go-openapi) project provides a stable and well-tested implementation of OpenAPI version 2.

The differences are:
  * This library targets OpenAPI version 3.
  * _go-openapi_ uses embedded structs in JSON marshalling/unmarshalling, while we use `jsoninfo` package.

### Others
See [this list](https://github.com/OAI/OpenAPI-Specification/blob/OpenAPI.next/IMPLEMENTATIONS.md).

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
