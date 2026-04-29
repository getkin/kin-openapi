// Package openapi3conv canonicalizes an OpenAPI 3.0 document into the OpenAPI 3.1
// representation. Schema-level constructs serialize differently between the two
// versions but represent the same semantics; this package rewrites the 3.0 forms
// into their 3.1 equivalents in place. The transformation is mechanical and
// lossless — every 3.0 construct has a direct 3.1 form.
//
// Use this when a downstream consumer (diff tools, validators, code generators)
// needs a single canonical representation regardless of the source spec's
// declared version.
//
// The opposite direction (3.1 → 3.0) is lossy by nature and out of scope.
package openapi3conv
