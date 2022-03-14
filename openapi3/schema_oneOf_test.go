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

func TestVisitJSON_OneOf_MissingDiscriptorProperty(t *testing.T) {
	s, err := NewLoader().LoadFromData(oneofSpec)
	require.NoError(t, err)
	err = s.Components.Schemas["Animal"].Value.VisitJSON(map[string]interface{}{
		"name": "snoopy",
	})
	require.EqualError(t, err, "input does not contain the discriminator property")
}

func TestVisitJSON_OneOf_MissingDiscriptorValue(t *testing.T) {
	s, err := NewLoader().LoadFromData(oneofSpec)
	require.NoError(t, err)
	err = s.Components.Schemas["Animal"].Value.VisitJSON(map[string]interface{}{
		"name":  "snoopy",
		"$type": "snake",
	})
	require.EqualError(t, err, "input does not contain a valid discriminator value")
}

func TestVisitJSON_OneOf_MissingField(t *testing.T) {
	s, err := NewLoader().LoadFromData(oneofSpec)
	require.NoError(t, err)
	err = s.Components.Schemas["Animal"].Value.VisitJSON(map[string]interface{}{
		"name":  "snoopy",
		"$type": "dog",
	})
	require.EqualError(t, err, "Error at \"/barks\": property \"barks\" is missing\nSchema:\n  {\n    \"properties\": {\n      \"$type\": {\n        \"enum\": [\n          \"dog\"\n        ],\n        \"type\": \"string\"\n      },\n      \"barks\": {\n        \"type\": \"boolean\"\n      },\n      \"name\": {\n        \"type\": \"string\"\n      }\n    },\n    \"required\": [\n      \"name\",\n      \"barks\",\n      \"$type\"\n    ],\n    \"type\": \"object\"\n  }\n\nValue:\n  {\n    \"$type\": \"dog\",\n    \"name\": \"snoopy\"\n  }\n")
}

func TestVisitJSON_OneOf_NoDiscriptor_MissingField(t *testing.T) {
	s, err := NewLoader().LoadFromData(oneofNoDiscriminatorSpec)
	require.NoError(t, err)
	err = s.Components.Schemas["Animal"].Value.VisitJSON(map[string]interface{}{
		"name": "snoopy",
	})
	require.EqualError(t, err, "doesn't match schema due to: Error at \"/scratches\": property \"scratches\" is missing\nSchema:\n  {\n    \"properties\": {\n      \"name\": {\n        \"type\": \"string\"\n      },\n      \"scratches\": {\n        \"type\": \"boolean\"\n      }\n    },\n    \"required\": [\n      \"name\",\n      \"scratches\"\n    ],\n    \"type\": \"object\"\n  }\n\nValue:\n  {\n    \"name\": \"snoopy\"\n  }\n Or Error at \"/barks\": property \"barks\" is missing\nSchema:\n  {\n    \"properties\": {\n      \"barks\": {\n        \"type\": \"boolean\"\n      },\n      \"name\": {\n        \"type\": \"string\"\n      }\n    },\n    \"required\": [\n      \"name\",\n      \"barks\"\n    ],\n    \"type\": \"object\"\n  }\n\nValue:\n  {\n    \"name\": \"snoopy\"\n  }\n")
}

func TestVisitJSON_OneOf_BadDescriminatorType(t *testing.T) {
	s, err := NewLoader().LoadFromData(oneofSpec)
	require.NoError(t, err)
	err = s.Components.Schemas["Animal"].Value.VisitJSON(map[string]interface{}{
		"name":      "snoopy",
		"scratches": true,
		"$type":     1,
	})
	require.EqualError(t, err, "descriminator value is not a string")

	err = s.Components.Schemas["Animal"].Value.VisitJSON(map[string]interface{}{
		"name":  "snoopy",
		"barks": true,
		"$type": nil,
	})
	require.EqualError(t, err, "descriminator value is not a string")
}
