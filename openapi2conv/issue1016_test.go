package openapi2conv

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/stretchr/testify/require"
)

func TestIssue1016(t *testing.T) {
	v2 := []byte(`
{
    "basePath": "/v2",
    "host": "test.example.com",
    "info": {
        "title": "MyAPI",
        "version": "0.1",
        "x-info": "info extension"
    },
    "paths": {
        "/foo": {
            "get": {
                "operationId": "getFoo",
                "responses": {
                    "200": {
                        "description": "returns all information",
                        "schema": {
                            "$ref": "#/definitions/PetDirectory"
                        }
                    },
                    "default": {
                        "description": "OK"
                    }
                },
                "summary": "get foo"
            }
        }
    },
    "schemes": [
        "http"
    ],
    "swagger": "2.0",
    "definitions": {
		"Pet": {
			"type": "object",
			"required": ["petType"],
			"properties": {
				"petType": {
					"type": "string"
				},
				"name": {
					"type": "string"
				},
				"age": {
					"type": "integer"
				}
			}
		},
		"PetDirectory":{
			"type": "object",
			"additionalProperties": {
				"$ref": "#/definitions/Pet"
			}
		}
    }
}
`)

	var doc2 openapi2.T
	err := json.Unmarshal(v2, &doc2)
	require.NoError(t, err)

	doc3, err := v2v3YAML(v2)
	require.NoError(t, err)

	err = doc3.Validate(context.Background())
	require.NoError(t, err)
	require.Equal(t, "#/components/schemas/Pet", doc3.Components.Schemas["PetDirectory"].Value.AdditionalProperties.Schema.Ref)
}
