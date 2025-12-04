package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func oneofSpec(t *testing.T) *T {
	t.Helper()

	spec := []byte(`
openapi: 3.0.1
paths: {}
info:
  version: 1.1.1
  title: title
components:
  schemas:
    Cat:
      type: object
      properties:
        name:
          type: string
        scratches:
          type: boolean
        $type:
          type: string
          enum:
            - cat
      required:
        - name
        - scratches
        - $type
    Dog:
      type: object
      properties:
        name:
          type: string
        barks:
          type: boolean
        $type:
          type: string
          enum:
            - dog
      required:
        - name
        - barks
        - $type
    Animal:
      type: object
      oneOf:
        - $ref: "#/components/schemas/Cat"
        - $ref: "#/components/schemas/Dog"
      discriminator:
        propertyName: $type
        mapping:
          cat: "#/components/schemas/Cat"
          dog: "#/components/schemas/Dog"
`[1:])

	loader := NewLoader()
	doc, err := loader.LoadFromData(spec)
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	return doc
}

func oneofNoDiscriminatorSpec(t *testing.T) *T {
	t.Helper()

	spec := []byte(`
openapi: 3.0.1
paths: {}
info:
  version: 1.1.1
  title: title
components:
  schemas:
    Cat:
      type: object
      properties:
        name:
          type: string
        scratches:
          type: boolean
      required:
        - name
        - scratches
    Dog:
      type: object
      properties:
        name:
          type: string
        barks:
          type: boolean
      required:
        - name
        - barks
    Animal:
      type: object
      oneOf:
        - $ref: "#/components/schemas/Cat"
        - $ref: "#/components/schemas/Dog"
`[1:])

	loader := NewLoader()
	doc, err := loader.LoadFromData(spec)
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	return doc
}

func TestVisitJSON_OneOf_MissingDescriptorProperty(t *testing.T) {
	doc := oneofSpec(t)
	err := doc.Components.Schemas["Animal"].Value.VisitJSON(map[string]any{
		"name": "snoopy",
	})
	require.ErrorContains(t, err, `input does not contain the discriminator property "$type"`)
}

func TestVisitJSON_OneOf_MissingDescriptorValue(t *testing.T) {
	doc := oneofSpec(t)
	err := doc.Components.Schemas["Animal"].Value.VisitJSON(map[string]any{
		"name":  "snoopy",
		"$type": "snake",
	})
	require.ErrorContains(t, err, `discriminator property "$type" has invalid value`)
}

func TestVisitJSON_OneOf_MissingField(t *testing.T) {
	doc := oneofSpec(t)
	err := doc.Components.Schemas["Animal"].Value.VisitJSON(map[string]any{
		"name":  "snoopy",
		"$type": "dog",
	})
	require.ErrorContains(t, err, `doesn't match schema due to: Error at "/barks": property "barks" is missing`)
}

func TestVisitJSON_OneOf_NoDescriptor_MissingField(t *testing.T) {
	doc := oneofNoDiscriminatorSpec(t)
	err := doc.Components.Schemas["Animal"].Value.VisitJSON(map[string]any{
		"name": "snoopy",
	})
	require.ErrorContains(t, err, `doesn't match schema due to: Error at "/scratches": property "scratches" is missing`)
}

func TestVisitJSON_OneOf_BadDiscriminatorType(t *testing.T) {
	doc := oneofSpec(t)
	err := doc.Components.Schemas["Animal"].Value.VisitJSON(map[string]any{
		"name":      "snoopy",
		"scratches": true,
		"$type":     1,
	})
	require.ErrorContains(t, err, `value of discriminator property "$type" is not a string`)

	err = doc.Components.Schemas["Animal"].Value.VisitJSON(map[string]any{
		"name":  "snoopy",
		"barks": true,
		"$type": nil,
	})
	require.ErrorContains(t, err, `value of discriminator property "$type" is not a string`)
}

func TestVisitJSON_OneOf_Path(t *testing.T) {
	spec := []byte(`
openapi: 3.0.0
paths: {}
info:
  version: 1.1.1
  title: title
components:
  schemas:
    Something:
      type: object
      properties:
        first:
          type: object
          properties:
            second:
              type: object
              properties:
                third:
                  oneOf:
                   - title: First rule
                     type: string
                     minLength: 5
                     maxLength: 5
                   - title: Second rule
                     type: string
                     minLength: 10
                     maxLength: 10
`[1:])

	loader := NewLoader()
	doc, err := loader.LoadFromData(spec)
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	err = doc.Components.Schemas["Something"].Value.VisitJSON(map[string]any{
		"first": map[string]any{
			"second": map[string]any{
				"third": "123456789",
			},
		},
	})

	require.ErrorContains(t, err, `Error at "/first/second/third"`)

	var sErr *SchemaError

	require.ErrorAs(t, err, &sErr)
	require.Equal(t, []string{"first", "second", "third"}, sErr.JSONPointer())
}
