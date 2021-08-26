package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var oneofSpec = []byte(`components:
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
`)

var oneofNoDiscriminatorSpec = []byte(`components:
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
`)

func TestVisitData_OneOf_MissingDiscriptorProperty(t *testing.T) {
	s, err := NewLoader().LoadFromData(oneofSpec)
	require.NoError(t, err)
	err = s.Components.Schemas["Animal"].Value.VisitData(nil, map[string]interface{}{
		"name": "snoopy",
	})
	require.EqualError(t, err, "input does not contain the discriminator property")
}

func TestVisitData_OneOf_MissingDiscriptorValue(t *testing.T) {
	s, err := NewLoader().LoadFromData(oneofSpec)
	require.NoError(t, err)
	err = s.Components.Schemas["Animal"].Value.VisitData(nil, map[string]interface{}{
		"name":  "snoopy",
		"$type": "snake",
	})
	require.EqualError(t, err, "input does not contain a valid discriminator value")
}

func TestVisitData_OneOf_MissingField(t *testing.T) {
	s, err := NewLoader().LoadFromData(oneofSpec)
	require.NoError(t, err)
	err = s.Components.Schemas["Animal"].Value.VisitData(nil, map[string]interface{}{
		"name":  "snoopy",
		"$type": "dog",
	})
	require.EqualError(t, err, `Error at "/barks": property "barks" is missing
Schema:
  {
    "properties": {
      "$type": {
        "enum": [
          "dog"
        ],
        "type": "string"
      },
      "barks": {
        "type": "boolean"
      },
      "name": {
        "type": "string"
      }
    },
    "required": [
      "name",
      "barks",
      "$type"
    ],
    "type": "object"
  }

Value:
  {
    "$type": "dog",
    "name": "snoopy"
  }
`)
}

func TestVisitData_OneOf_NoDiscriptor_MissingField(t *testing.T) {
	s, err := NewLoader().LoadFromData(oneofNoDiscriminatorSpec)
	require.NoError(t, err)
	err = s.Components.Schemas["Animal"].Value.VisitData(nil, map[string]interface{}{
		"name": "snoopy",
	})
	require.EqualError(t, err, `doesn't match schema due to: Error at "/scratches": property "scratches" is missing
Schema:
  {
    "properties": {
      "name": {
        "type": "string"
      },
      "scratches": {
        "type": "boolean"
      }
    },
    "required": [
      "name",
      "scratches"
    ],
    "type": "object"
  }

Value:
  {
    "name": "snoopy"
  }
 Or Error at "/barks": property "barks" is missing
Schema:
  {
    "properties": {
      "barks": {
        "type": "boolean"
      },
      "name": {
        "type": "string"
      }
    },
    "required": [
      "name",
      "barks"
    ],
    "type": "object"
  }

Value:
  {
    "name": "snoopy"
  }
`)
}
