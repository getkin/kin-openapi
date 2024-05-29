package openapi3

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

const identifierPattern = `^[a-zA-Z0-9._-]+$`

// IdentifierRegExp verifies whether Component object key matches 'identifierPattern' pattern, according to OpenAPI v3.x.
// However, to be able supporting legacy OpenAPI v2.x, there is a need to customize above pattern in order not to fail
// converted v2-v3 validation
var IdentifierRegExp = regexp.MustCompile(identifierPattern)

// ValidateIdentifier returns an error if the given component name does not match IdentifierRegExp.
func ValidateIdentifier(value string) error {
	if IdentifierRegExp.MatchString(value) {
		return nil
	}
	return fmt.Errorf("identifier %q is not supported by OpenAPIv3 standard (regexp: %q)", value, identifierPattern)
}

// Float64Ptr is a helper for defining OpenAPI schemas.
func Float64Ptr(value float64) *float64 {
	return &value
}

// BoolPtr is a helper for defining OpenAPI schemas.
func BoolPtr(value bool) *bool {
	return &value
}

// Int64Ptr is a helper for defining OpenAPI schemas.
func Int64Ptr(value int64) *int64 {
	return &value
}

// Uint64Ptr is a helper for defining OpenAPI schemas.
func Uint64Ptr(value uint64) *uint64 {
	return &value
}

type refPath interface {
	RefPath() *url.URL
}

// refersToSameDocument returns if the $ref refers to the same document.
//
// Documents in different directories will have distinct $ref values that resolve to
// the same document.
// For example, consider the 3 files:
//
//	/records.yaml
//	/root.yaml         $ref: records.yaml
//	/schema/other.yaml $ref: ../records.yaml
//
// The records.yaml reference in the 2 latter refers to the same document.
func refersToSameDocument(o1 refPath, o2 refPath) bool {
	if o1 == nil || o2 == nil {
		return false
	}

	r1 := o1.RefPath()
	r2 := o2.RefPath()

	// refURL is relative to the working directory & base spec file.
	return r1.String() == r2.String()
}

// referencesRootDocument returns if the $ref points to the root document of the OpenAPI spec.
//
// If the document has no location, perhaps loaded from data in memory, it always returns false.
func referencesRootDocument(doc *T, ref refPath) bool {
	if doc.url == nil || ref == nil {
		return false
	}

	refURL := *ref.RefPath()

	refURL.Path, _, _ = strings.Cut(refURL.Path, "#") // remove the document element reference

	// Check referenced element was in the root document.
	return doc.url.String() == refURL.String()
}
