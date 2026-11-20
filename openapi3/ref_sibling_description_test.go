package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

// What is written next to a $ref in 3.1 applies to that reference only: the
// component it points to, which every other reference shares, is left alone.
func TestRefSiblingDescriptionDoesNotChangeComponent(t *testing.T) {
	t.Parallel()

	doc := []byte(`
openapi: 3.1.0
info: {title: t, version: "1"}
paths:
  /a:
    get:
      parameters:
        - {$ref: '#/components/parameters/P', description: mine}
        - {$ref: '#/components/parameters/P', description}
      requestBody: {$ref: '#/components/requestBodies/B', description: mine}
      responses:
        "200": {$ref: '#/components/responses/R', description: mine}
        "201": {$ref: '#/components/responses/R', description}
components:
  parameters:
    P: {name: p, in: query, description: orig, schema: {type: string}}
  requestBodies:
    B: {description: orig, content: {a/b: {}}}
  responses:
    R: {description: orig}
`)

	loader := openapi3.NewLoader()
	got, err := loader.LoadFromData(doc)
	require.NoError(t, err)

	c := got.Components
	require.Equal(t, "orig", c.Parameters["P"].Value.Description)
	require.Equal(t, "orig", c.RequestBodies["B"].Value.Description)
	require.Equal(t, "orig", *c.Responses["R"].Value.Description)

	op := got.Paths.Find("/a").Get
	require.Equal(t, "mine", op.Parameters[0].Value.Description)
	require.Equal(t, "orig", op.Parameters[1].Value.Description)
	require.Equal(t, "mine", *op.Responses.Value("200").Value.Description)
	require.Equal(t, "orig", *op.Responses.Value("201").Value.Description)
}
