// Package openapi3conv canonicalizes an OpenAPI 3.x document into the latest
// 3.x representation in place. Schema-level constructs serialize differently
// between OpenAPI 3.0 and 3.1, but represent the same semantics; this package
// rewrites the 3.0 forms into their 3.1 equivalents and bumps the version
// string to the latest 3.x patch release the package knows about.
//
// The OAI commits to strict compatibility for 3.x going forward (see the
// 3.2.1 and 3.3.0 milestones), so a tool that handles the 3.1+ form
// correctly handles all later 3.x versions correctly too. The 3.0 → 3.1
// transition — the only break in the 3.x line — is the gap this package
// exists to bridge. 3.1 → 3.2 (and any future 3.x) is purely additive and
// requires no rewrites; the package handles those as a version-string bump.
//
// Use this when a downstream consumer (diff tools, validators, code
// generators) needs a single canonical representation regardless of the
// source spec's declared version.
//
// Scope:
//   - In scope: 3.x → latest 3.x.
//   - Out of scope: any → 3.0 (downgrade is lossy by nature).
//   - Out of scope: cross-major upgrades (3 → 4 if/when v4 ships). Those
//     belong in a dedicated package mirroring openapi2conv (which converts
//     Swagger 2.0 documents to OpenAPI 3.0).
//
// Documents must be Validate()'d before calling Upgrade — passing an
// invalid document is undefined behaviour.
package openapi3conv
