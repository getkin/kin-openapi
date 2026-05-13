package openapi3

import "fmt"

// SectionContextError wraps an error originating inside one of the
// top-level OpenAPI document sections (info, paths, components,
// security, servers, tags, externalDocs, webhooks, jsonSchemaDialect).
// Section is the OpenAPI field name as it appears in the document
// root.
//
// Use errors.As(err, &sce) to extract the section context from a
// validation error chain without parsing the rendered message.
type SectionContextError struct {
	Section string
	Cause   error
}

func (e *SectionContextError) Error() string {
	return fmt.Sprintf("invalid %s: %v", e.Section, e.Cause)
}

func (e *SectionContextError) Unwrap() error { return e.Cause }

// PathContextError wraps an error originating inside a specific path.
// Path is the path template as it appears in the document (e.g.
// "/users/{id}").
//
// Use errors.As(err, &pce) to extract the path from a validation
// error chain without parsing the rendered message.
type PathContextError struct {
	Path  string
	Cause error
}

func (e *PathContextError) Error() string {
	return fmt.Sprintf("invalid path %s: %v", e.Path, e.Cause)
}

func (e *PathContextError) Unwrap() error { return e.Cause }

// OperationContextError wraps an error originating inside a specific
// HTTP-method operation under a path. Method is the uppercase method
// (GET, POST, etc.).
//
// Use errors.As(err, &oce) to extract the method from a validation
// error chain without parsing the rendered message.
type OperationContextError struct {
	Method string
	Cause  error
}

func (e *OperationContextError) Error() string {
	return fmt.Sprintf("invalid operation %s: %v", e.Method, e.Cause)
}

func (e *OperationContextError) Unwrap() error { return e.Cause }
