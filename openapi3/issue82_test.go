package openapi3

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue82(t *testing.T) {
	payload := map[string]interface{}{
		"prop1": "val",
		"prop3": "val",
	}

	schemas := []string{`
{
	"type": "object",
	"additionalProperties": false,
	"required": ["prop1"],
	"properties": {
		"prop1": {
			"type": "string"
		}
	}
}`, `{
	"anyOf": [
		{
			"type": "object",
			"additionalProperties": false,
			"required": ["prop1"],
			"properties": {
				"prop1": {
					"type": "string"
				}
			}
		},
		{
			"type": "object",
			"additionalProperties": false,
			"properties": {
				"prop2": {
					"type": "string"
				}
			}
		}
	]
}`, `{
		"oneOf": [
			{
				"type": "object",
				"additionalProperties": false,
				"required": ["prop1"],
				"properties": {
					"prop1": {
						"type": "string"
					}
				}
			},
			{
				"type": "object",
				"additionalProperties": false,
				"properties": {
					"prop2": {
						"type": "string"
					}
				}
			}
		]
}`, `{
		"allOf": [
			{
				"type": "object",
				"additionalProperties": false,
				"required": ["prop1"],
				"properties": {
					"prop1": {
						"type": "string"
					}
				}
			},
			{
				"type": "object",
				"additionalProperties": false,
				"properties": {
					"prop2": {
						"type": "string"
					}
				}
			}
		]
	}
`}

	for _, jsonSchema := range schemas {
		var dataSchema Schema
		err := json.Unmarshal([]byte(jsonSchema), &dataSchema)
		require.NoError(t, err)

		err = dataSchema.VisitJSON(payload)
		require.Error(t, err)
	}
}
