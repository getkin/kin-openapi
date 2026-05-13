package openapi3

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// Error() text format: "invalid <section>: <inner>" — byte-identical to
// the fmt.Errorf wrapper this type replaced.
func TestSectionContextError_Error(t *testing.T) {
	inner := errors.New("value of version must be a non-empty string")
	e := &SectionContextError{Section: "info", Cause: inner}
	require.Equal(t, "invalid info: value of version must be a non-empty string", e.Error())
}

// Unwrap returns the inner error so errors.Is / errors.As walk the chain.
func TestSectionContextError_Unwrap(t *testing.T) {
	sentinel := errors.New("sentinel")
	e := &SectionContextError{Section: "components", Cause: sentinel}
	require.True(t, errors.Is(e, sentinel))
}

func TestPathContextError_Error(t *testing.T) {
	inner := errors.New("bad path item")
	e := &PathContextError{Path: "/users/{id}", Cause: inner}
	require.Equal(t, "invalid path /users/{id}: bad path item", e.Error())
}

func TestPathContextError_Unwrap(t *testing.T) {
	sentinel := errors.New("sentinel")
	e := &PathContextError{Path: "/x", Cause: sentinel}
	require.True(t, errors.Is(e, sentinel))
}

func TestOperationContextError_Error(t *testing.T) {
	inner := errors.New("bad operation")
	e := &OperationContextError{Method: "GET", Cause: inner}
	require.Equal(t, "invalid operation GET: bad operation", e.Error())
}

func TestOperationContextError_Unwrap(t *testing.T) {
	sentinel := errors.New("sentinel")
	e := &OperationContextError{Method: "POST", Cause: sentinel}
	require.True(t, errors.Is(e, sentinel))
}

// Three-layer chain (section + path + operation) is the typical shape
// for a schema-deep validation error inside paths. errors.As against
// each cluster pulls the corresponding context independently of the
// others; nothing depends on the order in which the layers were
// wrapped.
func TestSectionPathOperationChain_AsExtraction(t *testing.T) {
	leaf := errors.New("field const is for OpenAPI >=3.1")
	chain := &SectionContextError{
		Section: "paths",
		Cause: &PathContextError{
			Path: "/thing",
			Cause: &OperationContextError{
				Method: "GET",
				Cause:  leaf,
			},
		},
	}

	var sce *SectionContextError
	require.True(t, errors.As(chain, &sce))
	require.Equal(t, "paths", sce.Section)

	var pce *PathContextError
	require.True(t, errors.As(chain, &pce))
	require.Equal(t, "/thing", pce.Path)

	var oce *OperationContextError
	require.True(t, errors.As(chain, &oce))
	require.Equal(t, "GET", oce.Method)

	// And the leaf is still reachable via errors.Is.
	require.True(t, errors.Is(chain, leaf))

	// And the rendered string concatenates the layers exactly as the
	// previous fmt.Errorf chain did.
	require.Equal(t,
		"invalid paths: invalid path /thing: invalid operation GET: field const is for OpenAPI >=3.1",
		chain.Error(),
	)
}

// Mixed: a SectionContextError wrapping an arbitrary non-typed error
// still renders correctly. Guards against the typed wrapper changing
// behaviour when the underlying error is not one of our cluster types.
func TestSectionContextError_WrappingArbitrary(t *testing.T) {
	inner := fmt.Errorf("third-party error: %w", errors.New("x"))
	e := &SectionContextError{Section: "webhooks", Cause: inner}
	require.Equal(t, "invalid webhooks: third-party error: x", e.Error())
}
