package openapi3

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParsingDiscriminator(t *testing.T) {
	const spec = `
{
	"openapi": "3.0.0",
	"info": {
		"version": "1.0.0",
		"title": "title",
		"description": "desc",
		"contact": {
			"email": "email"
		}
	},
	"paths": {},
	"components": {
		"schemas": {
			"MyResponseType": {
				"discriminator": {
					"mapping": {
						"cat": "#/components/schemas/Cat",
						"dog": "#/components/schemas/Dog"
					},
					"propertyName": "pet_type"
				},
				"oneOf": [
					{
						"$ref": "#/components/schemas/Cat"
					},
					{
						"$ref": "#/components/schemas/Dog"
					}
				]
			},
			"Cat": {"enum": ["chat"]},
			"Dog": {"enum": ["chien"]}
		}
	}
}
`

	loader := NewLoader()
	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	discriminatorMap, ok := doc.Components.Schemas["MyResponseType"].Value.Discriminator.(map[string]interface{})
	require.True(t, ok)

	marshaledDiscriminator, err := json.Marshal(discriminatorMap)
	require.NoError(t, err)

	var discriminator *Discriminator
	err = json.Unmarshal(marshaledDiscriminator, &discriminator)
	require.NoError(t, err)

	require.Len(t, discriminator.Mapping, 2)
}
