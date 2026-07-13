package openapi3_test

import (
	"regexp"
	"slices"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

// codedErrorInventory holds one instance of every validation error the
// openapi3 validators emit, with the strings the emit sites use. When a
// validation is added, renamed, or removed, update both this inventory and
// ValidationErrorCodes -- TestValidationErrorCodes fails until the two agree,
// which keeps the exported catalog honest.
func codedErrorInventory() []openapi3.CodedError {
	var errs []openapi3.CodedError

	for _, f := range []string{
		"default", "externalDocs.url", "flows", "info", "info.title", "info.version",
		"jsonSchemaDialect", "license.name", "oAuthFlow.authorizationUrl", "oAuthFlow.scopes",
		"oAuthFlow.tokenUrl", "openapi", "openIdConnectUrl", "operation.responses",
		"parameter.name", "paths", "requestBody.content", "response.description",
		"responses", "schema.items", "securityScheme.name", "server.url",
	} {
		errs = append(errs, &openapi3.RequiredFieldError{Field: f})
	}

	// fields allowed from OAS 3.1 (see Schema.Validate and the per-object
	// validators) and from OAS 3.2 (itemSchema)
	for _, f := range []string{
		"summary", "identifier", "webhooks", "jsonschemadialect",
		"$anchor", "$comment", "$defs", "$dynamicAnchor", "$dynamicRef", "$id", "$schema",
		"const", "contains", "contentEncoding", "contentMediaType", "contentSchema",
		"dependentRequired", "dependentSchemas", "else", "examples", "if",
		"maxContains", "minContains", "patternProperties", "prefixItems",
		"propertyNames", "then", "unevaluatedItems", "unevaluatedProperties",
	} {
		errs = append(errs, &openapi3.FieldVersionMismatchError{Field: f, MinVersion: "3.1"})
	}
	errs = append(errs, &openapi3.FieldVersionMismatchError{Field: "itemSchema", MinVersion: "3.2"})

	for _, k := range []string{"example", "default"} {
		errs = append(errs, &openapi3.SchemaValueError{ValueKind: k})
	}

	for _, p := range [][2]string{
		{"value", "externalValue"}, {"example", "examples"}, {"url", "identifier"},
		{"operationId", "operationRef"}, {"readOnly", "writeOnly"},
	} {
		errs = append(errs, &openapi3.MutuallyExclusiveFieldsError{Field1: p[0], Field2: p[1]})
	}

	for _, f := range []string{"authorizationUrl", "bearerFormat", "flows", "in", "name", "tokenUrl"} {
		errs = append(errs, &openapi3.ForbiddenFieldError{Field: f})
	}

	for _, f := range []string{"additionalProperties", "unevaluatedItems", "unevaluatedProperties"} {
		errs = append(errs, &openapi3.SchemaBothFormsExclusive{Field: f})
	}

	for _, s := range []string{"header", "parameter"} {
		errs = append(errs, &openapi3.SingleEntryContentError{Subject: s})
	}

	return append(errs,
		&openapi3.EitherFieldRequiredError{Fields: []string{"value", "externalValue"}},
		&openapi3.EitherFieldRequiredError{Fields: []string{"operationId", "operationRef"}},
		&openapi3.ExactlyOneFieldError{Fields: []string{"content", "schema"}},
		&openapi3.DuplicateRequiredFieldError{},
		&openapi3.DuplicateTagError{},
		&openapi3.PathParametersError{},
		&openapi3.ServerURLTemplateError{},
		&openapi3.WebhookNilError{},
		&openapi3.PathParameterRequiredError{},
		&openapi3.DuplicateOperationIDError{},
		&openapi3.ExtraSiblingFieldsError{},
		&openapi3.SchemaTypeError{},
		&openapi3.InvalidParameterInError{},
		&openapi3.SchemaPatternRegexError{},
		&openapi3.InvalidSecuritySchemeTypeError{},
		&openapi3.InvalidHTTPSchemeError{},
		&openapi3.UnresolvedRefError{},
		&openapi3.APIKeyInInvalidError{},
		&openapi3.PathMustStartWithSlashError{},
		&openapi3.ConflictingPathsError{},
		&openapi3.DuplicateParameterError{},
		&openapi3.InvalidSerializationMethodError{},
	)
}

// TestValidationErrorCodes pins the code contract: every inventory error
// yields a code from the exported catalog, every catalog entry is yielded by
// some inventory error, and the catalog is sorted, unique, kebab-case.
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

// TestValidationErrorCodes_EndToEnd exercises the code surface through a real
// document validation.
func TestValidationErrorCodes_EndToEnd(t *testing.T) {
	const spec = `
openapi: 3.0.0
info: { title: t, version: "1" }
paths:
  /a:
    get:
      summary: no responses
`
	doc, err := openapi3.NewLoader().LoadFromData([]byte(spec))
	require.NoError(t, err)
	verr := doc.Validate(t.Context(), openapi3.EnableMultiError())
	require.Error(t, verr)

	var codes []string
	for _, e := range verr.(openapi3.MultiError) {
		var coded openapi3.CodedError
		require.ErrorAs(t, e, &coded, "every validation error carries a code")
		codes = append(codes, coded.Code())
	}
	require.Contains(t, codes, "operation-responses-required")
	for _, c := range codes {
		require.Contains(t, openapi3.ValidationErrorCodes(), c, "emitted codes come from the catalog")
	}
}
