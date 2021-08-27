package openapi3

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const oneofSpec = `
openapi: "3.0.1"
info:
  title: An API
  version: v1
paths: {}
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
`

const oneofNoDiscriminatorSpec = `
openapi: "3.0.1"
info:
  title: An API
  version: v1
paths: {}
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
`

func TestVisitData_OneOf_MissingDiscriptorProperty(t *testing.T) {
	loader := NewLoader()
	doc, err := loader.LoadFromData([]byte(oneofSpec))
	require.NoError(t, err)
	err = doc.Validate(loader.Context)
	require.NoError(t, err)
	err = doc.Components.Schemas["Animal"].Value.VisitData(doc, map[string]interface{}{
		"name": "snoopy",
	})
	require.EqualError(t, err, "input does not contain the discriminator property")
}

func TestVisitData_OneOf_MissingDiscriptorValue(t *testing.T) {
	loader := NewLoader()
	doc, err := loader.LoadFromData([]byte(oneofSpec))
	require.NoError(t, err)
	err = doc.Validate(loader.Context)
	require.NoError(t, err)
	err = doc.Components.Schemas["Animal"].Value.VisitData(doc, map[string]interface{}{
		"name":  "snoopy",
		"$type": "snake",
	})
	require.EqualError(t, err, "input does not contain a valid discriminator value")
}

func TestVisitData_OneOf_MissingField(t *testing.T) {
	loader := NewLoader()
	doc, err := loader.LoadFromData([]byte(oneofSpec))
	require.NoError(t, err)
	err = doc.Validate(loader.Context)
	require.NoError(t, err)
	err = doc.Components.Schemas["Animal"].Value.VisitData(doc, map[string]interface{}{
		"name":  "snoopy",
		"$type": "dog",
	})
	require.Contains(t, err.Error(), "barks")
	require.True(t, strings.Contains(err.Error(), "is required") || strings.Contains(err.Error(), "is missing"))
}

func TestVisitData_OneOf_NoDiscriptor_MissingField(t *testing.T) {
	loader := NewLoader()
	doc, err := loader.LoadFromData([]byte(oneofNoDiscriminatorSpec))
	require.NoError(t, err)
	err = doc.Validate(loader.Context)
	require.NoError(t, err)
	err = doc.Components.Schemas["Animal"].Value.VisitData(doc, map[string]interface{}{
		"name": "snoopy",
	})
	require.Contains(t, err.Error(), "scratches")
	require.True(t, strings.Contains(err.Error(), "is required") || strings.Contains(err.Error(), "is missing"))
}
