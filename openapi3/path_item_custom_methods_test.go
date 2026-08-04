package openapi3

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

// customMethodsSpec exercises both OpenAPI 3.2 path item additions: the fixed
// `query` field and an `additionalOperations` entry, each with a $ref so ref
// resolution through the new operations is covered too.
const customMethodsSpec = `
openapi: 3.2.0
info:
  title: Custom methods
  version: 1.0.0
paths:
  /things:
    query:
      parameters:
        - $ref: "#/components/parameters/Trace"
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                filter:
                  type: string
      responses:
        "200":
          description: results
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Query"
    additionalOperations:
      COPY:
        parameters:
          - name: destination
            in: header
            required: true
            schema:
              type: string
        responses:
          "204":
            description: copied
components:
  parameters:
    Trace:
      name: trace
      in: query
      schema:
        type: boolean
  schemas:
    Query:
      type: object
      properties:
        filter:
          type: string
`

func TestOpenAPI32PathItemCustomMethods(t *testing.T) {
	loader := NewLoader()
	doc, err := loader.LoadFromData([]byte(customMethodsSpec))
	require.NoError(t, err)
	require.NoError(t, doc.Validate(t.Context()))

	pathItem := doc.Paths.Value("/things")
	require.NotNil(t, pathItem)
	require.NotNil(t, pathItem.Query)
	require.Len(t, pathItem.AdditionalOperations, 1)
	require.NotNil(t, pathItem.AdditionalOperations["COPY"])

	// $ref inside the query operation got resolved.
	schemaRef := pathItem.Query.Responses.Status(200).Value.Content.Get("application/json").Schema
	require.Equal(t, "#/components/schemas/Query", schemaRef.Ref)
	require.NotNil(t, schemaRef.Value)
	require.True(t, schemaRef.Value.Type.Is(TypeObject))
	require.NotNil(t, pathItem.Query.Parameters[0].Value)
	require.Equal(t, "trace", pathItem.Query.Parameters[0].Value.Name)

	ops := pathItem.Operations()
	require.Contains(t, ops, "QUERY")
	require.Contains(t, ops, "COPY")
	require.Same(t, pathItem.Query, ops["QUERY"])
	require.Same(t, pathItem.AdditionalOperations["COPY"], ops["COPY"])

	require.Same(t, pathItem.Query, pathItem.GetOperation("QUERY"))
	require.Same(t, pathItem.AdditionalOperations["COPY"], pathItem.GetOperation("COPY"))
	require.Nil(t, pathItem.GetOperation("PURGE"))
}

func TestPathItemSetOperationCustomMethod(t *testing.T) {
	pathItem := &PathItem{}
	op := &Operation{Responses: NewResponses()}

	// Used to panic with "unsupported HTTP method".
	require.NotPanics(t, func() { pathItem.SetOperation("COPY", op) })
	require.Same(t, op, pathItem.AdditionalOperations["COPY"])
	require.Same(t, op, pathItem.GetOperation("COPY"))

	query := &Operation{Responses: NewResponses()}
	pathItem.SetOperation("QUERY", query)
	require.Same(t, query, pathItem.Query)

	// A nil operation removes the custom method instead of storing nil.
	pathItem.SetOperation("COPY", nil)
	require.NotContains(t, pathItem.AdditionalOperations, "COPY")
	require.Nil(t, pathItem.GetOperation("COPY"))

	// Fixed fields keep their nil-assignment semantics.
	pathItem.SetOperation("QUERY", nil)
	require.Nil(t, pathItem.Query)
}

func TestPathItemCustomMethodsRoundTrip(t *testing.T) {
	loader := NewLoader()
	doc, err := loader.LoadFromData([]byte(customMethodsSpec))
	require.NoError(t, err)

	pathItem := doc.Paths.Value("/things")
	require.Empty(t, pathItem.Extensions)

	data, err := json.Marshal(pathItem)
	require.NoError(t, err)

	var raw map[string]any
	require.NoError(t, json.Unmarshal(data, &raw))
	require.Contains(t, raw, "query")
	require.Contains(t, raw, "additionalOperations")
	additional, ok := raw["additionalOperations"].(map[string]any)
	require.True(t, ok)
	require.Contains(t, additional, "COPY")

	var back PathItem
	require.NoError(t, json.Unmarshal(data, &back))
	require.Empty(t, back.Extensions)
	require.NotNil(t, back.Query)
	require.NotNil(t, back.AdditionalOperations["COPY"])
}

func TestPathItemCustomMethodsValidationErrors(t *testing.T) {
	for _, tc := range []struct {
		name string
		spec string
		code string
	}{
		{
			name: "query needs OpenAPI 3.2",
			spec: `
openapi: 3.1.0
info: {title: t, version: "1"}
paths:
  /things:
    query:
      responses:
        "200": {description: ok}
`,
			code: "query-field-for-3-2-plus",
		},
		{
			name: "additionalOperations needs OpenAPI 3.2",
			spec: `
openapi: 3.1.0
info: {title: t, version: "1"}
paths:
  /things:
    additionalOperations:
      COPY:
        responses:
          "204": {description: copied}
`,
			code: "additional-operations-field-for-3-2-plus",
		},
		{
			name: "additionalOperations duplicates a fixed field",
			spec: `
openapi: 3.2.0
info: {title: t, version: "1"}
paths:
  /things:
    additionalOperations:
      GET:
        responses:
          "200": {description: ok}
`,
			code: "additional-operations-duplicate-method",
		},
		{
			name: "additionalOperations duplicates the query field",
			spec: `
openapi: 3.2.0
info: {title: t, version: "1"}
paths:
  /things:
    additionalOperations:
      QUERY:
        responses:
          "200": {description: ok}
`,
			code: "additional-operations-duplicate-method",
		},
		{
			name: "additionalOperations key is not an HTTP token",
			spec: `
openapi: 3.2.0
info: {title: t, version: "1"}
paths:
  /things:
    additionalOperations:
      "BAD METHOD":
        responses:
          "200": {description: ok}
`,
			code: "additional-operations-invalid-method",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			loader := NewLoader()
			doc, err := loader.LoadFromData([]byte(tc.spec))
			require.NoError(t, err)

			err = doc.Validate(t.Context())
			require.Error(t, err)
			var coded CodedError
			require.ErrorAs(t, err, &coded)
			require.Equal(t, tc.code, coded.Code())
		})
	}
}

func TestPathItemCustomMethodsVersionSentinels(t *testing.T) {
	loader := NewLoader()
	doc, err := loader.LoadFromData([]byte(`
openapi: 3.1.0
info: {title: t, version: "1"}
paths:
  /things:
    query:
      responses:
        "200": {description: ok}
    additionalOperations:
      COPY:
        responses:
          "204": {description: copied}
`))
	require.NoError(t, err)

	err = doc.Validate(t.Context(), EnableMultiError())
	require.Error(t, err)

	var multi MultiError
	require.ErrorAs(t, err, &multi)

	var query *QueryFieldFor32Plus
	var additional *AdditionalOperationsFieldFor32Plus
	var mismatches int
	for _, e := range multi {
		var fvm *FieldVersionMismatchError
		if errors.As(e, &fvm) {
			mismatches++
		}
		if q := new(QueryFieldFor32Plus); errors.As(e, &q) {
			query = q
		}
		if a := new(AdditionalOperationsFieldFor32Plus); errors.As(e, &a) {
			additional = a
		}
	}
	require.Equal(t, 2, mismatches)
	require.NotNil(t, query)
	require.NotNil(t, additional)
	require.Equal(t, "field query is for OpenAPI >=3.2", query.Error())
	require.Equal(t, "field additionalOperations is for OpenAPI >=3.2", additional.Error())
}

// A lowercase key is a valid RFC 9110 token, so it must not be rejected even
// though the convention is uppercase.
func TestPathItemCustomMethodsLowercaseKeyIsValid(t *testing.T) {
	loader := NewLoader()
	doc, err := loader.LoadFromData([]byte(`
openapi: 3.2.0
info: {title: t, version: "1"}
paths:
  /things:
    additionalOperations:
      copy:
        responses:
          "204": {description: copied}
`))
	require.NoError(t, err)
	require.NoError(t, doc.Validate(t.Context()))
}

func TestPathItemIsEmptyWithCustomMethods(t *testing.T) {
	require.True(t, (&PathItem{}).isEmpty())
	require.False(t, (&PathItem{Query: &Operation{}}).isEmpty())
	require.False(t, (&PathItem{AdditionalOperations: map[string]*Operation{
		"COPY": {},
	}}).isEmpty())
}

func TestWalkersVisitCustomMethodOperations(t *testing.T) {
	loader := NewLoader()
	doc, err := loader.LoadFromData([]byte(customMethodsSpec))
	require.NoError(t, err)

	var schemaPointers []string
	require.NoError(t, doc.WalkSchemas(func(jsonPointer string, _ *SchemaRef) error {
		schemaPointers = append(schemaPointers, jsonPointer)
		return nil
	}))
	require.Contains(t, schemaPointers,
		"/paths/~1things/query/requestBody/content/application~1json/schema")
	require.Contains(t, schemaPointers,
		"/paths/~1things/additionalOperations/COPY/parameters/0/schema")

	var paramPointers []string
	require.NoError(t, doc.WalkParameters(func(jsonPointer string, _ *ParameterRef) error {
		paramPointers = append(paramPointers, jsonPointer)
		return nil
	}))
	require.Contains(t, paramPointers,
		"/paths/~1things/additionalOperations/COPY/parameters/0")
}
