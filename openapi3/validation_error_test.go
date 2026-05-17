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
		{
			name: "responses non-empty required",
			doc: &openapi3.T{
				OpenAPI: "3.0.3",
				Info:    &openapi3.Info{Title: "x", Version: "1.0.0"},
				Paths: openapi3.NewPaths(openapi3.WithPath("/p", &openapi3.PathItem{
					Get: &openapi3.Operation{Responses: openapi3.NewResponses()},
				})),
			},
			leafCheck: func(t *testing.T, err error) {
				var l *openapi3.ResponsesNonEmptyRequired
				require.True(t, errors.As(err, &l))
			},
			field:   "responses",
			message: "the responses object MUST contain at least one response code",
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
// *T, which now carries an Origin when IncludeOrigin is set. Their
// RequiredFieldError / FieldVersionMismatchError therefore carries the
// document's Origin: scalar root fields resolve precisely via
// Origin.Fields; object/missing root fields fall back to Origin.Key.
func TestValidationError_OriginForDocumentRootFields(t *testing.T) {
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
	require.NotNil(t, rfe.Origin, "doc-root fields now carry the document's Origin")
	require.Same(t, doc.Origin, rfe.Origin, "the error carries T.Origin")
	require.Greater(t, rfe.Origin.Fields["openapi"].Line, 0,
		`Origin.Fields["openapi"] locates the openapi: line`)
}

// SchemaValueError clusters "<schema field>'s example/default value
// doesn't satisfy the schema's own constraints" failures. Reach the
// cluster via errors.As, the underlying *SchemaError via Unwrap (or
// nested errors.As), and the cluster's metadata via cluster.ValueKind.
func TestValidationError_SchemaValueErrorOnInvalidExample(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IncludeOrigin = true
	doc, err := loader.LoadFromData([]byte(`
openapi: 3.0.3
info: {title: t, version: '1'}
paths:
  /thing:
    get:
      parameters:
        - name: token
          in: query
          example: too-long
          schema:
            type: string
            maxLength: 4
      responses:
        "200": {description: ok}
`))
	require.NoError(t, err)

	verr := doc.Validate(context.Background())
	require.Error(t, verr)

	// Cluster is reachable.
	var sve *openapi3.SchemaValueError
	require.True(t, errors.As(verr, &sve))
	require.Equal(t, "example", sve.ValueKind)
	require.NotNil(t, sve.Origin, "parameter Origin should be carried through")

	// Underlying *SchemaError is reachable via the Unwrap chain.
	var se *openapi3.SchemaError
	require.True(t, errors.As(verr, &se))

	// Error() prefixes the cluster's ValueKind to keep the historical
	// "invalid example: ..." message format byte-identical.
	require.Contains(t, sve.Error(), "invalid example: ")
}

// Pin RequiredFieldError cluster + leaf reachability for required-field
// validation on operations and external docs.
func TestValidationError_FollowupRequiredFieldLeaves(t *testing.T) {
	type tc struct {
		name      string
		doc       *openapi3.T
		field     string
		message   string
		leafCheck func(t *testing.T, err error)
	}
	cases := []tc{
		{
			name: "operation responses required",
			doc: &openapi3.T{
				OpenAPI: "3.0.3",
				Info:    &openapi3.Info{Title: "x", Version: "1.0.0"},
				Paths: openapi3.NewPaths(openapi3.WithPath("/p", &openapi3.PathItem{
					Get: &openapi3.Operation{}, // no Responses
				})),
			},
			field:   "operation.responses",
			message: "value of responses must be an object",
			leafCheck: func(t *testing.T, err error) {
				var l *openapi3.OperationResponsesRequired
				require.True(t, errors.As(err, &l))
			},
		},
		{
			name: "external docs url required",
			doc: &openapi3.T{
				OpenAPI:      "3.0.3",
				Info:         &openapi3.Info{Title: "x", Version: "1.0.0"},
				Paths:        openapi3.NewPaths(),
				ExternalDocs: &openapi3.ExternalDocs{}, // empty URL
			},
			field:   "externalDocs.url",
			message: "url is required",
			leafCheck: func(t *testing.T, err error) {
				var l *openapi3.ExternalDocsURLRequired
				require.True(t, errors.As(err, &l))
			},
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

// Pin MutuallyExclusiveFieldsError cluster + leaf reachability for the
// four sites where two fields are forbidden from being set together.
func TestValidationError_MutuallyExclusiveFieldsLeaves(t *testing.T) {
	t.Run("example value vs externalValue", func(t *testing.T) {
		ex := &openapi3.Example{Value: "v", ExternalValue: "https://x"}
		err := ex.Validate(context.Background())
		require.EqualError(t, err, "value and externalValue are mutually exclusive")

		var mef *openapi3.MutuallyExclusiveFieldsError
		require.True(t, errors.As(err, &mef))
		require.Equal(t, "value", mef.Field1)
		require.Equal(t, "externalValue", mef.Field2)

		var leaf *openapi3.ExampleValueExternalValueExclusive
		require.True(t, errors.As(err, &leaf))

		var ve *openapi3.ValidationError
		require.True(t, errors.As(err, &ve))
	})

	t.Run("license url vs identifier", func(t *testing.T) {
		// identifier is a 3.1+ field; opt in so the URL/identifier check
		// is the one that fires.
		lic := &openapi3.License{Name: "MIT", URL: "https://x", Identifier: "MIT"}
		err := lic.Validate(context.Background(), openapi3.IsOpenAPI31OrLater())
		require.EqualError(t, err, "license must not specify both 'url' and 'identifier'")

		var mef *openapi3.MutuallyExclusiveFieldsError
		require.True(t, errors.As(err, &mef))
		require.Equal(t, "url", mef.Field1)
		require.Equal(t, "identifier", mef.Field2)

		var leaf *openapi3.LicenseURLIdentifierExclusive
		require.True(t, errors.As(err, &leaf))
	})

	t.Run("link operationId vs operationRef", func(t *testing.T) {
		link := &openapi3.Link{OperationID: "getX", OperationRef: "#/x"}
		err := link.Validate(context.Background())
		require.EqualError(t, err, `operationId "getX" and operationRef "#/x" are mutually exclusive`)

		var mef *openapi3.MutuallyExclusiveFieldsError
		require.True(t, errors.As(err, &mef))
		require.Equal(t, "operationId", mef.Field1)
		require.Equal(t, "operationRef", mef.Field2)

		var leaf *openapi3.LinkOperationIDRefExclusive
		require.True(t, errors.As(err, &leaf))
	})

	t.Run("schema readOnly vs writeOnly", func(t *testing.T) {
		schema := &openapi3.Schema{ReadOnly: true, WriteOnly: true}
		err := schema.Validate(context.Background())
		require.EqualError(t, err, "a property MUST NOT be marked as both readOnly and writeOnly being true")

		var mef *openapi3.MutuallyExclusiveFieldsError
		require.True(t, errors.As(err, &mef))
		require.Equal(t, "readOnly", mef.Field1)
		require.Equal(t, "writeOnly", mef.Field2)

		var leaf *openapi3.SchemaReadOnlyWriteOnlyExclusive
		require.True(t, errors.As(err, &leaf))
	})
}

// Pin ForbiddenFieldError cluster + leaf reachability for the four
// sites where a field is set but the spec forbids it in that context.
func TestValidationError_ForbiddenFieldLeaves(t *testing.T) {
	t.Run("header.name forbidden", func(t *testing.T) {
		h := &openapi3.Header{Parameter: openapi3.Parameter{Name: "X-Trace"}}
		err := h.Validate(context.Background())
		require.EqualError(t, err,
			"header 'name' MUST NOT be specified, it is given in the corresponding headers map")

		var ffe *openapi3.ForbiddenFieldError
		require.True(t, errors.As(err, &ffe))
		require.Equal(t, "name", ffe.Field)

		var leaf *openapi3.HeaderNameForbidden
		require.True(t, errors.As(err, &leaf))

		var ve *openapi3.ValidationError
		require.True(t, errors.As(err, &ve))
	})

	t.Run("header.in forbidden", func(t *testing.T) {
		h := &openapi3.Header{Parameter: openapi3.Parameter{In: "header"}}
		err := h.Validate(context.Background())
		require.EqualError(t, err,
			"header 'in' MUST NOT be specified, it is implicitly in header")

		var ffe *openapi3.ForbiddenFieldError
		require.True(t, errors.As(err, &ffe))
		require.Equal(t, "in", ffe.Field)

		var leaf *openapi3.HeaderInForbidden
		require.True(t, errors.As(err, &leaf))
	})
}

// Pin parameter.name and apiKey securityScheme.name leaves directly:
// these check Validate on a single component (without wiring a full
// document around it).
func TestValidationError_ParameterAndAPIKeyNameLeaves(t *testing.T) {
	t.Run("parameter name required", func(t *testing.T) {
		err := (&openapi3.Parameter{}).Validate(context.Background())
		require.EqualError(t, err, "parameter name can't be blank")

		var rfe *openapi3.RequiredFieldError
		require.True(t, errors.As(err, &rfe))
		require.Equal(t, "parameter.name", rfe.Field)

		var leaf *openapi3.ParameterNameRequired
		require.True(t, errors.As(err, &leaf))
	})

	t.Run("apiKey securityScheme name required", func(t *testing.T) {
		ss := &openapi3.SecurityScheme{Type: "apiKey", In: "header"}
		err := ss.Validate(context.Background())
		require.EqualError(t, err, "security scheme of type 'apiKey' should have 'name'")

		var rfe *openapi3.RequiredFieldError
		require.True(t, errors.As(err, &rfe))
		require.Equal(t, "securityScheme.name", rfe.Field)

		var leaf *openapi3.APIKeySecuritySchemeNameRequired
		require.True(t, errors.As(err, &leaf))
	})
}

// Pin ServerURLTemplateError cluster + leaf reachability for the three
// server URL template sites (mismatched braces, undeclared variables
// in two flavours).
func TestValidationError_ServerURLTemplateLeaves(t *testing.T) {
	t.Run("mismatched braces", func(t *testing.T) {
		s := &openapi3.Server{URL: "https://example.com/{x"}
		err := s.Validate(context.Background())
		require.EqualError(t, err, "server URL has mismatched { and }")

		var sue *openapi3.ServerURLTemplateError
		require.True(t, errors.As(err, &sue))
		require.Equal(t, "https://example.com/{x", sue.URL)

		var leaf *openapi3.ServerURLMismatchedBraces
		require.True(t, errors.As(err, &leaf))

		var ve *openapi3.ValidationError
		require.True(t, errors.As(err, &ve))
	})

	t.Run("undeclared variables (count mismatch)", func(t *testing.T) {
		s := &openapi3.Server{URL: "https://example.com/{x}"} // no Variables declared
		err := s.Validate(context.Background())
		require.EqualError(t, err, "server has undeclared variables")

		var sue *openapi3.ServerURLTemplateError
		require.True(t, errors.As(err, &sue))

		var leaf *openapi3.ServerURLUndeclaredVariables
		require.True(t, errors.As(err, &leaf))
	})

	t.Run("undeclared variables (name mismatch)", func(t *testing.T) {
		s := &openapi3.Server{
			URL:       "https://example.com/{x}",
			Variables: map[string]*openapi3.ServerVariable{"y": {Default: "z"}},
		}
		err := s.Validate(context.Background())
		require.EqualError(t, err, "server has undeclared variables")

		var leaf *openapi3.ServerURLUndeclaredVariables
		require.True(t, errors.As(err, &leaf))
	})
}

// Pin EitherFieldRequiredError cluster + leaf reachability for the
// two "at least one of these fields must be set" sites.
func TestValidationError_EitherFieldRequiredLeaves(t *testing.T) {
	t.Run("example value or externalValue", func(t *testing.T) {
		ex := &openapi3.Example{}
		err := ex.Validate(context.Background())
		require.EqualError(t, err, "no value or externalValue field")

		var efr *openapi3.EitherFieldRequiredError
		require.True(t, errors.As(err, &efr))
		require.Equal(t, []string{"value", "externalValue"}, efr.Fields)

		var leaf *openapi3.ExampleValueOrExternalValueRequired
		require.True(t, errors.As(err, &leaf))

		var ve *openapi3.ValidationError
		require.True(t, errors.As(err, &ve))
	})

	t.Run("link operationId or operationRef", func(t *testing.T) {
		link := &openapi3.Link{}
		err := link.Validate(context.Background())
		require.EqualError(t, err, "missing operationId or operationRef on link")

		var efr *openapi3.EitherFieldRequiredError
		require.True(t, errors.As(err, &efr))
		require.Equal(t, []string{"operationId", "operationRef"}, efr.Fields)

		var leaf *openapi3.LinkOperationIDOrRefRequired
		require.True(t, errors.As(err, &leaf))
	})
}

// Pin SchemaItemsRequired leaf reachability via the existing
// RequiredFieldError cluster.
func TestValidationError_SchemaItemsRequiredLeaf(t *testing.T) {
	schema := &openapi3.Schema{Type: &openapi3.Types{"array"}}
	err := schema.Validate(context.Background())
	require.EqualError(t, err, "when schema type is 'array', schema 'items' must be non-null")

	var rfe *openapi3.RequiredFieldError
	require.True(t, errors.As(err, &rfe))
	require.Equal(t, "schema.items", rfe.Field)

	var leaf *openapi3.SchemaItemsRequired
	require.True(t, errors.As(err, &leaf))
}

// Pin doc-root RequiredFieldError leaves: info, paths,
// jsonSchemaDialect-must-be-absolute-URI.
func TestValidationError_DocRootRequiredLeaves(t *testing.T) {
	t.Run("info required", func(t *testing.T) {
		// doc with no Info — fails with the wrap "invalid info: must be an object".
		doc := &openapi3.T{OpenAPI: "3.0.3", Paths: openapi3.NewPaths()}
		err := doc.Validate(context.Background())
		require.EqualError(t, err, "invalid info: must be an object")

		var rfe *openapi3.RequiredFieldError
		require.True(t, errors.As(err, &rfe))
		require.Equal(t, "info", rfe.Field)

		var leaf *openapi3.InfoRequired
		require.True(t, errors.As(err, &leaf))
	})

	t.Run("paths required (3.0)", func(t *testing.T) {
		// 3.0 doc with no Paths — fails with "invalid paths: must be an object".
		doc := &openapi3.T{
			OpenAPI: "3.0.3",
			Info:    &openapi3.Info{Title: "x", Version: "1.0.0"},
		}
		err := doc.Validate(context.Background())
		require.EqualError(t, err, "invalid paths: must be an object")

		var rfe *openapi3.RequiredFieldError
		require.True(t, errors.As(err, &rfe))
		require.Equal(t, "paths", rfe.Field)

		var leaf *openapi3.PathsRequired
		require.True(t, errors.As(err, &leaf))
	})

	t.Run("jsonSchemaDialect must be absolute URI", func(t *testing.T) {
		doc := &openapi3.T{
			OpenAPI:           "3.1.0",
			Info:              &openapi3.Info{Title: "x", Version: "1.0.0"},
			Paths:             openapi3.NewPaths(),
			JSONSchemaDialect: "no-scheme/relative",
		}
		err := doc.Validate(context.Background(), openapi3.IsOpenAPI31OrLater())
		require.EqualError(t, err, "invalid jsonSchemaDialect: must be an absolute URI with a scheme")

		var rfe *openapi3.RequiredFieldError
		require.True(t, errors.As(err, &rfe))
		require.Equal(t, "jsonSchemaDialect", rfe.Field)

		var leaf *openapi3.JSONSchemaDialectAbsoluteURIRequired
		require.True(t, errors.As(err, &leaf))
	})
}

// Pin SchemaBothFormsExclusive cluster + leaf reachability for the
// three union-typed schema fields set to both boolean and schema forms.
func TestValidationError_SchemaBothFormsLeaves(t *testing.T) {
	t.Run("additionalProperties both forms", func(t *testing.T) {
		yes := true
		schema := &openapi3.Schema{
			Type: &openapi3.Types{"object"},
			AdditionalProperties: openapi3.AdditionalProperties{
				Has:    &yes,
				Schema: &openapi3.SchemaRef{Value: &openapi3.Schema{}},
			},
		}
		err := schema.Validate(context.Background())
		require.EqualError(t, err, "additionalProperties are set to both boolean and schema")

		var sbf *openapi3.SchemaBothFormsExclusive
		require.True(t, errors.As(err, &sbf))
		require.Equal(t, "additionalProperties", sbf.Field)

		var leaf *openapi3.SchemaAdditionalPropertiesBothForms
		require.True(t, errors.As(err, &leaf))

		var ve *openapi3.ValidationError
		require.True(t, errors.As(err, &ve))
	})

	t.Run("unevaluatedItems both forms", func(t *testing.T) {
		yes := true
		schema := &openapi3.Schema{
			Type: &openapi3.Types{"object"},
			UnevaluatedItems: openapi3.BoolSchema{
				Has:    &yes,
				Schema: &openapi3.SchemaRef{Value: &openapi3.Schema{}},
			},
		}
		err := schema.Validate(context.Background(), openapi3.IsOpenAPI31OrLater())
		require.EqualError(t, err, "unevaluatedItems is set to both boolean and schema")

		var sbf *openapi3.SchemaBothFormsExclusive
		require.True(t, errors.As(err, &sbf))
		require.Equal(t, "unevaluatedItems", sbf.Field)

		var leaf *openapi3.SchemaUnevaluatedItemsBothForms
		require.True(t, errors.As(err, &leaf))
	})

	t.Run("unevaluatedProperties both forms", func(t *testing.T) {
		yes := true
		schema := &openapi3.Schema{
			Type: &openapi3.Types{"object"},
			UnevaluatedProperties: openapi3.BoolSchema{
				Has:    &yes,
				Schema: &openapi3.SchemaRef{Value: &openapi3.Schema{}},
			},
		}
		err := schema.Validate(context.Background(), openapi3.IsOpenAPI31OrLater())
		require.EqualError(t, err, "unevaluatedProperties is set to both boolean and schema")

		var sbf *openapi3.SchemaBothFormsExclusive
		require.True(t, errors.As(err, &sbf))
		require.Equal(t, "unevaluatedProperties", sbf.Field)

		var leaf *openapi3.SchemaUnevaluatedPropertiesBothForms
		require.True(t, errors.As(err, &leaf))
	})
}

// Pin ExactlyOneFieldError and SingleEntryContentError clusters for
// the four parameter/header content+schema sites.
func TestValidationError_ParameterHeaderContentSchemaLeaves(t *testing.T) {
	t.Run("parameter content/schema exactly one (neither set)", func(t *testing.T) {
		p := &openapi3.Parameter{Name: "p", In: "query"}
		err := p.Validate(context.Background())
		require.ErrorContains(t, err, "parameter must contain exactly one of content and schema")

		var efe *openapi3.ExactlyOneFieldError
		require.True(t, errors.As(err, &efe))
		require.Equal(t, []string{"content", "schema"}, efe.Fields)

		var leaf *openapi3.ParameterContentSchemaExactlyOne
		require.True(t, errors.As(err, &leaf))
	})

	t.Run("parameter content single entry", func(t *testing.T) {
		p := &openapi3.Parameter{
			Name: "p", In: "query",
			Content: openapi3.Content{
				"application/json": &openapi3.MediaType{},
				"application/xml":  &openapi3.MediaType{},
			},
		}
		err := p.Validate(context.Background())
		require.ErrorContains(t, err, "parameter content must only contain one entry")

		var sec *openapi3.SingleEntryContentError
		require.True(t, errors.As(err, &sec))
		require.Equal(t, "parameter", sec.Subject)

		var leaf *openapi3.ParameterContentSingleEntry
		require.True(t, errors.As(err, &leaf))
	})

	t.Run("header content/schema exactly one (neither set)", func(t *testing.T) {
		h := &openapi3.Header{}
		err := h.Validate(context.Background())
		require.ErrorContains(t, err, "parameter must contain exactly one of content and schema")

		var efe *openapi3.ExactlyOneFieldError
		require.True(t, errors.As(err, &efe))
		require.Equal(t, []string{"content", "schema"}, efe.Fields)

		var leaf *openapi3.HeaderContentSchemaExactlyOne
		require.True(t, errors.As(err, &leaf))
	})

	t.Run("header content single entry", func(t *testing.T) {
		h := &openapi3.Header{
			Parameter: openapi3.Parameter{
				Content: openapi3.Content{
					"application/json": &openapi3.MediaType{},
					"application/xml":  &openapi3.MediaType{},
				},
			},
		}
		err := h.Validate(context.Background())
		require.ErrorContains(t, err, "parameter content must only contain one entry")

		var sec *openapi3.SingleEntryContentError
		require.True(t, errors.As(err, &sec))
		require.Equal(t, "header", sec.Subject)

		var leaf *openapi3.HeaderContentSingleEntry
		require.True(t, errors.As(err, &leaf))
	})
}

// Pin WebhookNilError cluster + leaf reachability for the
// nil-pathitem webhook check in T.Validate.
func TestValidationError_WebhookNilLeaf(t *testing.T) {
	doc := &openapi3.T{
		OpenAPI:  "3.1.0",
		Info:     &openapi3.Info{Title: "x", Version: "1.0.0"},
		Paths:    openapi3.NewPaths(),
		Webhooks: map[string]*openapi3.PathItem{"onEvent": nil},
	}
	err := doc.Validate(context.Background(), openapi3.IsOpenAPI31OrLater())
	require.EqualError(t, err, `invalid webhooks: webhook "onEvent" is nil`)

	var wne *openapi3.WebhookNilError
	require.True(t, errors.As(err, &wne))
	require.Equal(t, "onEvent", wne.Name)

	var leaf *openapi3.WebhookNil
	require.True(t, errors.As(err, &leaf))

	var ve *openapi3.ValidationError
	require.True(t, errors.As(err, &ve))
}

func TestValidationError_PathParameterRequired(t *testing.T) {
	// Path parameters must be declared required: true. A parameter with
	// in: path and required: false (or omitted) triggers the cluster.
	doc := loadDocFromYAML(t, `
openapi: 3.0.3
info: { title: t, version: "1" }
paths:
  /things/{id}:
    get:
      parameters:
        - { name: id, in: path }
      responses: { "200": { description: ok } }
`)
	err := doc.Validate(context.Background())
	require.ErrorContains(t, err, `path parameter "id" must be required`)

	var ppr *openapi3.PathParameterRequiredError
	require.True(t, errors.As(err, &ppr))
	require.Equal(t, "id", ppr.Param)
}

func TestValidationError_DuplicateOperationID(t *testing.T) {
	// Two operations sharing the same operationId across paths must
	// surface a DuplicateOperationIDError carrying both endpoints.
	doc := loadDocFromYAML(t, `
openapi: 3.0.3
info: { title: t, version: "1" }
paths:
  /a:
    get:
      operationId: shared
      responses: { "200": { description: ok } }
  /b:
    get:
      operationId: shared
      responses: { "200": { description: ok } }
`)
	err := doc.Validate(context.Background())
	require.ErrorContains(t, err, `operations "GET /a" and "GET /b" have the same operation id "shared"`)

	var doe *openapi3.DuplicateOperationIDError
	require.True(t, errors.As(err, &doe))
	require.Equal(t, "shared", doe.OperationID)
	require.Equal(t, "GET /a", doe.Endpoint1)
	require.Equal(t, "GET /b", doe.Endpoint2)
}

func TestValidationError_ExtraSiblingFields(t *testing.T) {
	// A non-x- key in Extensions triggers validateExtensions's
	// "extra sibling fields" error, now typed as ExtraSiblingFieldsError.
	// Construct a non-empty Responses so the empty-responses guard
	// doesn't fire first; the only finding then comes from extensions.
	responses := openapi3.NewResponses(
		openapi3.WithStatus(200, &openapi3.ResponseRef{
			Value: openapi3.NewResponse().WithDescription("ok"),
		}),
	)
	responses.Extensions = map[string]any{"bogus": "value"}
	err := responses.Validate(context.Background())
	require.ErrorContains(t, err, "extra sibling fields: [bogus]")

	var esf *openapi3.ExtraSiblingFieldsError
	require.True(t, errors.As(err, &esf))
	require.Equal(t, []string{"bogus"}, esf.Fields)
}

func TestValidationError_SchemaTypeError(t *testing.T) {
	// Unsupported 'type' value on a schema (e.g., "bool" instead of
	// "boolean") triggers SchemaTypeError carrying the bad value.
	doc := loadDocFromYAML(t, `
openapi: 3.0.3
info: { title: t, version: "1" }
paths: {}
components:
  schemas:
    Bad: { type: bool }
`)
	err := doc.Validate(context.Background())
	require.ErrorContains(t, err, `unsupported 'type' value "bool"`)

	var ste *openapi3.SchemaTypeError
	require.True(t, errors.As(err, &ste))
	require.Equal(t, "bool", ste.Type)
}

// Origin tracking for DuplicateOperationIDError. When IncludeOrigin is
// set, the cluster carries the offending (second) operation's Origin so
// consumers can pin the finding at the duplicate operationId rather
// than at the document root.
func TestValidationError_DuplicateOperationID_CarriesOrigin(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IncludeOrigin = true
	doc, err := loader.LoadFromData([]byte(`
openapi: 3.0.3
info: { title: t, version: "1" }
paths:
  /a:
    get:
      operationId: shared
      responses: { "200": { description: ok } }
  /b:
    get:
      operationId: shared
      responses: { "200": { description: ok } }
`))
	require.NoError(t, err)

	verr := doc.Validate(context.Background())
	var doe *openapi3.DuplicateOperationIDError
	require.True(t, errors.As(verr, &doe))
	require.NotNil(t, doe.Origin, "cluster should carry the offending operation's Origin when loader tracks origins")
	require.NotNil(t, doe.Origin.Key, "Origin.Key set by the loader")
	require.Greater(t, doe.Origin.Key.Line, 0)
}

// Without IncludeOrigin, DuplicateOperationIDError.Origin is nil — no
// fabrication of location info that wasn't tracked.
func TestValidationError_DuplicateOperationID_OriginNilWithoutLoaderTracking(t *testing.T) {
	doc := loadDocFromYAML(t, `
openapi: 3.0.3
info: { title: t, version: "1" }
paths:
  /a:
    get:
      operationId: shared
      responses: { "200": { description: ok } }
  /b:
    get:
      operationId: shared
      responses: { "200": { description: ok } }
`)
	verr := doc.Validate(context.Background())
	var doe *openapi3.DuplicateOperationIDError
	require.True(t, errors.As(verr, &doe))
	require.Nil(t, doe.Origin, "Origin should be nil when loader didn't track origins")
}

// Origin tracking for ExtraSiblingFieldsError. The cluster carries the
// parent object's Origin so consumers can pin the finding at the
// container holding the unexpected sibling fields. Exercised here via
// a $ref with a disallowed sibling, which is the most common surface.
func TestValidationError_ExtraSiblingFields_CarriesOrigin(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IncludeOrigin = true
	doc, err := loader.LoadFromData([]byte(`
openapi: 3.0.3
info: { title: t, version: "1" }
paths:
  /x:
    get:
      responses:
        "200":
          description: ok
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/T"
                description: should-not-be-here
components:
  schemas:
    T: { type: string }
`))
	require.NoError(t, err)

	verr := doc.Validate(context.Background())
	var esf *openapi3.ExtraSiblingFieldsError
	require.True(t, errors.As(verr, &esf))
	require.NotNil(t, esf.Origin, "cluster should carry the parent object's Origin when loader tracks origins")
	require.NotNil(t, esf.Origin.Key)
	require.Greater(t, esf.Origin.Key.Line, 0)
}

// A parameter with `in:` set to anything outside {path, query, header,
// cookie} triggers InvalidParameterInError carrying the rejected
// value. Most common offender is `in: body` from Swagger 2.0 specs
// that didn't fully migrate to 3.x.
func TestValidationError_InvalidParameterIn(t *testing.T) {
	doc := loadDocFromYAML(t, `
openapi: 3.0.3
info: { title: t, version: "1" }
paths:
  /x:
    post:
      parameters:
        - { name: payload, in: body, schema: { type: object } }
      responses: { "200": { description: ok } }
`)
	err := doc.Validate(context.Background())
	require.ErrorContains(t, err, `parameter can't have 'in' value "body"`)

	var ipe *openapi3.InvalidParameterInError
	require.True(t, errors.As(err, &ipe))
	require.Equal(t, "body", ipe.Value)
}

// Origin tracking for InvalidParameterInError.
func TestValidationError_InvalidParameterIn_CarriesOrigin(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IncludeOrigin = true
	doc, err := loader.LoadFromData([]byte(`
openapi: 3.0.3
info: { title: t, version: "1" }
paths:
  /x:
    post:
      parameters:
        - { name: payload, in: body, schema: { type: object } }
      responses: { "200": { description: ok } }
`))
	require.NoError(t, err)

	verr := doc.Validate(context.Background())
	var ipe *openapi3.InvalidParameterInError
	require.True(t, errors.As(verr, &ipe))
	require.NotNil(t, ipe.Origin)
	require.NotNil(t, ipe.Origin.Key)
	require.Greater(t, ipe.Origin.Key.Line, 0)
}

// A schema `pattern:` using a Perl-only regex feature (lookahead /
// lookbehind etc.) fails to compile against Go's RE2 and triggers
// SchemaPatternRegexError. The cluster carries the offending pattern
// AND chains through to the original *SchemaError via Unwrap so
// callers using errors.As against the legacy *SchemaError still match.
func TestValidationError_SchemaPatternRegex(t *testing.T) {
	doc := loadDocFromYAML(t, `
openapi: 3.0.3
info: { title: t, version: "1" }
paths: {}
components:
  schemas:
    Bad:
      type: string
      pattern: "(?!foo)bar"
`)
	err := doc.Validate(context.Background())
	require.Error(t, err)

	var spre *openapi3.SchemaPatternRegexError
	require.True(t, errors.As(err, &spre))
	require.Equal(t, "(?!foo)bar", spre.Pattern)

	// Backward compat: Unwrap reaches the legacy *SchemaError.
	var se *openapi3.SchemaError
	require.True(t, errors.As(err, &se), "*SchemaError must still be reachable via Unwrap chain")
	require.Equal(t, "pattern", se.SchemaField)
}

// Origin tracking for SchemaPatternRegexError.
func TestValidationError_SchemaPatternRegex_CarriesOrigin(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IncludeOrigin = true
	doc, err := loader.LoadFromData([]byte(`
openapi: 3.0.3
info: { title: t, version: "1" }
paths: {}
components:
  schemas:
    Bad:
      type: string
      pattern: "(?!foo)bar"
`))
	require.NoError(t, err)

	verr := doc.Validate(context.Background())
	var spre *openapi3.SchemaPatternRegexError
	require.True(t, errors.As(verr, &spre))
	require.NotNil(t, spre.Origin)
	require.NotNil(t, spre.Origin.Key)
	require.Greater(t, spre.Origin.Key.Line, 0)
}

// Security scheme with a `type:` outside the spec-permitted set
// {apiKey, http, oauth2, openIdConnect, mutualTLS} triggers
// InvalidSecuritySchemeTypeError carrying the rejected value.
func TestValidationError_InvalidSecuritySchemeType(t *testing.T) {
	doc := loadDocFromYAML(t, `
openapi: 3.0.3
info: { title: t, version: "1" }
paths: {}
components:
  securitySchemes:
    Bad:
      type: cookie
`)
	err := doc.Validate(context.Background())
	require.ErrorContains(t, err, `security scheme 'type' can't be "cookie"`)

	var iste *openapi3.InvalidSecuritySchemeTypeError
	require.True(t, errors.As(err, &iste))
	require.Equal(t, "cookie", iste.Type)
}

// Origin tracking for InvalidSecuritySchemeTypeError.
func TestValidationError_InvalidSecuritySchemeType_CarriesOrigin(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IncludeOrigin = true
	doc, err := loader.LoadFromData([]byte(`
openapi: 3.0.3
info: { title: t, version: "1" }
paths: {}
components:
  securitySchemes:
    Bad:
      type: cookie
`))
	require.NoError(t, err)

	verr := doc.Validate(context.Background())
	var iste *openapi3.InvalidSecuritySchemeTypeError
	require.True(t, errors.As(verr, &iste))
	require.NotNil(t, iste.Origin)
	require.NotNil(t, iste.Origin.Key)
	require.Greater(t, iste.Origin.Key.Line, 0)
}

// HTTP security scheme with a `scheme:` outside {bearer, basic,
// negotiate, digest} triggers InvalidHTTPSchemeError carrying the
// rejected value.
func TestValidationError_InvalidHTTPScheme(t *testing.T) {
	doc := loadDocFromYAML(t, `
openapi: 3.0.3
info: { title: t, version: "1" }
paths: {}
components:
  securitySchemes:
    Bad:
      type: http
      scheme: mutual
`)
	err := doc.Validate(context.Background())
	require.ErrorContains(t, err, `security scheme of type 'http' has invalid 'scheme' value "mutual"`)

	var ihse *openapi3.InvalidHTTPSchemeError
	require.True(t, errors.As(err, &ihse))
	require.Equal(t, "mutual", ihse.Scheme)
}

// Origin tracking for InvalidHTTPSchemeError.
func TestValidationError_InvalidHTTPScheme_CarriesOrigin(t *testing.T) {
	loader := openapi3.NewLoader()
	loader.IncludeOrigin = true
	doc, err := loader.LoadFromData([]byte(`
openapi: 3.0.3
info: { title: t, version: "1" }
paths: {}
components:
  securitySchemes:
    Bad:
      type: http
      scheme: mutual
`))
	require.NoError(t, err)

	verr := doc.Validate(context.Background())
	var ihse *openapi3.InvalidHTTPSchemeError
	require.True(t, errors.As(verr, &ihse))
	require.NotNil(t, ihse.Origin)
	require.NotNil(t, ihse.Origin.Key)
	require.Greater(t, ihse.Origin.Key.Line, 0)
}

// A $ref left with a non-nil Ref string but nil Value at Validate time
// triggers UnresolvedRefError. Constructed programmatically because
// the YAML loader is strict about ref resolution at load time; this
// shape models the in-the-wild case where a spec uses an external
// $ref that wasn't fetched (testdata/apis_guru_openapi_directory has
// real examples).
func TestValidationError_UnresolvedRef(t *testing.T) {
	doc := &openapi3.T{
		OpenAPI: "3.0.3",
		Info:    &openapi3.Info{Title: "t", Version: "1"},
		Paths:   openapi3.NewPaths(),
		Components: &openapi3.Components{
			Schemas: openapi3.Schemas{
				"X": &openapi3.SchemaRef{
					Ref:   "external.yaml#/T",
					Value: nil, // unresolved
				},
			},
		},
	}
	err := doc.Validate(context.Background())
	require.ErrorContains(t, err, `found unresolved ref: "external.yaml#/T"`)

	var ure *openapi3.UnresolvedRefError
	require.True(t, errors.As(err, &ure))
	require.Equal(t, "external.yaml#/T", ure.Ref)
}

// Without IncludeOrigin, ExtraSiblingFieldsError.Origin is nil.
func TestValidationError_ExtraSiblingFields_OriginNilWithoutLoaderTracking(t *testing.T) {
	responses := openapi3.NewResponses(
		openapi3.WithStatus(200, &openapi3.ResponseRef{
			Value: openapi3.NewResponse().WithDescription("ok"),
		}),
	)
	responses.Extensions = map[string]any{"bogus": "value"}
	verr := responses.Validate(context.Background())
	var esf *openapi3.ExtraSiblingFieldsError
	require.True(t, errors.As(verr, &esf))
	require.Nil(t, esf.Origin, "Origin should be nil when the parent object's Origin is unset")
}

func loadDocFromYAML(t *testing.T, src string) *openapi3.T {
	t.Helper()
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(src))
	require.NoError(t, err)
	return doc
}
