package openapi3

import (
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
	require.Equal(t, 0, len(doc.Paths))

	require.Equal(t, "string", doc.Components.Schemas["schema2"].Value.Properties["prop"].Value.Type)
}

func TestUnmarshallingMultijsonTag(t *testing.T) {
	spec := []byte(`
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

	loader := NewLoader()

	doc, err := loader.LoadFromData(spec)
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	for propName, propSchema := range doc.Components.Schemas {
		ap := propSchema.Value.AdditionalProperties
		apa := propSchema.Value.AdditionalPropertiesAllowed

		if propName == "unset" {
			require.True(t, ap == nil && apa == nil)
			continue
		}

		apStr := ""
		if ap != nil {
			apStr = fmt.Sprintf("{Ref:- Value.AdditionalProperties:%+v Value.AdditionalPropertiesAllowed:%+v}", (*ap).Value.AdditionalProperties, (*ap).Value.AdditionalPropertiesAllowed)
		}
		apaStr := ""
		if apa != nil {
			apaStr = fmt.Sprintf("%v", *apa)
		}

		require.Truef(t, (ap != nil && apa == nil) || (ap == nil && apa != nil),
			"%s: isnil(%s) xor isnil(%s)", propName, apaStr, apStr)
	}
}
