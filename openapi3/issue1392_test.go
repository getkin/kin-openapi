package openapi3_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestInternalizeRefsWebhooks(t *testing.T) {
	t.Parallel()

	const operation = `
post:
  requestBody:
    content:
      application/json:
        schema:
          $ref: ./Item.yaml
  responses:
    '200':
      description: OK
`
	for _, test := range []struct {
		name       string
		document   string
		schemaName string
	}{
		{
			name: "external path control",
			document: `paths:
  /ping:
    $ref: ./ping.yaml
`,
			schemaName: "Item",
		},
		{
			name: "external webhook without paths",
			document: `webhooks:
  ping:
    $ref: ./ping.yaml
`,
			schemaName: "Item",
		},
		{
			name: "inline webhook",
			document: `paths: {}
webhooks:
  ping:
    post:
      requestBody:
        content:
          application/json:
            schema:
              $ref: ./Item.yaml
      responses:
        '200':
          description: OK
`,
			schemaName: "Item",
		},
		{
			name: "external webhook with local component reference",
			document: `webhooks:
  ping:
    $ref: ./external.yaml#/paths/~1ping
`,
			schemaName: "external_Item",
		},
		{
			name: "path and webhook sharing references",
			document: `paths:
  /ping:
    $ref: ./ping.yaml
webhooks:
  ping:
    $ref: ./ping.yaml
`,
			schemaName: "Item",
		},
	} {
		for _, prefix := range []string{"", "custom_"} {
			t.Run(test.name+"/"+prefix, func(t *testing.T) {
				t.Parallel()
				dir := t.TempDir()
				for name, content := range map[string]string{
					"root.yaml":  "openapi: 3.1.0\ninfo: {title: t, version: '1'}\n" + test.document,
					"ping.yaml":  operation,
					"Item.yaml":  "type: object\nproperties:\n  child:\n    $ref: ./Child.yaml\n",
					"Child.yaml": "type: integer\n",
					"external.yaml": `openapi: 3.1.0
info: {title: external, version: '1'}
paths:
  /ping:
    post:
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Item'
      responses:
        '200':
          description: OK
components:
  schemas:
    Item:
      type: object
      properties:
        child:
          $ref: ./Child.yaml
`,
				} {
					require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600))
				}

				loader := openapi3.NewLoader()
				loader.IsExternalRefsAllowed = true
				doc, err := loader.LoadFromFile(filepath.Join(dir, "root.yaml"))
				require.NoError(t, err)
				require.NoError(t, doc.Validate(t.Context()))

				var resolver openapi3.RefNameResolver
				if prefix != "" {
					resolver = func(doc *openapi3.T, ref openapi3.ComponentRef) string {
						return prefix + openapi3.DefaultRefNameResolver(doc, ref)
					}
				}
				doc.InternalizeRefs(t.Context(), resolver)
				require.NoError(t, doc.Validate(t.Context()))

				for _, item := range []*openapi3.PathItem{doc.Webhooks["ping"], doc.Paths.Value("/ping")} {
					if item == nil {
						continue
					}
					require.Empty(t, item.Ref)
					schema := item.Post.RequestBody.Value.Content["application/json"].Schema
					require.Equal(t, "#/components/schemas/"+prefix+test.schemaName, schema.Ref)
					require.Equal(t, schema.Value, doc.Components.Schemas[prefix+test.schemaName].Value)
					require.Equal(t, "#/components/schemas/"+prefix+"Child", schema.Value.Properties["child"].Ref)
				}
				require.Len(t, doc.Components.Schemas, 2)

				data, err := doc.MarshalJSON()
				require.NoError(t, err)
				loader = openapi3.NewLoader()
				reloaded, err := loader.LoadFromData(data)
				require.NoError(t, err, "internalized webhooks must not require external access")
				require.NoError(t, reloaded.Validate(t.Context()))

				doc.InternalizeRefs(t.Context(), resolver)
				again, err := doc.MarshalJSON()
				require.NoError(t, err)
				require.JSONEq(t, string(data), string(again))
			})
		}
	}
}

func TestInternalizeRefsEmptyWebhooks(t *testing.T) {
	t.Parallel()
	for _, webhooks := range []map[string]*openapi3.PathItem{nil, {}} {
		doc := &openapi3.T{Webhooks: webhooks}
		require.NotPanics(t, func() { doc.InternalizeRefs(t.Context(), nil) })
		require.Nil(t, doc.Components)
	}
}
