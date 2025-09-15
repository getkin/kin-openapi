package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMapping(t *testing.T) {

	schema := `
openapi: 3.0.0
info:
  title: ACME
  version: 1.0.0
components:
  schemas:
    Pet:
      type: object
      required:
      - petType
      properties:
        petType:
          type: string
      discriminator:
        propertyName: petType
        mapping:
          dog: Dog
    Cat:
      allOf:
      - $ref: '#/components/schemas/Pet'
      - type: object
        # all other properties specific to a Cat
        properties:
          name:
            type: string
    Dog:
      allOf:
      - $ref: '#/components/schemas/Pet'
      - type: object
        # all other properties specific to a Dog
        properties:
          bark:
            type: string
    Lizard:
      allOf:
      - $ref: '#/components/schemas/Pet'
      - type: object
        # all other properties specific to a Lizard
        properties:
          lovesRocks:
            type: boolean
`
	loader := NewLoader()
	loader.IsExternalRefsAllowed = true
	_, err := loader.LoadFromData([]byte(schema))
	require.NoError(t, err)
}
