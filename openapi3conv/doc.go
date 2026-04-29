// Package openapi3conv canonicalizes an OpenAPI 3.x document into the
// representation of a chosen target version. Schema-level constructs serialize
// differently between OpenAPI 3.0 and 3.1, but represent the same semantics;
// this package rewrites the 3.0 forms into their 3.1 equivalents in place.
// The transformation is mechanical and lossless — every 3.0 construct has a
// direct 3.1 form. OpenAPI 3.2 is purely additive over 3.1, so 3.1 → 3.2 is
// a version-string change with no further rewrites.
//
// Use this when a downstream consumer (diff tools, validators, code
// generators) needs a single canonical representation regardless of the
// source spec's declared version.
//
// Scope:
//   - In scope: 3.0 ↔ 3.1 ↔ 3.2 ↔ any future 3.x where representational
//     differences are rewritable in place.
//   - Out of scope: 3.x → 3.0 (lossy by nature).
//   - Out of scope: cross-major upgrades (3 → 4 if/when v4 ships). Those
//     belong in a dedicated package mirroring openapi2conv (which converts
//     Swagger 2.0 documents to OpenAPI 3.0).
package openapi3conv
