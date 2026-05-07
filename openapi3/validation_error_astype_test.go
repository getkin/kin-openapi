//go:build go1.26

// Build-tagged go1.26 because errors.AsType was introduced in Go 1.26.
// The file compiles only on toolchains where the function exists; on
// older toolchains it's silently excluded.

package openapi3_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

// errors.AsType is the generic complement to errors.As — same chain
// walk, but returns the typed value directly instead of populating an
// out-parameter. Demonstrating it works with the validation-error
// hierarchy at all three layers (cluster, leaf, base) confirms the
// design is idiomatic under Go 1.26's improved errors API.
func TestValidationError_AsTypeWalksAllLayers(t *testing.T) {
	err := (&openapi3.Info{Title: "x"}).Validate(context.Background())

	// Cluster.
	rfe, ok := errors.AsType[*openapi3.RequiredFieldError](err)
	require.True(t, ok)
	require.Equal(t, "info.version", rfe.Field)

	// Leaf — reached via errors.Unwrap-walking, same as errors.As.
	_, ok = errors.AsType[*openapi3.InfoVersionRequired](err)
	require.True(t, ok)

	// Base — reached via the leaf's As method exposing the embedded
	// *ValidationError, same as errors.As.
	ve, ok := errors.AsType[*openapi3.ValidationError](err)
	require.True(t, ok)
	require.Equal(t, "value of version must be a non-empty string", ve.Message)
}
