// Package openapi3 parses and writes OpenAPI 3 specification documents.
//
// Supports both OpenAPI 3.0 and OpenAPI 3.1:
//   - OpenAPI 3.0.x: https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.3.md
//   - OpenAPI 3.1.x: https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.1.0.md
//
// OpenAPI 3.1 Features:
//   - Type arrays with null support (e.g., ["string", "null"])
//   - JSON Schema 2020-12 keywords (const, examples, prefixItems, etc.)
//   - Webhooks for defining callback operations
//   - JSON Schema dialect specification
//   - SPDX license identifiers
//
// The implementation maintains 100% backward compatibility with OpenAPI 3.0.
//
// For OpenAPI 3.1 validation, use the JSON Schema 2020-12 validator option:
//
//	schema.VisitJSON(value, openapi3.EnableJSONSchema2020())
//
// Version detection is available via helper methods:
//
//	if doc.IsOpenAPI3_1() {
//	    // Handle OpenAPI 3.1 specific features
//	}
package openapi3
