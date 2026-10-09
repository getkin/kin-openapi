package openapi3_test

import (
	"errors"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func securityRequirementsDoc(t *testing.T) *openapi3.T {
	t.Helper()
	return loadDoc(t, `
openapi: 3.0.3
info: {title: Example, version: "1"}
paths:
  /pets:
    get:
      responses:
        "200": {description: ok}
components:
  securitySchemes:
    bearerAuth: {type: http, scheme: bearer}
`)
}

func securityRequirementsOperation() *openapi3.Operation {
	return &openapi3.Operation{
		Responses: openapi3.NewResponses(openapi3.WithName("200", openapi3.NewResponse().WithDescription("ok"))),
		Security:  &openapi3.SecurityRequirements{{"missing": {}}},
	}
}

func securityRequirementsCallback(op *openapi3.Operation) *openapi3.CallbackRef {
	return &openapi3.CallbackRef{Value: openapi3.NewCallback(
		openapi3.WithCallback("{$request.body#/url}", &openapi3.PathItem{Post: op}),
	)}
}

func TestIssue1082_UndeclaredSecuritySchemes(t *testing.T) {
	t.Parallel()

	for _, version := range []string{"3.0.3", "3.1.2"} {
		for _, tc := range []struct {
			name   string
			change func(*openapi3.T)
		}{
			{"root", func(doc *openapi3.T) {
				doc.Security = openapi3.SecurityRequirements{{"missing": {}}}
			}},
			{"operation", func(doc *openapi3.T) {
				doc.Paths.Value("/pets").Get = securityRequirementsOperation()
			}},
			{"callback", func(doc *openapi3.T) {
				doc.Paths.Value("/pets").Get.Callbacks = openapi3.Callbacks{"event": securityRequirementsCallback(securityRequirementsOperation())}
			}},
			{"nested callback", func(doc *openapi3.T) {
				op := securityRequirementsOperation()
				op.Security = nil
				op.Callbacks = openapi3.Callbacks{"nested": securityRequirementsCallback(securityRequirementsOperation())}
				doc.Paths.Value("/pets").Get.Callbacks = openapi3.Callbacks{"event": securityRequirementsCallback(op)}
			}},
			{"component callback", func(doc *openapi3.T) {
				doc.Components.Callbacks = openapi3.Callbacks{"event": securityRequirementsCallback(securityRequirementsOperation())}
			}},
		} {
			t.Run(version+"/"+tc.name, func(t *testing.T) {
				doc := securityRequirementsDoc(t)
				doc.OpenAPI = version
				tc.change(doc)
				err := doc.Validate(t.Context())
				require.Error(t, err)
				require.ErrorContains(t, err, "missing")
			})
		}
	}

	t.Run("webhook", func(t *testing.T) {
		doc := securityRequirementsDoc(t)
		doc.OpenAPI = "3.1.2"
		doc.Webhooks = map[string]*openapi3.PathItem{"event": {Post: securityRequirementsOperation()}}
		require.ErrorContains(t, doc.Validate(t.Context()), "missing")
	})
}

func TestIssue1082_Controls(t *testing.T) {
	t.Parallel()

	for _, security := range []openapi3.SecurityRequirements{
		nil, {}, {{}}, {{"bearerAuth": {}}}, {{"bearerAuth": {"role"}}},
	} {
		doc := securityRequirementsDoc(t)
		doc.Security = security
		doc.Paths.Value("/pets").Get.Security = &openapi3.SecurityRequirements{}
		require.NoError(t, doc.Validate(t.Context()))
	}
	for _, name := range []string{"https://example.com/security", "./scheme", "missing", "bearerAuth"} {
		t.Run("3.2/"+name, func(t *testing.T) {
			doc := securityRequirementsDoc(t)
			doc.OpenAPI = "3.2.0"
			doc.Security = openapi3.SecurityRequirements{{name: {}}}
			doc.Paths.Value("/pets").Get = securityRequirementsOperation()
			require.NoError(t, doc.Validate(t.Context()))
		})
	}
	standalone := openapi3.SecurityRequirement{"missing": {}}
	require.NoError(t, standalone.Validate(t.Context()))
	require.NoError(t, openapi3.SecurityRequirements{standalone}.Validate(t.Context()))
}

func TestIssue1082_ErrorMetadataAndMultiError(t *testing.T) {
	t.Parallel()

	doc := securityRequirementsDoc(t)
	doc.Security = openapi3.SecurityRequirements{{"z": {}, "a~/": {}, "bearerAuth": {}}, {"other": {}}}
	op := securityRequirementsOperation()
	doc.Paths.Value("/pets").Get = op

	var leaf *openapi3.SecurityRequirementSchemeUndefinedError
	err := doc.Validate(t.Context())
	require.ErrorAs(t, err, &leaf)
	require.Equal(t, "a~/", leaf.Scheme)
	require.Equal(t, "/security/0/a~0~1", leaf.JSONPointer)
	var base *openapi3.ValidationError
	require.ErrorAs(t, err, &base)
	require.Equal(t, leaf.Message, base.Message)
	var coded openapi3.CodedError
	require.ErrorAs(t, err, &coded)
	require.Equal(t, "security-requirement-scheme-undefined", coded.Code())
	var section *openapi3.SectionValidationError
	require.ErrorAs(t, err, &section)
	require.Equal(t, "security", section.Section)

	var multi openapi3.MultiError
	require.ErrorAs(t, doc.Validate(t.Context(), openapi3.EnableMultiError()), &multi)
	require.Len(t, multi, 4)
	var pointers []string
	for _, err := range multi {
		require.ErrorAs(t, err, &leaf)
		pointers = append(pointers, leaf.JSONPointer)
	}
	require.Equal(t, []string{"/security/0/a~0~1", "/security/0/z", "/security/1/other", "/paths/~1pets/get/security/0/missing"}, pointers)
	var path *openapi3.PathValidationError
	require.ErrorAs(t, multi[3], &path)
	require.Equal(t, "/pets", path.Path)
	var operation *openapi3.OperationValidationError
	require.ErrorAs(t, multi[3], &operation)
	require.Equal(t, "GET", operation.Method)

	// Existing structural failures retain precedence; multi-error mode also
	// includes the independently invalid security requirements.
	doc.Info.Title = ""
	var title *openapi3.InfoTitleRequired
	require.ErrorAs(t, doc.Validate(t.Context()), &title)
	require.ErrorAs(t, doc.Validate(t.Context(), openapi3.EnableMultiError()), &multi)
	require.Len(t, multi, 5)
}

func TestIssue1082_SharedCallbacksAndDocumentNamespace(t *testing.T) {
	t.Parallel()

	doc := securityRequirementsDoc(t)
	op := securityRequirementsOperation()
	callback := securityRequirementsCallback(op)
	// A callback referring back to the same operation is a finite graph.
	op.Callbacks = openapi3.Callbacks{"again": callback}
	doc.Paths.Value("/pets").Get.Callbacks = openapi3.Callbacks{"first": callback, "second": callback}
	var multi openapi3.MultiError
	require.ErrorAs(t, doc.Validate(t.Context(), openapi3.EnableMultiError()), &multi)
	require.Len(t, multi, 1)
	var leaf *openapi3.SecurityRequirementSchemeUndefinedError
	require.ErrorAs(t, multi[0], &leaf)
	require.Equal(t, "/paths/~1pets/get/callbacks/first/{$request.body#~1url}/post/security/0/missing", leaf.JSONPointer)

	other := securityRequirementsDoc(t)
	other.Paths = doc.Paths
	other.Components.SecuritySchemes["missing"] = other.Components.SecuritySchemes["bearerAuth"]
	require.NoError(t, other.Validate(t.Context()))
	require.ErrorAs(t, doc.Validate(t.Context()), &leaf)
}

func TestIssue1082_EntryDocumentForExternalPath(t *testing.T) {
	t.Parallel()

	loader := openapi3.NewLoader()
	loader.IncludeOrigin = true
	loader.ReadFromURIFunc = func(_ *openapi3.Loader, location *url.URL) ([]byte, error) {
		require.Equal(t, "external.yml", location.Path)
		return []byte(`
get:
  security: [{entryAuth: []}]
  responses:
    "200": {description: ok}
`), nil
	}
	doc, err := loader.LoadFromDataWithPath([]byte(`
openapi: 3.1.2
info: {title: Example, version: "1"}
paths:
  /pets: {$ref: external.yml}
components:
  securitySchemes:
    entryAuth: {type: http, scheme: bearer}
`), &url.URL{Path: "entry.yml"})
	require.NoError(t, err)
	require.NoError(t, doc.Validate(t.Context()))
	delete(doc.Components.SecuritySchemes, "entryAuth")
	var leaf *openapi3.SecurityRequirementSchemeUndefinedError
	require.ErrorAs(t, doc.Validate(t.Context()), &leaf)
	require.Same(t, doc.Paths.Value("/pets").Get.Origin, leaf.Origin)
	require.NotNil(t, leaf.Origin)
	require.Equal(t, "entryAuth", leaf.Scheme)
}

func TestIssue1082_VersionAndUnresolvedControls(t *testing.T) {
	t.Parallel()

	for _, version := range []string{"3.2.0", "3.3.0", "unrecognized"} {
		doc := securityRequirementsDoc(t)
		doc.OpenAPI = version
		doc.Paths.Value("/pets").Get.Callbacks = openapi3.Callbacks{"event": securityRequirementsCallback(securityRequirementsOperation())}
		require.NoError(t, doc.Validate(t.Context()))
	}
	doc := securityRequirementsDoc(t)
	doc.Security = openapi3.SecurityRequirements{{"declared": {}}}
	doc.Components.SecuritySchemes["declared"] = &openapi3.SecuritySchemeRef{Ref: "#/components/securitySchemes/absent"}
	var unresolved *openapi3.UnresolvedRefError
	err := doc.Validate(t.Context(), openapi3.EnableMultiError())
	require.ErrorAs(t, err, &unresolved)
	var undefined *openapi3.SecurityRequirementSchemeUndefinedError
	require.False(t, errors.As(err, &undefined), "declared but unresolved is an existing reference error")
}

func TestIssue1082_RootExtensionsLast(t *testing.T) {
	t.Parallel()

	for _, version := range []string{"3.0.3", "3.1.2", "3.2.0", "3.3.0", "unrecognized"} {
		t.Run(version, func(t *testing.T) {
			doc := securityRequirementsDoc(t)
			doc.OpenAPI = version
			doc.Security = openapi3.SecurityRequirements{{"missing": {}}}
			doc.Extensions = map[string]any{"extra": true, "x-custom": true}
			var scheme *openapi3.SecurityRequirementSchemeUndefinedError
			var extension *openapi3.ExtraSiblingFieldsError
			var multi openapi3.MultiError

			err := doc.Validate(t.Context())
			require.ErrorAs(t, doc.Validate(t.Context(), openapi3.EnableMultiError()), &multi)
			if version == "3.0.3" || version == "3.1.2" {
				// As elsewhere in document validation, fields precede extensions.
				require.ErrorAs(t, err, &scheme)
				require.Equal(t, "/security/0/missing", scheme.JSONPointer)
				require.Len(t, multi, 2)
				require.ErrorAs(t, multi[0], &scheme)
			} else {
				// URI-capable and unrecognized versions do not resolve names locally.
				require.ErrorAs(t, err, &extension)
				require.Len(t, multi, 1)
			}
			require.ErrorAs(t, multi[len(multi)-1], &extension)

			err = doc.Validate(t.Context(), openapi3.AllowExtraSiblingFields("extra"))
			if version == "3.0.3" || version == "3.1.2" {
				require.ErrorAs(t, err, &scheme)
			} else {
				require.NoError(t, err)
			}
			doc.Security = openapi3.SecurityRequirements{{"bearerAuth": {}}}
			require.ErrorAs(t, doc.Validate(t.Context()), &extension)
			require.NoError(t, doc.Validate(t.Context(), openapi3.AllowExtraSiblingFields("extra")))

			doc.Info.Title = ""
			var title *openapi3.InfoTitleRequired
			require.ErrorAs(t, doc.Validate(t.Context()), &title)
		})
	}
}
