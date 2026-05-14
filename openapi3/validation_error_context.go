package openapi3

import "fmt"

// SectionValidationError wraps an error originating inside one of the
// top-level OpenAPI document sections (info, paths, components,
// security, servers, tags, externalDocs, webhooks, jsonSchemaDialect).
// Section is the OpenAPI field name as it appears in the document
// root.
//
// Use errors.As(err, &sve) to extract the section context from a
// validation error chain without parsing the rendered message.
type SectionValidationError struct {
	Section string
	Cause   error
}

func (e *SectionValidationError) Error() string {
	return fmt.Sprintf("invalid %s: %v", e.Section, e.Cause)
}

func (e *SectionValidationError) Unwrap() error { return e.Cause }

// PathValidationError wraps an error originating inside a specific path.
// Path is the path template as it appears in the document (e.g.
// "/users/{id}").
//
// Use errors.As(err, &pve) to extract the path from a validation
// error chain without parsing the rendered message.
type PathValidationError struct {
	Path  string
	Cause error
}

func (e *PathValidationError) Error() string {
	return fmt.Sprintf("invalid path %s: %v", e.Path, e.Cause)
}

func (e *PathValidationError) Unwrap() error { return e.Cause }

// OperationValidationError wraps an error originating inside a specific
// HTTP-method operation under a path. Method is the uppercase method
// (GET, POST, etc.).
//
// Use errors.As(err, &ove) to extract the method from a validation
// error chain without parsing the rendered message.
type OperationValidationError struct {
	Method string
	Cause  error
}

func (e *OperationValidationError) Error() string {
	return fmt.Sprintf("invalid operation %s: %v", e.Method, e.Cause)
}

func (e *OperationValidationError) Unwrap() error { return e.Cause }
