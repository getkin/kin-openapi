package openapi2

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue1010(t *testing.T) {
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
                            "$ref": "#/definitions/Pet"
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
            },
            "discriminator": "petType"
        },
        "Dog": {
            "allOf": [
                {
                    "$ref": "#/definitions/Pet"
                },
                {
                    "type": "object",
                    "properties": {
                        "breed": {
                            "type": "string"
                        }
                    }
                }
            ]
        },
        "Cat": {
            "allOf": [
                {
                    "$ref": "#/definitions/Pet"
                },
                {
                    "type": "object",
                    "properties": {
                        "color": {
                            "type": "string"
                        }
                    }
                }
            ]
        }
    }
}
`)

	var doc2 T
	err := json.Unmarshal(v2, &doc2)
	require.NoError(t, err)
	require.Equal(t, "petType", doc2.Definitions["Pet"].Value.Discriminator)
}
