package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var jsonSpecWithDiscriminator = []byte(`
{
	"openapi": "3.0.0",
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
`)

func TestParsingDiscriminator(t *testing.T) {
	loader, err := NewSwaggerLoader().LoadSwaggerFromData(jsonSpecWithDiscriminator)
	require.NoError(t, err)
	require.Equal(t, 2, len(loader.Components.Schemas["MyResponseType"].Value.OneOf))
}
