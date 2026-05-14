package openapi3

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// Error() text format: "invalid <section>: <inner>" — byte-identical to
// the fmt.Errorf wrapper this type replaced.
func TestSectionValidationError_Error(t *testing.T) {
	inner := errors.New("value of version must be a non-empty string")
	e := &SectionValidationError{Section: "info", Cause: inner}
	require.Equal(t, "invalid info: value of version must be a non-empty string", e.Error())
}

// Unwrap returns the inner error so errors.Is / errors.As walk the chain.
func TestSectionValidationError_Unwrap(t *testing.T) {
	sentinel := errors.New("sentinel")
	e := &SectionValidationError{Section: "components", Cause: sentinel}
	require.True(t, errors.Is(e, sentinel))
}

func TestPathValidationError_Error(t *testing.T) {
	inner := errors.New("bad path item")
	e := &PathValidationError{Path: "/users/{id}", Cause: inner}
	require.Equal(t, "invalid path /users/{id}: bad path item", e.Error())
}

func TestPathValidationError_Unwrap(t *testing.T) {
	sentinel := errors.New("sentinel")
	e := &PathValidationError{Path: "/x", Cause: sentinel}
	require.True(t, errors.Is(e, sentinel))
}

func TestOperationValidationError_Error(t *testing.T) {
	inner := errors.New("bad operation")
	e := &OperationValidationError{Method: "GET", Cause: inner}
	require.Equal(t, "invalid operation GET: bad operation", e.Error())
}

func TestOperationValidationError_Unwrap(t *testing.T) {
	sentinel := errors.New("sentinel")
	e := &OperationValidationError{Method: "POST", Cause: sentinel}
	require.True(t, errors.Is(e, sentinel))
}

// Three-layer chain (section + path + operation) is the typical shape
// for a schema-deep validation error inside paths. errors.As against
// each cluster pulls the corresponding context independently of the
// others; nothing depends on the order in which the layers were
// wrapped.
func TestSectionPathOperationChain_AsExtraction(t *testing.T) {
	leaf := errors.New("field const is for OpenAPI >=3.1")
	chain := &SectionValidationError{
		Section: "paths",
		Cause: &PathValidationError{
			Path: "/thing",
			Cause: &OperationValidationError{
				Method: "GET",
				Cause:  leaf,
			},
		},
	}

	var sve *SectionValidationError
	require.True(t, errors.As(chain, &sve))
	require.Equal(t, "paths", sve.Section)

	var pve *PathValidationError
	require.True(t, errors.As(chain, &pve))
	require.Equal(t, "/thing", pve.Path)

	var ove *OperationValidationError
	require.True(t, errors.As(chain, &ove))
	require.Equal(t, "GET", ove.Method)

	// And the leaf is still reachable via errors.Is.
	require.True(t, errors.Is(chain, leaf))

	// And the rendered string concatenates the layers exactly as the
	// previous fmt.Errorf chain did.
	require.Equal(t,
		"invalid paths: invalid path /thing: invalid operation GET: field const is for OpenAPI >=3.1",
		chain.Error(),
	)
}

// Mixed: a SectionValidationError wrapping an arbitrary non-typed error
// still renders correctly. Guards against the typed wrapper changing
// behaviour when the underlying error is not one of our cluster types.
func TestSectionValidationError_WrappingArbitrary(t *testing.T) {
	inner := fmt.Errorf("third-party error: %w", errors.New("x"))
	e := &SectionValidationError{Section: "webhooks", Cause: inner}
	require.Equal(t, "invalid webhooks: third-party error: x", e.Error())
}
