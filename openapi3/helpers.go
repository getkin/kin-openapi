package openapi3

import (
	"fmt"
	"net/url"
	"path"
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

type componentRef interface {
	RefString() string
	RefPath() *url.URL
	ComponentType() string
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
func refersToSameDocument(o1 componentRef, o2 componentRef) bool {
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
func referencesRootDocument(doc *T, ref componentRef) bool {
	if doc.url == nil || ref == nil {
		return false
	}

	refURL := *ref.RefPath()
	refURL.Path, _, _ = strings.Cut(refURL.Path, "#") // remove the document element reference

	// Check referenced element was in the root document.
	return doc.url.String() == refURL.String()
}

// MatchesSchemaInRootDocument returns if the given schema is identical
// to a schema defined in the root document's '#/components/schemas'.
// It returns a reference to the schema in the form
// '#/components/schemas/NameXXX'
//
// Of course given it a schema from the root document will always match.
//
// https://swagger.io/docs/specification/using-ref/#syntax
//
// Case 1: Directly via
//
//	../openapi.yaml#/components/schemas/Record
//
// Case 2: Or indirectly by using a $ref which matches a schema
// in the root document's '#/components/schemas' using the same
// $ref.
//
// In schemas/record.yaml
//
//	$ref: ./record.yaml
//
// In openapi.yaml
//
//	  components:
//	    schemas:
//		  Record:
//		    $ref: schemas/record.yaml
func MatchesComponentInRootDocument(doc *T, ref componentRef) (string, bool) {
	// Case 1:
	// Something like: ../another-folder/document.json#/myElement
	if isRemoteReference(ref.RefString()) && isRootComponentReference(ref.RefString(), ref.ComponentType()) {
		// Determine if it is *this* root doc.
		if referencesRootDocument(doc, ref) {
			_, name, _ := strings.Cut(ref.RefString(), path.Join("#/components/", ref.ComponentType()))

			return path.Join("#/components/", ref.ComponentType(), name), true
		}
	}

	// If there are no schemas defined in the root document return early.
	if doc.Components == nil || doc.Components.Schemas == nil {
		return "", false
	}

	// Case 2:
	// Something like: ../openapi.yaml#/components/schemas/myElement
	for name, s := range doc.Components.Schemas {
		// Must be a reference to a YAML file.
		if !isWholeDocumentReference(s.Ref) {
			continue
		}

		// Is the schema a ref to the same resource.
		if !refersToSameDocument(s, ref) {
			continue
		}

		// Transform the remote ref to the equivalent schema in the root document.
		return path.Join("#/components/", ref.ComponentType(), name), true
	}

	return "", false
}

// isElementReference takes a $ref value and checks if it references a specific element.
func isElementReference(ref string) bool {
	return ref != "" && !isWholeDocumentReference(ref)
}

// isSchemaReference takes a $ref value and checks if it references a schema element.
func isRootComponentReference(ref string, compType string) bool {
	return isElementReference(ref) && strings.Contains(ref, path.Join("#/components/", compType))
}

// isWholeDocumentReference takes a $ref value and checks if it is whole document reference.
func isWholeDocumentReference(ref string) bool {
	return ref != "" && !strings.ContainsAny(ref, "#")
}

// isRemoteReference takes a $ref value and checks if it is remote reference.
func isRemoteReference(ref string) bool {
	return ref != "" && !strings.HasPrefix(ref, "#") && !isURLReference(ref)
}

// isURLReference takes a $ref value and checks if it is URL reference.
func isURLReference(ref string) bool {
	return strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") || strings.HasPrefix(ref, "//")
}
