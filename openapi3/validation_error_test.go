package openapi3_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

// Existing string-comparison consumers must keep working unchanged. The
// converted Info.Validate sites must produce the exact same Error()
// strings they used to produce as plain errors.New(...) values.
func TestValidationError_BackwardCompatibleErrorString(t *testing.T) {
	missingVersion := &openapi3.Info{Title: "x"}
	err := missingVersion.Validate(context.Background())
	require.EqualError(t, err, "value of version must be a non-empty string")

	missingTitle := &openapi3.Info{Version: "1.0.0"}
	err = missingTitle.Validate(context.Background())
	require.EqualError(t, err, "value of title must be a non-empty string")
}

// New consumers can switch on Code via errors.As.
func TestValidationError_StructuredCodeViaErrorsAs(t *testing.T) {
	missingVersion := &openapi3.Info{Title: "x"}
	err := missingVersion.Validate(context.Background())

	var ve *openapi3.ValidationError
	require.True(t, errors.As(err, &ve))
	require.Equal(t, "info-version-required", ve.Code)

	missingTitle := &openapi3.Info{Version: "1.0.0"}
	err = missingTitle.Validate(context.Background())
	require.True(t, errors.As(err, &ve))
	require.Equal(t, "info-title-required", ve.Code)
}

// MultiError already implements As() that recurses into elements, so a
// ValidationError wrapped inside a MultiError must remain reachable via
// errors.As. This pins that no special wiring is needed for the typed
// error to flow through the MultiError tree that doc.Validate returns.
func TestValidationError_FlowsThroughMultiError(t *testing.T) {
	inner := &openapi3.ValidationError{Code: "info-version-required", Message: "x"}
	me := openapi3.MultiError{errors.New("unrelated"), inner}

	var ve *openapi3.ValidationError
	require.True(t, errors.As(me, &ve))
	require.Equal(t, "info-version-required", ve.Code)
}
