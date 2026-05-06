package openapi3_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

// Existing string-comparison consumers must keep working unchanged.
// The leaf-typed sites in Info.Validate must produce the exact same
// Error() strings they used to produce as plain errors.New(...) values.
func TestValidationError_BackwardCompatibleErrorString(t *testing.T) {
	missingVersion := &openapi3.Info{Title: "x"}
	require.EqualError(t, missingVersion.Validate(context.Background()),
		"value of version must be a non-empty string")

	missingTitle := &openapi3.Info{Version: "1.0.0"}
	require.EqualError(t, missingTitle.Validate(context.Background()),
		"value of title must be a non-empty string")
}

// Three layers of granularity, all reachable from the same returned
// error: base ValidationError, cluster RequiredFieldError, and the
// per-site leaf type (e.g. *InfoVersionRequired).
func TestValidationError_ThreeLayers_RequiredField(t *testing.T) {
	err := (&openapi3.Info{Title: "x"}).Validate(context.Background())

	// Layer 1: cluster — carries field-level metadata.
	var rfe *openapi3.RequiredFieldError
	require.True(t, errors.As(err, &rfe))
	require.Equal(t, "info.version", rfe.Field)

	// Layer 2: leaf — exact identification of which site fired.
	var ivr *openapi3.InfoVersionRequired
	require.True(t, errors.As(err, &ivr))

	// Layer 3: base — catchall for "this is a validation error".
	var ve *openapi3.ValidationError
	require.True(t, errors.As(err, &ve))
	require.Equal(t, "value of version must be a non-empty string", ve.Message)
}

func TestValidationError_LeafDifferentiation(t *testing.T) {
	verErr := (&openapi3.Info{Title: "x"}).Validate(context.Background())
	titleErr := (&openapi3.Info{Version: "1.0.0"}).Validate(context.Background())

	// Title's leaf type does NOT match the version's leaf type, even
	// though both flow through the same RequiredFieldError cluster.
	var ivr *openapi3.InfoVersionRequired
	require.True(t, errors.As(verErr, &ivr))
	require.False(t, errors.As(titleErr, &ivr))

	var itr *openapi3.InfoTitleRequired
	require.True(t, errors.As(titleErr, &itr))
	require.False(t, errors.As(verErr, &itr))

	// Both still share the cluster type.
	var rfe *openapi3.RequiredFieldError
	require.True(t, errors.As(verErr, &rfe))
	require.Equal(t, "info.version", rfe.Field)
	require.True(t, errors.As(titleErr, &rfe))
	require.Equal(t, "info.title", rfe.Field)
}

// FieldVersionMismatchError cluster, exercised by the existing
// errFieldFor31Plus helper (license.identifier in a 3.0 doc).
func TestValidationError_ThreeLayers_FieldVersionMismatch(t *testing.T) {
	doc := &openapi3.T{
		OpenAPI: "3.0.3",
		Info: &openapi3.Info{
			Title:   "x",
			Version: "1.0.0",
			License: &openapi3.License{
				Name:       "MIT",
				Identifier: "MIT", // 3.1+ only
			},
		},
		Paths: openapi3.NewPaths(),
	}
	err := doc.Validate(context.Background())
	require.Error(t, err)

	var fvm *openapi3.FieldVersionMismatchError
	require.True(t, errors.As(err, &fvm))
	require.Equal(t, "identifier", fvm.Field)
	require.Equal(t, "3.1", fvm.MinVersion)

	var lif *openapi3.LicenseIdentifierFieldFor31Plus
	require.True(t, errors.As(err, &lif))

	var ve *openapi3.ValidationError
	require.True(t, errors.As(err, &ve))
	require.Contains(t, ve.Message, "identifier")
}

// Untyped fields (those not yet assigned a leaf in newFieldFor31Plus's
// switch) still flow through the cluster + base, so callers retain
// the same discrimination layers as their typed cousins. Only the
// per-leaf type isn't there to assert against.
func TestValidationError_FieldVersionMismatch_UntypedFallback(t *testing.T) {
	// "webhooks" is an existing untyped 3.1+-only field — see
	// openapi3.go's errFieldFor31Plus("webhooks") site.
	doc := &openapi3.T{
		OpenAPI: "3.0.3",
		Info:    &openapi3.Info{Title: "x", Version: "1.0.0"},
		Paths:    openapi3.NewPaths(),
		Webhooks: map[string]*openapi3.PathItem{},
	}
	err := doc.Validate(context.Background())
	require.Error(t, err)

	var fvm *openapi3.FieldVersionMismatchError
	require.True(t, errors.As(err, &fvm))
	require.Equal(t, "webhooks", fvm.Field)

	var ve *openapi3.ValidationError
	require.True(t, errors.As(err, &ve))
}

// MultiError already implements As() that recurses into elements, so a
// typed validation error wrapped inside a MultiError must remain
// reachable. This pins that no special wiring is needed for the typed
// errors to flow through the MultiError tree.
func TestValidationError_FlowsThroughMultiError(t *testing.T) {
	leaf := &openapi3.InfoVersionRequired{
		ValidationError: openapi3.ValidationError{Message: "x"},
	}
	cluster := &openapi3.RequiredFieldError{Field: "info.version", Cause: leaf}
	me := openapi3.MultiError{errors.New("unrelated"), cluster}

	var rfe *openapi3.RequiredFieldError
	require.True(t, errors.As(me, &rfe))

	var ivr *openapi3.InfoVersionRequired
	require.True(t, errors.As(me, &ivr))

	var ve *openapi3.ValidationError
	require.True(t, errors.As(me, &ve))
}
