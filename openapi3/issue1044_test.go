package openapi3_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

// Go's regexp is RE2 and rejects the lookarounds ECMA 262 allows, so this
// pattern cannot be compiled.
const issue1044Pattern = `^((?!-)[A-Za-z0-9-]{1,63}(?<!-)\.)+[A-Za-z]{2,6}$`

func TestIssue1044(t *testing.T) {
	schema := openapi3.NewStringSchema().WithPattern(issue1044Pattern)

	err := schema.VisitJSON("example.com", openapi3.MultiErrors())

	var patternErr *openapi3.SchemaPatternRegexError
	require.ErrorAs(t, err, &patternErr)
	require.Equal(t, issue1044Pattern, patternErr.Pattern)
}

// A custom compiler reports failure by returning no matcher at all.
func TestIssue1044CustomRegexCompiler(t *testing.T) {
	schema := openapi3.NewStringSchema().WithPattern(`^whatever$`)
	compile := func(string) (openapi3.RegexMatcher, error) {
		return nil, errors.New("compiler said no")
	}

	err := schema.VisitJSON("example.com", openapi3.MultiErrors(), openapi3.SetSchemaRegexCompiler(compile))

	require.ErrorContains(t, err, "compiler said no")
}
