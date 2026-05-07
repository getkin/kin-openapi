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
		OpenAPI:  "3.0.3",
		Info:     &openapi3.Info{Title: "x", Version: "1.0.0"},
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

// Spot-check the rest of the converted call sites: each required field
// produces its own leaf type plus the shared cluster, and each typed
// 3.1+-only schema field produces its own leaf type plus the shared
// cluster.
func TestValidationError_AllRequiredFieldLeaves(t *testing.T) {
	type tc struct {
		name      string
		doc       *openapi3.T
		leafCheck func(t *testing.T, err error)
		field     string
		message   string
	}
	cases := []tc{
		{
			name: "openapi version required",
			doc:  &openapi3.T{},
			leafCheck: func(t *testing.T, err error) {
				var l *openapi3.OpenAPIVersionRequired
				require.True(t, errors.As(err, &l))
			},
			field:   "openapi",
			message: "value of openapi must be a non-empty string",
		},
		{
			name: "license name required",
			doc: &openapi3.T{
				OpenAPI: "3.0.3",
				Info: &openapi3.Info{
					Title: "x", Version: "1.0.0",
					License: &openapi3.License{},
				},
				Paths: openapi3.NewPaths(),
			},
			leafCheck: func(t *testing.T, err error) {
				var l *openapi3.LicenseNameRequired
				require.True(t, errors.As(err, &l))
			},
			field:   "license.name",
			message: "value of license name must be a non-empty string",
		},
		{
			name: "server url required",
			doc: &openapi3.T{
				OpenAPI: "3.0.3",
				Info:    &openapi3.Info{Title: "x", Version: "1.0.0"},
				Paths:   openapi3.NewPaths(),
				Servers: openapi3.Servers{&openapi3.Server{}},
			},
			leafCheck: func(t *testing.T, err error) {
				var l *openapi3.ServerURLRequired
				require.True(t, errors.As(err, &l))
			},
			field:   "server.url",
			message: "value of url must be a non-empty string",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.doc.Validate(context.Background())
			require.Error(t, err)

			var rfe *openapi3.RequiredFieldError
			require.True(t, errors.As(err, &rfe))
			require.Equal(t, c.field, rfe.Field)

			c.leafCheck(t, err)

			var ve *openapi3.ValidationError
			require.True(t, errors.As(err, &ve))
			require.Equal(t, c.message, ve.Message)
		})
	}
}

// Spot-check a couple of the schema-field leaves that flow through
// errFieldFor31Plus (used by schema.go's reject() helper). Full
// per-field coverage is left to the package's existing schema_test.go.
func TestValidationError_SchemaFieldFor31PlusLeaves(t *testing.T) {
	type tc struct {
		name      string
		schema    *openapi3.Schema
		leafCheck func(t *testing.T, err error)
		field     string
	}
	cases := []tc{
		{
			name:   "const",
			schema: &openapi3.Schema{Const: "x"},
			leafCheck: func(t *testing.T, err error) {
				var l *openapi3.ConstFieldFor31Plus
				require.True(t, errors.As(err, &l))
			},
			field: "const",
		},
		{
			name: "patternProperties",
			schema: &openapi3.Schema{
				PatternProperties: map[string]*openapi3.SchemaRef{"foo": {Value: &openapi3.Schema{}}},
			},
			leafCheck: func(t *testing.T, err error) {
				var l *openapi3.PatternPropertiesFieldFor31Plus
				require.True(t, errors.As(err, &l))
			},
			field: "patternProperties",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.schema.Validate(context.Background())
			require.Error(t, err)

			var fvm *openapi3.FieldVersionMismatchError
			require.True(t, errors.As(err, &fvm))
			require.Equal(t, c.field, fvm.Field)
			require.Equal(t, "3.1", fvm.MinVersion)

			c.leafCheck(t, err)

			var ve *openapi3.ValidationError
			require.True(t, errors.As(err, &ve))
		})
	}
}

// Cluster types wrap their leaves through standard Go error wrapping,
// so errors.Unwrap walks from the cluster to the leaf in a single step.
// Pinning this directly (rather than only via errors.As) demonstrates
// that the chain follows the conventional Unwrap contract — useful for
// any consumer that walks the error tree by hand instead of asking for
// a specific type.
func TestValidationError_UnwrapWalksClusterToLeaf(t *testing.T) {
	// RequiredFieldError cluster wrapping an InfoVersionRequired leaf.
	verErr := (&openapi3.Info{Title: "x"}).Validate(context.Background())

	// The returned error IS the cluster, not the leaf.
	rfe, ok := verErr.(*openapi3.RequiredFieldError)
	require.True(t, ok, "validator returns the cluster type")
	require.Equal(t, "info.version", rfe.Field)

	// errors.Unwrap takes us to the leaf in one step.
	leaf := errors.Unwrap(verErr)
	require.NotNil(t, leaf)
	_, isLeaf := leaf.(*openapi3.InfoVersionRequired)
	require.True(t, isLeaf, "Unwrap reaches the leaf type")

	// The leaf is terminal — nothing further to unwrap.
	require.Nil(t, errors.Unwrap(leaf), "leaf has no inner error")

	// Same shape for the FieldVersionMismatchError cluster wrapping a
	// LicenseIdentifierFieldFor31Plus leaf.
	doc := &openapi3.T{
		OpenAPI: "3.0.3",
		Info: &openapi3.Info{
			Title: "x", Version: "1.0.0",
			License: &openapi3.License{Name: "MIT", Identifier: "MIT"},
		},
		Paths: openapi3.NewPaths(),
	}
	docErr := doc.Validate(context.Background())

	// doc.Validate wraps the License error in MultiError variants. Walk
	// to the FieldVersionMismatchError cluster via errors.As (since
	// MultiError sits between).
	var fvm *openapi3.FieldVersionMismatchError
	require.True(t, errors.As(docErr, &fvm))
	require.Equal(t, "identifier", fvm.Field)

	// From the cluster, Unwrap reaches the leaf directly.
	licenseLeaf := errors.Unwrap(fvm)
	require.NotNil(t, licenseLeaf)
	_, isLicenseLeaf := licenseLeaf.(*openapi3.LicenseIdentifierFieldFor31Plus)
	require.True(t, isLicenseLeaf, "Unwrap reaches the leaf type")
	require.Nil(t, errors.Unwrap(licenseLeaf), "leaf has no inner error")
}

// Origin is populated on cluster types when the document was loaded
// with Loader.IncludeOrigin = true. The cluster carries the offending
// element's Origin (info, license, server, schema, ...) so consumers
// can attach file/line/column to a finding without re-walking the doc.
func TestValidationError_OriginPopulatedOnLoaderTracking(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IncludeOrigin = true
	doc, err := loader.LoadFromData([]byte(`
openapi: 3.0.3
info:
  title: x
  version: ""
paths: {}
`))
	require.NoError(t, err)

	verr := doc.Validate(context.Background())
	var rfe *openapi3.RequiredFieldError
	require.True(t, errors.As(verr, &rfe))
	require.Equal(t, "info.version", rfe.Field)
	require.NotNil(t, rfe.Origin, "cluster should carry info.Origin when loader tracks origins")
	require.NotNil(t, rfe.Origin.Key, "Origin.Key set by the loader")
	// File is empty for LoadFromData (no path associated); LoadFromFile
	// would populate it. Line/Column are populated either way.
	require.Greater(t, rfe.Origin.Key.Line, 0)
}

// Without IncludeOrigin, Origin is nil — we don't fabricate location
// info that wasn't tracked.
func TestValidationError_OriginNilWithoutLoaderTracking(t *testing.T) {
	loader := openapi3.NewLoader()
	// IncludeOrigin defaults to false
	doc, err := loader.LoadFromData([]byte(`
openapi: 3.0.3
info:
  title: x
  version: ""
paths: {}
`))
	require.NoError(t, err)

	verr := doc.Validate(context.Background())
	var rfe *openapi3.RequiredFieldError
	require.True(t, errors.As(verr, &rfe))
	require.Nil(t, rfe.Origin, "Origin should be nil when loader didn't track origins")
}

// Document-root fields (openapi, webhooks, jsonSchemaDialect) live on
// *T which the loader doesn't track, so their Origin is always nil
// even when IncludeOrigin is enabled. Pinned so callers know to fall
// back to file-only when the field is at the doc root.
func TestValidationError_OriginNilForDocumentRootFields(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IncludeOrigin = true
	doc, err := loader.LoadFromData([]byte(`
openapi: ""
info:
  title: x
  version: "1"
paths: {}
`))
	require.NoError(t, err)

	verr := doc.Validate(context.Background())
	var rfe *openapi3.RequiredFieldError
	require.True(t, errors.As(verr, &rfe))
	require.Equal(t, "openapi", rfe.Field)
	require.Nil(t, rfe.Origin, "doc-root fields have no Origin (loader doesn't track *T)")
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
