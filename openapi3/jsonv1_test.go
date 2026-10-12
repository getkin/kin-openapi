//go:build !go1.27 || !goexperiment.jsonv2

package openapi3

// UsingJSONv2 is whether this package decodes JSON with encoding/json/v2.
const UsingJSONv2 = false
