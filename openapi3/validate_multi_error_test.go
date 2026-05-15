package openapi3_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

// twoBadPathsSpec is a document with two independent path-level problems:
// both operations are missing the required "responses" object.
const twoBadPathsSpec = `
openapi: 3.0.0
info: { title: t, version: "1" }
paths:
  /a:
    get: {}
  /b:
    get: {}
`

// twoBadSectionsSpec has problems in two different document sections.
const twoBadSectionsSpec = `
openapi: 3.0.0
info: { title: t, version: "1" }
paths:
  /a:
    get: {}
components:
  schemas:
    "bad name with spaces":
      type: string
`

func loadDoc(t *testing.T, src string) *openapi3.T {
	t.Helper()
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(src))
	require.NoError(t, err)
	return doc
}

// countLeaves returns the number of non-MultiError leaves in err, walking
// through MultiError nodes and Unwrap chains. The validation tree groups
// per-section MultiErrors under SectionValidationError wrappers, so counting
// leaves gives the total number of independent problems regardless of shape.
func countLeaves(err error) int {
	if err == nil {
		return 0
	}
	if me, ok := err.(openapi3.MultiError); ok {
		n := 0
		for _, e := range me {
			n += countLeaves(e)
		}
		return n
	}
	if u, ok := err.(interface{ Unwrap() error }); ok {
		if inner := u.Unwrap(); inner != nil {
			return countLeaves(inner)
		}
	}
	return 1
}

func TestValidate_MultiError_Off_PreservesFailFast(t *testing.T) {
	// Without EnableMultiError, Validate returns the first error and stops.
	// The returned error is a single typed error, not a MultiError.
	doc := loadDoc(t, twoBadPathsSpec)
	err := doc.Validate(context.Background())
	require.Error(t, err)

	var me openapi3.MultiError
	require.False(t, errors.As(err, &me), "without EnableMultiError, result must not be a MultiError")
}

func TestValidate_MultiError_On_AggregatesAcrossPaths(t *testing.T) {
	// With EnableMultiError, Validate aggregates problems across paths.
	// The result is a tree: T-level MultiError -> SectionValidationError("paths")
	// -> Paths-level MultiError -> PathValidationError per bad path. Walking
	// the tree must yield one leaf per independent problem.
	doc := loadDoc(t, twoBadPathsSpec)
	err := doc.Validate(context.Background(), openapi3.EnableMultiError())
	require.Error(t, err)

	var me openapi3.MultiError
	require.True(t, errors.As(err, &me), "expected MultiError")
	require.Equal(t, 2, countLeaves(err), "expected one leaf per bad path")

	require.ErrorContains(t, err, "/a")
	require.ErrorContains(t, err, "/b")
	// ErrorContains can't express an exact count, so for the "each defect
	// mentions 'responses' exactly twice" check we materialize the error
	// string into a local first; CI grep rejects require lines that read
	// the error string inline.
	combined := err.Error()
	require.Equal(t, 2, strings.Count(combined, "responses"),
		"each defect should mention the missing 'responses' object")
}

func TestValidate_MultiError_On_AggregatesAcrossSections(t *testing.T) {
	// Problems in different document sections are also aggregated.
	doc := loadDoc(t, twoBadSectionsSpec)
	err := doc.Validate(context.Background(), openapi3.EnableMultiError())
	require.Error(t, err)

	var me openapi3.MultiError
	require.True(t, errors.As(err, &me), "expected MultiError")
	require.Equal(t, 2, len(me),
		"expected one error per affected section")

	// Confirm both sections are represented in the chain.
	var foundComponents, foundPaths bool
	for _, e := range me {
		var sec *openapi3.SectionValidationError
		if errors.As(e, &sec) {
			switch sec.Section {
			case "components":
				foundComponents = true
			case "paths":
				foundPaths = true
			}
		}
	}
	require.True(t, foundComponents, "components section error missing")
	require.True(t, foundPaths, "paths section error missing")
}

func TestValidate_MultiError_On_SingleError_StillReturnsMultiError(t *testing.T) {
	// With EnableMultiError, even a single-defect spec returns a MultiError
	// (containing one element). MultiError.Error() of a single element is
	// byte-identical to the contained error's Error(), so the string output
	// is unchanged; only the static type differs.
	const oneBadPathSpec = `
openapi: 3.0.0
info: { title: t, version: "1" }
paths:
  /a:
    get: {}
`
	doc := loadDoc(t, oneBadPathSpec)
	err := doc.Validate(context.Background(), openapi3.EnableMultiError())
	require.Error(t, err)

	var me openapi3.MultiError
	require.True(t, errors.As(err, &me))
	require.Len(t, me, 1)

	// MultiError.Is / MultiError.As walk into the contained errors, so typed
	// consumers using errors.As keep working seamlessly.
	var sec *openapi3.SectionValidationError
	require.True(t, errors.As(err, &sec), "errors.As should walk through MultiError")
}

func TestValidate_MultiError_On_NoErrors_ReturnsNil(t *testing.T) {
	// A well-formed document still returns nil, regardless of the option.
	const goodSpec = `
openapi: 3.0.0
info: { title: t, version: "1" }
paths:
  /a:
    get:
      responses:
        "200":
          description: ok
`
	doc := loadDoc(t, goodSpec)
	require.NoError(t, doc.Validate(context.Background(), openapi3.EnableMultiError()))
}
