package openapi3_test

import (
	"regexp"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

// codedErrorInventory holds one instance of every typed validation error.
// Codes are literals on the types, so zero values suffice. When a validation
// is added, renamed, or removed, update this inventory and
// ValidationErrorCodes together; TestValidationErrorCodes fails until the two
// agree, which keeps the exported catalog honest.
func codedErrorInventory() []openapi3.CodedError {
	return []openapi3.CodedError{
		&openapi3.AnchorFieldFor31Plus{},
		&openapi3.APIKeyInInvalidError{},
		&openapi3.APIKeySecuritySchemeNameRequired{},
		&openapi3.CommentFieldFor31Plus{},
		&openapi3.ConflictingPathsError{},
		&openapi3.ConstFieldFor31Plus{},
		&openapi3.ContainsFieldFor31Plus{},
		&openapi3.ContentEncodingFieldFor31Plus{},
		&openapi3.ContentMediaTypeFieldFor31Plus{},
		&openapi3.ContentSchemaFieldFor31Plus{},
		&openapi3.DefaultViolatesSchema{},
		&openapi3.DefsFieldFor31Plus{},
		&openapi3.DependentRequiredFieldFor31Plus{},
		&openapi3.DependentSchemasFieldFor31Plus{},
		&openapi3.DuplicateOperationIDError{},
		&openapi3.DuplicateParameterError{},
		&openapi3.DuplicateRequiredFieldError{},
		&openapi3.DuplicateTagError{},
		&openapi3.DynamicAnchorFieldFor31Plus{},
		&openapi3.DynamicRefFieldFor31Plus{},
		&openapi3.ElseFieldFor31Plus{},
		&openapi3.ExamplesFieldFor31Plus{},
		&openapi3.ExampleValueExternalValueExclusive{},
		&openapi3.ExampleValueOrExternalValueRequired{},
		&openapi3.ExampleViolatesSchema{},
		&openapi3.ExternalDocsURLRequired{},
		&openapi3.ExtraSiblingFieldsError{},
		&openapi3.HeaderContentSchemaExactlyOne{},
		&openapi3.HeaderContentSingleEntry{},
		&openapi3.HeaderInForbidden{},
		&openapi3.HeaderNameForbidden{},
		&openapi3.IDFieldFor31Plus{},
		&openapi3.IfFieldFor31Plus{},
		&openapi3.InfoRequired{},
		&openapi3.InfoSummaryFieldFor31Plus{},
		&openapi3.InfoTitleRequired{},
		&openapi3.InfoVersionRequired{},
		&openapi3.InvalidHTTPSchemeError{},
		&openapi3.InvalidParameterInError{},
		&openapi3.InvalidSecuritySchemeTypeError{},
		&openapi3.InvalidSerializationMethodError{},
		&openapi3.ItemSchemaFieldFor32Plus{},
		&openapi3.JSONSchemaDialectAbsoluteURIRequired{},
		&openapi3.JSONSchemaDialectFieldFor31Plus{},
		&openapi3.LicenseIdentifierFieldFor31Plus{},
		&openapi3.LicenseNameRequired{},
		&openapi3.LicenseURLIdentifierExclusive{},
		&openapi3.LinkOperationIDOrRefRequired{},
		&openapi3.LinkOperationIDRefExclusive{},
		&openapi3.MaxContainsFieldFor31Plus{},
		&openapi3.MediaTypeExampleExamplesExclusive{},
		&openapi3.MinContainsFieldFor31Plus{},
		&openapi3.OAuthFlowAuthorizationURLForbidden{},
		&openapi3.OAuthFlowAuthorizationURLRequired{},
		&openapi3.OAuthFlowScopesRequired{},
		&openapi3.OAuthFlowTokenURLForbidden{},
		&openapi3.OAuthFlowTokenURLRequired{},
		&openapi3.OpenAPIVersionRequired{},
		&openapi3.OpenIDConnectURLRequired{},
		&openapi3.OperationResponsesRequired{},
		&openapi3.ParameterContentSchemaExactlyOne{},
		&openapi3.ParameterContentSingleEntry{},
		&openapi3.ParameterExampleAndExamplesExclusive{},
		&openapi3.ParameterNameRequired{},
		&openapi3.PathMustStartWithSlashError{},
		&openapi3.PathParameterRequiredError{},
		&openapi3.PathParametersError{},
		&openapi3.PathsRequired{},
		&openapi3.PatternPropertiesFieldFor31Plus{},
		&openapi3.PrefixItemsFieldFor31Plus{},
		&openapi3.PropertyNamesFieldFor31Plus{},
		&openapi3.RequestBodyContentRequired{},
		&openapi3.ResponseDescriptionRequired{},
		&openapi3.ResponsesNonEmptyRequired{},
		&openapi3.SchemaAdditionalPropertiesBothForms{},
		&openapi3.SchemaFieldFor31Plus{},
		&openapi3.SchemaItemsRequired{},
		&openapi3.SchemaPatternRegexError{},
		&openapi3.SchemaReadOnlyWriteOnlyExclusive{},
		&openapi3.SchemaTypeError{},
		&openapi3.SchemaUnevaluatedItemsBothForms{},
		&openapi3.SchemaUnevaluatedPropertiesBothForms{},
		&openapi3.SecuritySchemeBearerFormatForbidden{},
		&openapi3.SecuritySchemeFlowsForbidden{},
		&openapi3.SecuritySchemeFlowsRequired{},
		&openapi3.SecuritySchemeInForbidden{},
		&openapi3.SecuritySchemeNameForbidden{},
		&openapi3.ServerURLRequired{},
		&openapi3.ServerURLTemplateError{},
		&openapi3.ServerVariableDefaultRequired{},
		&openapi3.ThenFieldFor31Plus{},
		&openapi3.UnevaluatedItemsFieldFor31Plus{},
		&openapi3.UnevaluatedPropertiesFieldFor31Plus{},
		&openapi3.UnresolvedRefError{},
		&openapi3.WebhookNilError{},
		&openapi3.WebhooksFieldFor31Plus{},
	}
}

// TestValidationErrorCodes pins the code contract: every typed error yields a
// code from the exported catalog, every catalog entry is yielded by some
// type, and the catalog is sorted, unique, kebab-case.
func TestValidationErrorCodes(t *testing.T) {
	catalog := openapi3.ValidationErrorCodes()
	require.True(t, slices.IsSorted(catalog), "catalog must be sorted")
	require.Equal(t, len(slices.Compact(slices.Clone(catalog))), len(catalog), "catalog must be unique")
	kebab := regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	for _, c := range catalog {
		require.Truef(t, kebab.MatchString(c), "code %q is not kebab-case", c)
	}

	emitted := map[string]struct{}{}
	for _, e := range codedErrorInventory() {
		emitted[e.Code()] = struct{}{}
	}
	emittedSorted := make([]string, 0, len(emitted))
	for c := range emitted {
		emittedSorted = append(emittedSorted, c)
	}
	slices.Sort(emittedSorted)
	require.Equal(t, catalog, emittedSorted,
		"catalog and inventory diverge: update ValidationErrorCodes and codedErrorInventory together")
}

// TestValidationErrorCodes_EndToEnd proves codes are reachable with errors.As
// through every wrapper shape Validate produces: the inventory test covers
// which codes exist; this covers that a consumer can get to them. One row per
// unwrap path, not per code.
func TestValidationErrorCodes_EndToEnd(t *testing.T) {
	for _, tc := range []struct {
		name    string // the wrapper chain under test
		spec    string
		expects string
	}{
		{
			name: "doc root, unwrapped",
			spec: `
openapi: ""
info: { title: t, version: "1" }
paths: {}
`,
			expects: "openapi-required",
		},
		{
			name: "path and operation wrappers",
			spec: `
openapi: 3.0.0
info: { title: t, version: "1" }
paths:
  /a:
    get:
      summary: no responses
`,
			expects: "operation-responses-required",
		},
		{
			name: "section wrapper inside an operation",
			spec: `
openapi: 3.0.0
info: { title: t, version: "1" }
paths:
  /a:
    get:
      externalDocs: { description: d }
      responses:
        "200": { description: ok }
`,
			expects: "external-docs-url-required",
		},
		{
			name: "component wrapper",
			spec: `
openapi: 3.0.0
info: { title: t, version: "1" }
paths: {}
components:
  securitySchemes:
    s: { type: bogus }
`,
			expects: "security-scheme-type-invalid",
		},
		{
			name: "schema nested in a component",
			spec: `
openapi: 3.0.0
info: { title: t, version: "1" }
paths: {}
components:
  schemas:
    S: { type: string, const: x }
`,
			expects: "const-field-for-3-1-plus",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			loader := openapi3.NewLoader()
			doc, err := loader.LoadFromData([]byte(tc.spec))
			require.NoError(t, err)

			verr := doc.Validate(t.Context(), openapi3.EnableMultiError())
			require.Error(t, verr)
			var multi openapi3.MultiError
			require.ErrorAs(t, verr, &multi)
			require.NotEmpty(t, multi)

			var codes []string
			for _, e := range multi {
				var coded openapi3.CodedError
				require.ErrorAs(t, e, &coded, "a code must be reachable through the %s chain (error: %v)", tc.name, e)
				codes = append(codes, coded.Code())
			}
			require.Contains(t, codes, tc.expects)
			for _, c := range codes {
				require.Contains(t, openapi3.ValidationErrorCodes(), c, "emitted codes come from the catalog")
			}
		})
	}
}
