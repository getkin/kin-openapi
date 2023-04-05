package openapi3

import (
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

	require.Equal(t, 2, len(doc.Components.Schemas["MyResponseType"].Value.Discriminator.Mapping))
}
