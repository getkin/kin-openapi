package openapi2conv

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi2"
)

func TestIssue1091_PropertyExtensions(t *testing.T) {
	// Create a v2 schema with x-order extensions on properties
	v2SchemaJSON := `{
		"type": "object",
		"properties": {
			"field1": {
				"type": "string",
				"x-order": 1
			},
			"field2": {
				"type": "integer",
				"x-order": 2
			},
			"field3": {
				"type": "object",
				"properties": {
					"nestedField": {
						"type": "string",
						"x-order": 10
					}
				},
				"x-order": 3
			}
		}
	}`

	var v2Schema openapi2.Schema
	err := json.Unmarshal([]byte(v2SchemaJSON), &v2Schema)
	require.NoError(t, err)

	v2SchemaRef := &openapi2.SchemaRef{
		Value: &v2Schema,
	}

	// Convert to v3
	v3SchemaRef := ToV3SchemaRef(v2SchemaRef)

	// Verify that the conversion was successful
	require.NotNil(t, v3SchemaRef)
	require.NotNil(t, v3SchemaRef.Value)
	require.NotNil(t, v3SchemaRef.Value.Properties)

	// Verify that extensions are preserved on properties
	field1 := v3SchemaRef.Value.Properties["field1"]
	require.NotNil(t, field1.Value)
	require.NotNil(t, field1.Value.Extensions)
	require.Equal(t, float64(1), field1.Value.Extensions["x-order"])

	field2 := v3SchemaRef.Value.Properties["field2"]
	require.NotNil(t, field2.Value)
	require.NotNil(t, field2.Value.Extensions)
	require.Equal(t, float64(2), field2.Value.Extensions["x-order"])

	// Verify nested properties also preserve extensions
	field3 := v3SchemaRef.Value.Properties["field3"]
	require.NotNil(t, field3.Value)
	require.NotNil(t, field3.Value.Extensions)
	require.Equal(t, float64(3), field3.Value.Extensions["x-order"])

	nestedField := field3.Value.Properties["nestedField"]
	require.NotNil(t, nestedField.Value)
	require.NotNil(t, nestedField.Value.Extensions)
	require.Equal(t, float64(10), nestedField.Value.Extensions["x-order"])
}

func TestIssue1091_SchemaLevelExtensions(t *testing.T) {
	// Create a v2 schema with schema-level extensions
	v2SchemaJSON := `{
		"type": "object",
		"x-schema-level": "test-value",
		"properties": {
			"field1": {
				"type": "string"
			}
		}
	}`

	var v2Schema openapi2.Schema
	err := json.Unmarshal([]byte(v2SchemaJSON), &v2Schema)
	require.NoError(t, err)

	v2SchemaRef := &openapi2.SchemaRef{
		Value:      &v2Schema,
		Extensions: map[string]interface{}{"x-ref-level": "ref-value"},
	}

	// Convert to v3
	v3SchemaRef := ToV3SchemaRef(v2SchemaRef)

	// Verify that the conversion was successful
	require.NotNil(t, v3SchemaRef)
	require.NotNil(t, v3SchemaRef.Value)

	// Verify that schema-level extensions are preserved
	require.NotNil(t, v3SchemaRef.Value.Extensions)
	require.Equal(t, "test-value", v3SchemaRef.Value.Extensions["x-schema-level"])

	// Verify that ref-level extensions are preserved
	require.NotNil(t, v3SchemaRef.Extensions)
	require.Equal(t, "ref-value", v3SchemaRef.Extensions["x-ref-level"])
}

func TestIssue1091_CompleteV2ToV3Conversion(t *testing.T) {
	// Create a complete v2 spec with extensions in definitions
	v2SpecJSON := `{
		"swagger": "2.0",
		"info": {
			"title": "Test API",
			"version": "1.0.0"
		},
		"host": "api.example.com",
		"basePath": "/v1",
		"schemes": ["https"],
		"definitions": {
			"User": {
				"type": "object",
				"x-model-type": "entity",
				"properties": {
					"id": {
						"type": "integer",
						"x-order": 1,
						"x-primary-key": true
					},
					"name": {
						"type": "string",
						"x-order": 2
					},
					"email": {
						"type": "string",
						"x-order": 3,
						"x-sensitive": true
					}
				}
			}
		},
		"paths": {
			"/users": {
				"get": {
					"responses": {
						"200": {
							"description": "List of users",
							"schema": {
								"type": "array",
								"items": {
									"$ref": "#/definitions/User"
								}
							}
						}
					}
				}
			}
		}
	}`

	var v2Doc openapi2.T
	err := json.Unmarshal([]byte(v2SpecJSON), &v2Doc)
	require.NoError(t, err)

	// Convert to v3
	v3Doc, err := ToV3(&v2Doc)
	require.NoError(t, err)

	// Verify that User schema has extensions preserved
	userSchema := v3Doc.Components.Schemas["User"]
	require.NotNil(t, userSchema)
	require.NotNil(t, userSchema.Value)
	require.NotNil(t, userSchema.Value.Extensions)
	require.Equal(t, "entity", userSchema.Value.Extensions["x-model-type"])

	// Verify that property-level extensions are preserved
	idProp := userSchema.Value.Properties["id"]
	require.NotNil(t, idProp.Value)
	require.NotNil(t, idProp.Value.Extensions)
	require.Equal(t, float64(1), idProp.Value.Extensions["x-order"])
	require.Equal(t, true, idProp.Value.Extensions["x-primary-key"])

	nameProp := userSchema.Value.Properties["name"]
	require.NotNil(t, nameProp.Value)
	require.NotNil(t, nameProp.Value.Extensions)
	require.Equal(t, float64(2), nameProp.Value.Extensions["x-order"])

	emailProp := userSchema.Value.Properties["email"]
	require.NotNil(t, emailProp.Value)
	require.NotNil(t, emailProp.Value.Extensions)
	require.Equal(t, float64(3), emailProp.Value.Extensions["x-order"])
	require.Equal(t, true, emailProp.Value.Extensions["x-sensitive"])
}
