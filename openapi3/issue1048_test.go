package openapi3_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

// issue1048Doc builds a document shaped like the one in issue #1048: a Device
// whose spec.config items are a ConfigProviderSpec with a oneOf of several
// $refs (Git/KubernetesSecret/Inline/Http variants).
func issue1048Doc(t *testing.T) *openapi3.T {
	t.Helper()

	spec := []byte(`
openapi: 3.0.1
paths: {}
info:
  version: 1.1.1
  title: title
components:
  schemas:
    Device:
      type: object
      properties:
        apiVersion:
          type: string
        kind:
          type: string
        metadata:
          type: object
        spec:
          type: object
          properties:
            config:
              type: array
              items:
                $ref: "#/components/schemas/ConfigProviderSpec"
    ConfigProviderSpec:
      oneOf:
        - $ref: "#/components/schemas/GitConfigProviderSpec"
        - $ref: "#/components/schemas/KubernetesSecretProviderSpec"
        - $ref: "#/components/schemas/InlineConfigProviderSpec"
        - $ref: "#/components/schemas/HttpConfigProviderSpec"
    GitConfigProviderSpec:
      type: object
      properties:
        name:
          type: string
        gitRef:
          type: object
          properties:
            repository:
              type: string
            targetRevision:
              type: string
            path:
              type: string
            mountPath:
              type: string
              default: "/"
          required:
            - repository
            - targetRevision
            - path
      required:
        - name
        - gitRef
    KubernetesSecretProviderSpec:
      type: object
      properties:
        name:
          type: string
        secretRef:
          type: object
          properties:
            name:
              type: string
            namespace:
              type: string
            mountPath:
              type: string
          required:
            - name
            - namespace
            - mountPath
      required:
        - name
        - secretRef
    InlineConfigProviderSpec:
      type: object
      properties:
        name:
          type: string
        inline:
          type: array
          items:
            $ref: '#/components/schemas/FileSpec'
      required:
        - name
        - inline
    FileSpec:
      type: object
      properties:
        path:
          type: string
        content:
          type: string
        contentEncoding:
          type: string
          enum:
            - plain
            - base64
        mode:
          type: integer
        user:
          type: string
        group:
          type: string
      required:
        - path
        - content
    HttpConfigProviderSpec:
      type: object
      properties:
        name:
          type: string
        httpRef:
          type: object
          properties:
            repository:
              type: string
            suffix:
              type: string
            filePath:
              type: string
          required:
            - repository
            - filePath
      required:
        - name
        - httpRef
`[1:])

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(spec)
	require.NoError(t, err)
	require.NoError(t, doc.Validate(loader.Context))
	return doc
}

// TestIssue1048_OneOfNoMatchListsEachBranchError validates the invalid
// configuration from the issue (an inline provider whose FileSpec is missing
// the required "content" property) and checks that the oneOf error names the
// oneOf and lists every branch's error, instead of reporting only one branch
// or dumping each branch's whole schema.
func TestIssue1048_OneOfNoMatchListsEachBranchError(t *testing.T) {
	t.Parallel()

	doc := issue1048Doc(t)
	device := doc.Components.Schemas["Device"].Value

	// The payload from the issue.
	payload := map[string]any{
		"apiVersion": "v1alpha1",
		"kind":       "Device",
		"metadata": map[string]any{
			"name":   "fab9839018890a88b898b980f8f809f8e8ac333761977d987a777a777a987c",
			"labels": map[string]any{"fleet": "fleet-b"},
		},
		"spec": map[string]any{
			"config": []any{
				map[string]any{
					"name":   "motd",
					"inline": []any{map[string]any{"mode": 511, "path": "/etc/motd"}},
				},
			},
		},
	}

	err := device.VisitJSON(payload)
	require.Error(t, err)

	// The error is still a *SchemaError rooted at the oneOf, carrying the
	// path to the offending value.
	var schemaErr *openapi3.SchemaError
	require.ErrorAs(t, err, &schemaErr)
	require.Equal(t, "oneOf", schemaErr.SchemaField)
	require.Equal(t, `value doesn't match any schema from "oneOf"`, schemaErr.Reason)
	require.Equal(t, []string{"spec", "config", "0"}, schemaErr.JSONPointer())

	// The message names the oneOf and lists each branch's error, labeled by
	// branch index and $ref, with the full path to each failure.
	require.ErrorContains(t, err, `Error at "/spec/config/0": value doesn't match any schema from "oneOf":`)
	require.ErrorContains(t, err, `oneOf[0] ($ref "#/components/schemas/GitConfigProviderSpec"): Error at "/spec/config/0/gitRef": property "gitRef" is missing`)
	require.ErrorContains(t, err, `oneOf[1] ($ref "#/components/schemas/KubernetesSecretProviderSpec"): Error at "/spec/config/0/secretRef": property "secretRef" is missing`)
	require.ErrorContains(t, err, `oneOf[2] ($ref "#/components/schemas/InlineConfigProviderSpec"): Error at "/spec/config/0/inline/0/content": property "content" is missing`)
	require.ErrorContains(t, err, `oneOf[3] ($ref "#/components/schemas/HttpConfigProviderSpec"): Error at "/spec/config/0/httpRef": property "httpRef" is missing`)

	// Branch errors are listed compactly: no per-branch schema/value dump
	// drowns the reasons.
	msg := err.Error()
	require.NotContains(t, msg, "Schema:")
	require.NotContains(t, msg, "Value:")

	// The branch errors remain reachable programmatically as a MultiError of
	// *SchemaError, each with its own path, schema and value.
	var me openapi3.MultiError
	require.ErrorAs(t, err, &me)
	require.Len(t, me, 4)
	for _, branch := range me {
		var branchErr *openapi3.SchemaError
		require.ErrorAs(t, branch, &branchErr)
		require.NotNil(t, branchErr.Schema)
		require.NotNil(t, branchErr.Value)
	}
	var inlineBranch *openapi3.SchemaError
	require.ErrorAs(t, me[2], &inlineBranch)
	require.Equal(t, []string{"spec", "config", "0", "inline", "0", "content"}, inlineBranch.JSONPointer())
	require.Equal(t, "required", inlineBranch.SchemaField)
	require.Equal(t, `property "content" is missing`, inlineBranch.Reason)
}

// TestIssue1048_OneOfMultipleMatchNamesIndices checks that a value matching
// more than one oneOf branch reports which branches matched.
func TestIssue1048_OneOfMultipleMatchNamesIndices(t *testing.T) {
	t.Parallel()

	doc := issue1048Doc(t)
	schema := doc.Components.Schemas["ConfigProviderSpec"].Value

	// Matches both the Git and the KubernetesSecret variants.
	err := schema.VisitJSON(map[string]any{
		"name":      "x",
		"gitRef":    map[string]any{"repository": "r", "targetRevision": "t", "path": "p"},
		"secretRef": map[string]any{"name": "n", "namespace": "ns", "mountPath": "/"},
	})
	require.Error(t, err)

	var schemaErr *openapi3.SchemaError
	require.ErrorAs(t, err, &schemaErr)
	require.Equal(t, "oneOf", schemaErr.SchemaField)
	require.Equal(t, `value matches more than one schema from "oneOf" (matches schemas at indices [0 1])`, schemaErr.Reason)
	require.ErrorIs(t, err, openapi3.ErrOneOfConflict)
	require.ErrorContains(t, err, `value matches more than one schema from "oneOf" (matches schemas at indices [0 1])`)
}

// TestIssue1048_NestedOneOf checks that a oneOf nested inside a oneOf branch
// is reported as its own labeled list within the outer branch's error.
func TestIssue1048_NestedOneOf(t *testing.T) {
	t.Parallel()

	spec := []byte(`
openapi: 3.0.1
paths: {}
info:
  version: 1.1.1
  title: title
components:
  schemas:
    Outer:
      oneOf:
        - $ref: "#/components/schemas/Inner"
        - type: object
          properties:
            name:
              type: string
          required:
            - name
    Inner:
      oneOf:
        - type: string
          minLength: 5
        - type: integer
`[1:])

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(spec)
	require.NoError(t, err)
	require.NoError(t, doc.Validate(loader.Context))

	// A boolean matches neither the inner oneOf's branches nor the object branch.
	err = doc.Components.Schemas["Outer"].Value.VisitJSON(true)
	require.Error(t, err)

	require.ErrorContains(t, err, `value doesn't match any schema from "oneOf":`)
	require.ErrorContains(t, err, `oneOf[0] ($ref "#/components/schemas/Inner"): value doesn't match any schema from "oneOf":`)
	require.ErrorContains(t, err, `oneOf[0]: value must be a string`)
	require.ErrorContains(t, err, `oneOf[1]: value must be an integer`)
	require.ErrorContains(t, err, `oneOf[1]: value must be an object`)

	// The outer branch errors remain reachable as a MultiError.
	var me openapi3.MultiError
	require.ErrorAs(t, err, &me)
	require.Len(t, me, 2)
	var inner *openapi3.SchemaError
	require.True(t, errors.As(me[0], &inner))
	require.Equal(t, "oneOf", inner.SchemaField)
}

// TestIssue1048_OneOfNoMatchDetailsDisabled checks the message shape when
// schema error details are disabled: each branch is still listed, on its own
// line, without schema/value dumps.
func TestIssue1048_OneOfNoMatchDetailsDisabled(t *testing.T) {
	t.Parallel()

	doc := issue1048Doc(t)
	schema := doc.Components.Schemas["ConfigProviderSpec"].Value

	err := schema.VisitJSON(map[string]any{
		"name":   "motd",
		"inline": []any{map[string]any{"mode": 511, "path": "/etc/motd"}},
	}, openapi3.DisableSchemaErrorDetails())
	require.Error(t, err)

	var schemaErr *openapi3.SchemaError
	require.ErrorAs(t, err, &schemaErr)
	require.Equal(t, "oneOf", schemaErr.SchemaField)
	require.Equal(t, `value doesn't match any schema from "oneOf"`, schemaErr.Reason)
	require.ErrorContains(t, err, `value doesn't match any schema from "oneOf":`)
	require.ErrorContains(t, err, `oneOf[2] ($ref "#/components/schemas/InlineConfigProviderSpec"): Error at "/inline/0/content": property "content" is missing`)

	// The branch errors remain reachable as a MultiError.
	var me openapi3.MultiError
	require.ErrorAs(t, err, &me)
	require.Len(t, me, 4)
}
