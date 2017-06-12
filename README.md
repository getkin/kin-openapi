# kinapi
The library provides packages for dealing with OpenAPI specifications.

The main features are:
  * Targets OpenAPI version 3.
  * Transforms OpenAPI files from v2 to v3
  * Transforms OpenAPI files from v3 to v2

## kinapi vs go-openapi
The [go-openapi](https://github.com/go-openapi) project provides a stable and well-tested implementation of OpenAPI version 2.

The main difference is that this library targets OpenAPI version 3. You might also find that the API is slightly less cluttered because of the approach we chose to serializing/deserializing JSON.

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

# Dependencies
  * Go 1.5
  * Tests require [github.com/jban332/kincore](github.com/jban332/kincore)
