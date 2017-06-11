# Packages
  * `jsoninfo`
    * Provides information and functions for marshalling/unmarshalling JSON. The purpose is a nice implementation of JSON references and OpenAPI extension properties.
  * `openapi2` 
    * Parses/writes OpenAPI 2. For a much more complete implementation, we recommend [https://github.com/go-openapi](https://github.com/go-openapi).
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