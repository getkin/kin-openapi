package openapi2conv_test

import (
	"testing"

	"github.com/oasdiff/yaml"
	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
)

// Regression test for #1062.
//
// FromV3RequestBodyFormData iterates the properties of a form-data schema
// and, for any property typed as an array, recurses into the items via
// FromV3SchemaRef(val.Items, nil) — passing the components table as nil.
// With the old code, an items entry that was a $ref into
// #/components/schemas/... nil-dereferenced on `components.Schemas[name]`
// before the lookup could even happen.
//
// The minified spec below reproduces the OpenAI OpenAPI 3 shape that
// triggered the original report: a multipart/form-data request body whose
// `documents` field is an array of $ref-ed component schemas. This must
// convert cleanly without a panic.
func TestIssue1062_FormDataArrayOfRefDoesNotPanic(t *testing.T) {
	const v3Spec = `
openapi: 3.0.3
info:
  title: issue 1062 minified
  version: 1.0.0
paths:
  /v1/upload:
    post:
      operationId: uploadDocuments
      requestBody:
        required: true
        content:
          multipart/form-data:
            schema:
              type: object
              properties:
                documents:
                  type: array
                  items:
                    $ref: '#/components/schemas/Document'
      responses:
        '200':
          description: ok
components:
  schemas:
    Document:
      type: object
      properties:
        id:
          type: string
        name:
          type: string
`

	var doc3 openapi3.T
	err := yaml.Unmarshal([]byte(v3Spec), &doc3)
	require.NoError(t, err, "unmarshal v3 spec")

	// Pre-fix: this call panicked with
	//   "runtime error: invalid memory reference or nil pointer dereference"
	// inside FromV3SchemaRef when it deref'd nil components.Schemas.
	doc2, err := openapi2conv.FromV3(&doc3)
	require.NoError(t, err, "FromV3 must not error on form-data array of $refs")
	require.NotNil(t, doc2)

	// Sanity: the operation made it through to v2 with a formData parameter.
	op := doc2.Paths["/v1/upload"].Post
	require.NotNil(t, op, "POST /v1/upload should be present after conversion")
	var sawDocuments bool
	for _, p := range op.Parameters {
		if p.In == "formData" && p.Name == "documents" {
			sawDocuments = true
			break
		}
	}
	require.True(t, sawDocuments, "expected a formData parameter named 'documents'")
}
