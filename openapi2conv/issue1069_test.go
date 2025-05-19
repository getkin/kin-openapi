package openapi2conv

import (
	"context"
	"testing"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIssue1069V2ToV3(t *testing.T) {
	tests := []struct {
		name     string
		v2Spec   string
		validate func(t *testing.T, v3 *openapi3.T)
	}{
		{
			name: "root level externalDocs",
			v2Spec: `
swagger: '2.0'
info:
  version: 1.0.0
  title: title
externalDocs:
  url: https://example/root
  description: Root level documentation
paths:
  /test:
    get:
      responses:
        '200':
          description: description
`,
			validate: func(t *testing.T, v3 *openapi3.T) {
				require.NotNil(t, v3.ExternalDocs)
				assert.Equal(t, "https://example/root", v3.ExternalDocs.URL)
				assert.Equal(t, "Root level documentation", v3.ExternalDocs.Description)
			},
		},
		{
			name: "operation level externalDocs",
			v2Spec: `
swagger: '2.0'
info:
  version: 1.0.0
  title: title
paths:
  /test:
    get:
      externalDocs:
        url: https://example/operation
        description: Operation level documentation
      responses:
        '200':
          description: description
`,
			validate: func(t *testing.T, v3 *openapi3.T) {
				op := v3.Paths.Value("/test").Get
				require.NotNil(t, op.ExternalDocs)
				assert.Equal(t, "https://example/operation", op.ExternalDocs.URL)
				assert.Equal(t, "Operation level documentation", op.ExternalDocs.Description)
			},
		},
		{
			name: "schema level externalDocs",
			v2Spec: `
swagger: '2.0'
info:
  version: 1.0.0
  title: title
definitions:
  TestSchema:
    type: object
    externalDocs:
      url: https://example/schema
      description: Schema level documentation
paths:
  /test:
    get:
      responses:
        '200':
          description: description
          schema:
            $ref: '#/definitions/TestSchema'
`,
			validate: func(t *testing.T, v3 *openapi3.T) {
				schema := v3.Components.Schemas["TestSchema"].Value
				require.NotNil(t, schema.ExternalDocs)
				assert.Equal(t, "https://example/schema", schema.ExternalDocs.URL)
				assert.Equal(t, "Schema level documentation", schema.ExternalDocs.Description)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v3, err := v2v3YAML([]byte(tt.v2Spec))
			require.NoError(t, err)
			err = v3.Validate(context.Background())
			require.NoError(t, err)
			tt.validate(t, v3)
		})
	}
}

func TestIssue1069V3ToV2(t *testing.T) {
	tests := []struct {
		name     string
		v3Spec   string
		validate func(t *testing.T, v2 *openapi2.T)
	}{
		{
			name: "root level externalDocs",
			v3Spec: `
openapi: 3.0.3
info:
  version: 1.0.0
  title: title
externalDocs:
  url: https://example/root
  description: Root level documentation
components: {}
paths:
  /test:
    get:
      responses:
        '200':
          description: description
`,
			validate: func(t *testing.T, v2 *openapi2.T) {
				require.NotNil(t, v2.ExternalDocs)
				assert.Equal(t, "https://example/root", v2.ExternalDocs.URL)
				assert.Equal(t, "Root level documentation", v2.ExternalDocs.Description)
			},
		},
		{
			name: "operation level externalDocs",
			v3Spec: `
openapi: 3.0.3
info:
  version: 1.0.0
  title: title
components: {}
paths:
  /test:
    get:
      externalDocs:
        url: https://example/operation
        description: Operation level documentation
      responses:
        '200':
          description: description
`,
			validate: func(t *testing.T, v2 *openapi2.T) {
				op := v2.Paths["/test"].Get
				require.NotNil(t, op.ExternalDocs)
				assert.Equal(t, "https://example/operation", op.ExternalDocs.URL)
				assert.Equal(t, "Operation level documentation", op.ExternalDocs.Description)
			},
		},
		{
			name: "schema level externalDocs",
			v3Spec: `
openapi: 3.0.3
info:
  version: 1.0.0
  title: title
components:
  schemas:
    TestSchema:
      type: object
      externalDocs:
        url: https://example/schema
        description: Schema level documentation
paths:
  /test:
    get:
      responses:
        '200':
          description: description
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/TestSchema'
`,
			validate: func(t *testing.T, v2 *openapi2.T) {
				schema := v2.Definitions["TestSchema"].Value
				require.NotNil(t, schema.ExternalDocs)
				assert.Equal(t, "https://example/schema", schema.ExternalDocs.URL)
				assert.Equal(t, "Schema level documentation", schema.ExternalDocs.Description)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var doc3 openapi3.T
			err := yaml.Unmarshal([]byte(tt.v3Spec), &doc3)
			require.NoError(t, err)

			v2, err := FromV3(&doc3)
			require.NoError(t, err)
			tt.validate(t, v2)
		})
	}
}
