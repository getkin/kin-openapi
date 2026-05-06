package openapi3

// ValidationError is the embedded base for every typed validation error
// emitted by the document validation walker (T.Validate, Info.Validate,
// Paths.Validate, etc.). Three layers of granularity are exposed; pick
// whichever the caller needs:
//
//  1. Base — *ValidationError. Catchall for "this is a validation
//     issue, here is the message". Reachable from any leaf via the As
//     method that each leaf implements.
//  2. Cluster — types like *RequiredFieldError or
//     *FieldVersionMismatchError. Group families of related failures
//     and expose the family-level metadata (Field, MinVersion, ...).
//     Wrap the underlying leaf via Unwrap, so errors.As can still walk
//     to the leaf.
//  3. Leaf — one type per call site (e.g. *InfoVersionRequired,
//     *LicenseIdentifierFieldFor31Plus). Lets callers match an exact
//     failure point without string comparison.
//
// All three are reachable from the same returned error through
// standard Go error wrapping (errors.As, errors.Is, errors.Unwrap),
// so a caller that only needs "is it a validation error?" stops at
// the base and a caller that wants "is it specifically license.identifier
// being used in 3.0?" matches the leaf.
//
// Backward compatibility: every site that today returns errors.New(msg)
// migrates to a leaf type that embeds ValidationError with Message set
// to the original string. (*ValidationError).Error() returns Message
// unchanged, so existing string-matching consumers see identical output.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string { return e.Message }

// asValidationError is a small helper used by every leaf type's As
// method to expose the embedded *ValidationError to errors.As. Defined
// once here so the leaf-side boilerplate stays a single line.
func asValidationError(target any, ve *ValidationError) bool {
	t, ok := target.(**ValidationError)
	if !ok {
		return false
	}
	*t = ve
	return true
}

// ---------------------------------------------------------------------
// Cluster types — group families of related failures.
// ---------------------------------------------------------------------

// RequiredFieldError clusters "X must be a non-empty value" failures
// across the spec (e.g. info.title, info.version, license.name,
// the openapi version string). The cluster carries the field path and
// wraps the per-site leaf so callers can match either:
//
//	var rfe *RequiredFieldError
//	if errors.As(err, &rfe) { /* knows the field */ }
//
//	var ivr *InfoVersionRequired
//	if errors.As(err, &ivr) { /* knows it's exactly info.version */ }
type RequiredFieldError struct {
	// Field is the JSON-pointer-style path of the required field
	// (e.g. "info.version", "license.name").
	Field string
	// Cause is the underlying leaf error. Walked by errors.Unwrap.
	Cause error
}

func (e *RequiredFieldError) Error() string { return e.Cause.Error() }
func (e *RequiredFieldError) Unwrap() error { return e.Cause }

// FieldVersionMismatchError clusters "field X is for OpenAPI >=Y"
// failures (3.1+ keywords used in 3.0 documents). Carries the field
// name and minimum version, wraps the per-site leaf.
type FieldVersionMismatchError struct {
	// Field is the field name flagged (e.g. "summary", "identifier").
	Field string
	// MinVersion is the minimum OpenAPI version that allows the field
	// (e.g. "3.1").
	MinVersion string
	// Cause is the underlying leaf error. Walked by errors.Unwrap.
	Cause error
}

func (e *FieldVersionMismatchError) Error() string { return e.Cause.Error() }
func (e *FieldVersionMismatchError) Unwrap() error { return e.Cause }

// ---------------------------------------------------------------------
// Leaf types — one per call site. Each embeds ValidationError for
// Error() and As-to-base, and is wrapped in its cluster type when
// returned from a validator.
// ---------------------------------------------------------------------

// InfoVersionRequired flags an empty info.version (RequiredFieldError cluster).
type InfoVersionRequired struct{ ValidationError }

func (e *InfoVersionRequired) As(target any) bool {
	return asValidationError(target, &e.ValidationError)
}

// InfoTitleRequired flags an empty info.title (RequiredFieldError cluster).
type InfoTitleRequired struct{ ValidationError }

func (e *InfoTitleRequired) As(target any) bool {
	return asValidationError(target, &e.ValidationError)
}

// LicenseIdentifierFieldFor31Plus flags use of license.identifier in
// an OAS 3.0 document (FieldVersionMismatchError cluster).
type LicenseIdentifierFieldFor31Plus struct{ ValidationError }

func (e *LicenseIdentifierFieldFor31Plus) As(target any) bool {
	return asValidationError(target, &e.ValidationError)
}

// InfoSummaryFieldFor31Plus flags use of info.summary in an OAS 3.0
// document (FieldVersionMismatchError cluster).
type InfoSummaryFieldFor31Plus struct{ ValidationError }

func (e *InfoSummaryFieldFor31Plus) As(target any) bool {
	return asValidationError(target, &e.ValidationError)
}

// ---------------------------------------------------------------------
// Constructors — the validator-side entry points. Build the leaf, wrap
// in the cluster, return the cluster (which exposes the leaf via
// Unwrap and the base via the leaf's As method).
// ---------------------------------------------------------------------

func newInfoVersionRequired() error {
	const msg = "value of version must be a non-empty string"
	return &RequiredFieldError{
		Field: "info.version",
		Cause: &InfoVersionRequired{ValidationError{Message: msg}},
	}
}

func newInfoTitleRequired() error {
	const msg = "value of title must be a non-empty string"
	return &RequiredFieldError{
		Field: "info.title",
		Cause: &InfoTitleRequired{ValidationError{Message: msg}},
	}
}

// newFieldFor31Plus dispatches errFieldFor31Plus's per-field message
// to the right typed leaf and wraps it in a FieldVersionMismatchError.
// Fields not yet typed fall back to a generic *ValidationError so the
// caller still gets a stable Message.
func newFieldFor31Plus(field string) error {
	msg := "field " + field + " is for OpenAPI >=3.1"
	var leaf error
	switch field {
	case "identifier":
		leaf = &LicenseIdentifierFieldFor31Plus{ValidationError{Message: msg}}
	case "summary":
		leaf = &InfoSummaryFieldFor31Plus{ValidationError{Message: msg}}
	default:
		// Untyped fallback — preserves the original Message and
		// still embeds *ValidationError, so errors.As(&ve) works.
		leaf = &ValidationError{Message: msg}
	}
	return &FieldVersionMismatchError{
		Field:      field,
		MinVersion: "3.1",
		Cause:      leaf,
	}
}
