package openapi3

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue376(t *testing.T) {
	spec := []byte(`
openapi: 3.0.0
components:
  schemas:
    schema1:
      type: object
      additionalProperties:
        type: string
    schema2:
      type: object
      properties:
        prop:
          $ref: '#/components/schemas/schema1/additionalProperties'
paths: {}
info:
  title: An API
  version: 1.2.3.4
`)

	loader := NewLoader()

	doc, err := loader.LoadFromData(spec)
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	require.Equal(t, "An API", doc.Info.Title)
	require.Equal(t, 2, len(doc.Components.Schemas))
	require.Equal(t, 0, doc.Paths.Len())

	require.Equal(t, "string", doc.Components.Schemas["schema2"].Value.Properties["prop"].Value.Type)
}

func TestExclusiveValuesOfValuesAdditionalProperties(t *testing.T) {
	schema := &Schema{
		AdditionalProperties: AdditionalProperties{
			Has:    BoolPtr(false),
			Schema: NewSchemaRef("", &Schema{}),
		},
	}
	err := schema.Validate(context.Background())
	require.ErrorContains(t, err, ` to both `)

	schema = &Schema{
		AdditionalProperties: AdditionalProperties{
			Has: BoolPtr(false),
		},
	}
	err = schema.Validate(context.Background())
	require.NoError(t, err)

	schema = &Schema{
		AdditionalProperties: AdditionalProperties{
			Schema: NewSchemaRef("", &Schema{}),
		},
	}
	err = schema.Validate(context.Background())
	require.NoError(t, err)
}

func TestMultijsonTagSerialization(t *testing.T) {
	specYAML := []byte(`
openapi: 3.0.0
components:
  schemas:
    unset:
      type: number
    empty-object:
      additionalProperties: {}
    object:
      additionalProperties: {type: string}
    boolean:
      additionalProperties: false
paths: {}
info:
  title: An API
  version: 1.2.3.4
`)

	specJSON := []byte(`{
  "openapi": "3.0.0",
  "components": {
    "schemas": {
      "unset": {
        "type": "number"
      },
      "empty-object": {
        "additionalProperties": {
        }
      },
      "object": {
        "additionalProperties": {
          "type": "string"
        }
      },
      "boolean": {
        "additionalProperties": false
      }
    }
  },
  "paths": {
  },
  "info": {
    "title": "An API",
    "version": "1.2.3.4"
  }
}`)

	for i, spec := range [][]byte{specJSON, specYAML} {
		t.Run(fmt.Sprintf("spec%02d", i), func(t *testing.T) {
			loader := NewLoader()

			doc, err := loader.LoadFromData(spec)
			require.NoError(t, err)

			err = doc.Validate(loader.Context)
			require.NoError(t, err)

			for propName, propSchema := range doc.Components.Schemas {
				t.Run(propName, func(t *testing.T) {
					ap := propSchema.Value.AdditionalProperties.Schema
					apa := propSchema.Value.AdditionalProperties.Has

					apStr := ""
					if ap != nil {
						apStr = fmt.Sprintf("{Ref:%s Value.Type:%v}", (*ap).Ref, (*ap).Value.Type)
					}
					apaStr := ""
					if apa != nil {
						apaStr = fmt.Sprintf("%v", *apa)
					}

					encoded, err := propSchema.MarshalJSON()
					require.NoError(t, err)
					require.Equal(t, map[string]string{
						"unset":        `{"type":"number"}`,
						"empty-object": `{"additionalProperties":{}}`,
						"object":       `{"additionalProperties":{"type":"string"}}`,
						"boolean":      `{"additionalProperties":false}`,
					}[propName], string(encoded))

					if propName == "unset" {
						require.True(t, ap == nil && apa == nil)
						return
					}

					require.Truef(t, (ap != nil && apa == nil) || (ap == nil && apa != nil),
						"%s: isnil(%s) xor isnil(%s)", propName, apaStr, apStr)
				})
			}
		})
	}
}
