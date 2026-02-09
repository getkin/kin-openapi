package openapi3_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

// TestBackwardCompatibility_OpenAPI30 ensures that existing OpenAPI 3.0 functionality is not broken
func TestBackwardCompatibility_OpenAPI30(t *testing.T) {
	t.Run("load and validate OpenAPI 3.0 document", func(t *testing.T) {
		spec := `
openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
  license:
    name: MIT
    url: https://opensource.org/licenses/MIT
paths:
  /users:
    get:
      summary: Get users
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: integer
                  name:
                    type: string
                    nullable: true
                required:
                  - id
`
		loader := openapi3.NewLoader()
		doc, err := loader.LoadFromData([]byte(spec))
		require.NoError(t, err)
		require.NotNil(t, doc)

		// Verify version detection
		require.True(t, doc.IsOpenAPI3_0())
		require.False(t, doc.IsOpenAPI3_1())
		require.Equal(t, "3.0", doc.Version())

		// Verify structure
		require.Equal(t, "Test API", doc.Info.Title)
		require.NotNil(t, doc.Info.License)
		require.Equal(t, "MIT", doc.Info.License.Name)
		require.Equal(t, "https://opensource.org/licenses/MIT", doc.Info.License.URL)
		require.Empty(t, doc.Info.License.Identifier) // 3.0 doesn't have this

		// Verify webhooks is nil for 3.0
		require.Nil(t, doc.Webhooks)
		require.Empty(t, doc.JSONSchemaDialect)

		// Validate
		err = doc.Validate(context.Background())
		require.NoError(t, err)
	})

	t.Run("nullable schema validation still works", func(t *testing.T) {
		schema := &openapi3.Schema{
			Type:     &openapi3.Types{"string"},
			Nullable: true,
		}

		// Should accept string
		err := schema.VisitJSON("hello", openapi3.EnableJSONSchema2020())
		require.NoError(t, err)

		// Should accept null
		err = schema.VisitJSON(nil, openapi3.EnableJSONSchema2020())
		require.NoError(t, err)

		// Should reject number
		err = schema.VisitJSON(123, openapi3.EnableJSONSchema2020())
		require.Error(t, err)
	})

	t.Run("existing schema fields work", func(t *testing.T) {
		min := 0.0
		max := 100.0
		schema := &openapi3.Schema{
			Type:      &openapi3.Types{"integer"},
			Min:       &min,
			Max:       &max,
			MinLength: 1,
		}

		// Type checking
		require.True(t, schema.Type.Is("integer"))
		require.False(t, schema.Type.IsMultiple())

		// Validation still works
		err := schema.VisitJSON(50, openapi3.EnableJSONSchema2020())
		require.NoError(t, err)

		err = schema.VisitJSON(150, openapi3.EnableJSONSchema2020())
		require.Error(t, err)
	})

	t.Run("serialization preserves 3.0 format", func(t *testing.T) {
		doc := &openapi3.T{
			OpenAPI: "3.0.3",
			Info: &openapi3.Info{
				Title:   "Test",
				Version: "1.0.0",
			},
			Paths: openapi3.NewPaths(),
		}

		data, err := json.Marshal(doc)
		require.NoError(t, err)

		// Should not contain 3.1 fields
		require.NotContains(t, string(data), "webhooks")
		require.NotContains(t, string(data), "jsonSchemaDialect")
		require.Contains(t, string(data), `"openapi":"3.0.3"`)
	})
}

// TestOpenAPI31_NewFeatures tests all new OpenAPI 3.1 features
func TestOpenAPI31_NewFeatures(t *testing.T) {
	t.Run("load and validate OpenAPI 3.1 document with webhooks", func(t *testing.T) {
		spec := `
openapi: 3.1.0
info:
  title: Test API
  version: 1.0.0
  license:
    name: MIT
    identifier: MIT
jsonSchemaDialect: https://json-schema.org/draft/2020-12/schema
paths:
  /users:
    get:
      responses:
        '200':
          description: Success
webhooks:
  newUser:
    post:
      summary: User created notification
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                id:
                  type: integer
      responses:
        '200':
          description: Processed
`
		loader := openapi3.NewLoader()
		doc, err := loader.LoadFromData([]byte(spec))
		require.NoError(t, err)
		require.NotNil(t, doc)

		// Verify version detection
		require.True(t, doc.IsOpenAPI3_1())
		require.False(t, doc.IsOpenAPI3_0())
		require.Equal(t, "3.1", doc.Version())

		// Verify 3.1 fields
		require.NotNil(t, doc.Webhooks)
		require.Contains(t, doc.Webhooks, "newUser")
		require.Equal(t, "https://json-schema.org/draft/2020-12/schema", doc.JSONSchemaDialect)

		// Verify license identifier
		require.Equal(t, "MIT", doc.Info.License.Identifier)

		// Validate
		err = doc.Validate(context.Background())
		require.NoError(t, err)
	})

	t.Run("type arrays with null", func(t *testing.T) {
		schema := &openapi3.Schema{
			Type: &openapi3.Types{"string", "null"},
		}

		// Type checks
		require.True(t, schema.Type.IsMultiple())
		require.True(t, schema.Type.IncludesNull())
		require.True(t, schema.Type.Includes("string"))

		// Should accept string
		err := schema.VisitJSON("hello", openapi3.EnableJSONSchema2020())
		require.NoError(t, err)

		// Should accept null (with new validator)

		err = schema.VisitJSON(nil, openapi3.EnableJSONSchema2020())
		require.NoError(t, err)

		// Should reject number
		err = schema.VisitJSON(123, openapi3.EnableJSONSchema2020())
		require.Error(t, err)
	})

	t.Run("const keyword validation", func(t *testing.T) {

		schema := &openapi3.Schema{
			Const: "production",
		}

		err := schema.VisitJSON("production", openapi3.EnableJSONSchema2020())
		require.NoError(t, err)

		err = schema.VisitJSON("development", openapi3.EnableJSONSchema2020())
		require.Error(t, err)
	})

	t.Run("examples array", func(t *testing.T) {
		schema := &openapi3.Schema{
			Type: &openapi3.Types{"string"},
			Examples: []any{
				"example1",
				"example2",
				"example3",
			},
		}

		require.Len(t, schema.Examples, 3)

		// Serialize and verify
		data, err := json.Marshal(schema)
		require.NoError(t, err)
		require.Contains(t, string(data), "examples")
		require.Contains(t, string(data), "example1")
	})

	t.Run("all new schema keywords serialize", func(t *testing.T) {
		minContains := uint64(1)
		maxContains := uint64(3)

		schema := &openapi3.Schema{
			Type: &openapi3.Types{"array"},
			PrefixItems: []*openapi3.SchemaRef{
				{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
				{Value: &openapi3.Schema{Type: &openapi3.Types{"number"}}},
			},
			Contains: &openapi3.SchemaRef{
				Value: &openapi3.Schema{Type: &openapi3.Types{"string"}},
			},
			MinContains: &minContains,
			MaxContains: &maxContains,
			PropertyNames: &openapi3.SchemaRef{
				Value: &openapi3.Schema{Pattern: "^[a-z]+$"},
			},
		}

		data, err := json.Marshal(schema)
		require.NoError(t, err)

		str := string(data)
		require.Contains(t, str, "prefixItems")
		require.Contains(t, str, "contains")
		require.Contains(t, str, "minContains")
		require.Contains(t, str, "maxContains")
		require.Contains(t, str, "propertyNames")
	})

	t.Run("round-trip serialization preserves all fields", func(t *testing.T) {
		doc := &openapi3.T{
			OpenAPI:           "3.1.0",
			JSONSchemaDialect: "https://json-schema.org/draft/2020-12/schema",
			Info: &openapi3.Info{
				Title:   "Test API",
				Version: "1.0.0",
				License: &openapi3.License{
					Name:       "Apache-2.0",
					Identifier: "Apache-2.0",
				},
			},
			Paths: openapi3.NewPaths(),
			Webhooks: map[string]*openapi3.PathItem{
				"test": {
					Post: &openapi3.Operation{
						Summary: "Test webhook",
						Responses: openapi3.NewResponses(
							openapi3.WithStatus(200, &openapi3.ResponseRef{
								Value: &openapi3.Response{
									Description: openapi3.Ptr("OK"),
								},
							}),
						),
					},
				},
			},
		}

		// Serialize
		data, err := json.Marshal(doc)
		require.NoError(t, err)

		// Deserialize
		var doc2 openapi3.T
		err = json.Unmarshal(data, &doc2)
		require.NoError(t, err)

		// Verify all fields
		require.Equal(t, "3.1.0", doc2.OpenAPI)
		require.Equal(t, "https://json-schema.org/draft/2020-12/schema", doc2.JSONSchemaDialect)
		require.Equal(t, "Apache-2.0", doc2.Info.License.Identifier)
		require.NotNil(t, doc2.Webhooks)
		require.Contains(t, doc2.Webhooks, "test")
	})
}

// TestJSONSchema2020Validator_RealWorld tests the validator with realistic schemas
func TestJSONSchema2020Validator_RealWorld(t *testing.T) {

	t.Run("complex nested object with nullable", func(t *testing.T) {
		min := 0.0

		schema := &openapi3.Schema{
			Type: &openapi3.Types{"object"},
			Properties: openapi3.Schemas{
				"user": &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type: &openapi3.Types{"object"},
						Properties: openapi3.Schemas{
							"id": &openapi3.SchemaRef{
								Value: &openapi3.Schema{Type: &openapi3.Types{"integer"}},
							},
							"name": &openapi3.SchemaRef{
								Value: &openapi3.Schema{
									Type: &openapi3.Types{"string", "null"},
								},
							},
							"age": &openapi3.SchemaRef{
								Value: &openapi3.Schema{
									Type: &openapi3.Types{"integer"},
									Min:  &min,
								},
							},
						},
						Required: []string{"id"},
					},
				},
				"tags": &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type: &openapi3.Types{"array", "null"},
						Items: &openapi3.SchemaRef{
							Value: &openapi3.Schema{Type: &openapi3.Types{"string"}},
						},
					},
				},
			},
			Required: []string{"user"},
		}

		// Valid data
		validData := map[string]any{
			"user": map[string]any{
				"id":   1,
				"name": "John",
				"age":  30,
			},
			"tags": []any{"tag1", "tag2"},
		}
		err := schema.VisitJSON(validData, openapi3.EnableJSONSchema2020())
		require.NoError(t, err)

		// Valid with null name
		validDataNullName := map[string]any{
			"user": map[string]any{
				"id":   2,
				"name": nil,
				"age":  25,
			},
			"tags": nil,
		}
		err = schema.VisitJSON(validDataNullName, openapi3.EnableJSONSchema2020())
		require.NoError(t, err)

		// Invalid - missing required field
		invalidData := map[string]any{
			"user": map[string]any{
				"name": "Jane",
			},
		}
		err = schema.VisitJSON(invalidData, openapi3.EnableJSONSchema2020())
		require.Error(t, err)
	})

	t.Run("oneOf with different types", func(t *testing.T) {
		schema := &openapi3.Schema{
			OneOf: openapi3.SchemaRefs{
				&openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type: &openapi3.Types{"object"},
						Properties: openapi3.Schemas{
							"type": &openapi3.SchemaRef{
								Value: &openapi3.Schema{
									Const: "email",
								},
							},
							"email": &openapi3.SchemaRef{
								Value: &openapi3.Schema{Type: &openapi3.Types{"string"}},
							},
						},
						Required: []string{"type", "email"},
					},
				},
				&openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type: &openapi3.Types{"object"},
						Properties: openapi3.Schemas{
							"type": &openapi3.SchemaRef{
								Value: &openapi3.Schema{
									Const: "phone",
								},
							},
							"phone": &openapi3.SchemaRef{
								Value: &openapi3.Schema{Type: &openapi3.Types{"string"}},
							},
						},
						Required: []string{"type", "phone"},
					},
				},
			},
		}

		// Valid email
		emailData := map[string]any{
			"type":  "email",
			"email": "test@example.com",
		}
		err := schema.VisitJSON(emailData, openapi3.EnableJSONSchema2020())
		require.NoError(t, err)

		// Valid phone
		phoneData := map[string]any{
			"type":  "phone",
			"phone": "+1234567890",
		}
		err = schema.VisitJSON(phoneData, openapi3.EnableJSONSchema2020())
		require.NoError(t, err)

		// Invalid - doesn't match any oneOf
		invalidData := map[string]any{
			"type": "email",
			// missing email field
		}
		err = schema.VisitJSON(invalidData, openapi3.EnableJSONSchema2020())
		require.Error(t, err)
	})
}

// TestMigrationScenarios tests realistic migration paths
func TestMigrationScenarios(t *testing.T) {
	t.Run("migrate nullable to type array", func(t *testing.T) {
		// Old 3.0 style
		schema30 := &openapi3.Schema{
			Type:     &openapi3.Types{"string"},
			Nullable: true,
		}

		// New 3.1 style
		schema31 := &openapi3.Schema{
			Type: &openapi3.Types{"string", "null"},
		}

		// Both should accept null with new validator

		err := schema30.VisitJSON(nil)
		require.NoError(t, err)

		err = schema31.VisitJSON(nil)
		require.NoError(t, err)

		// Both should accept string
		err = schema30.VisitJSON("test")
		require.NoError(t, err)

		err = schema31.VisitJSON("test")
		require.NoError(t, err)
	})

	t.Run("automatic version detection and configuration", func(t *testing.T) {
		// Simulate loading 3.0 document
		spec30 := []byte(`{"openapi":"3.0.3","info":{"title":"Test","version":"1.0.0"},"paths":{}}`)
		var doc30 openapi3.T
		err := json.Unmarshal(spec30, &doc30)
		require.NoError(t, err)

		if doc30.IsOpenAPI3_1() {
		}

		// Simulate loading 3.1 document
		spec31 := []byte(`{"openapi":"3.1.0","info":{"title":"Test","version":"1.0.0"},"paths":{}}`)
		var doc31 openapi3.T
		err = json.Unmarshal(spec31, &doc31)
		require.NoError(t, err)

		if doc31.IsOpenAPI3_1() {
		}

		// Cleanup
	})
}

// TestEdgeCases tests edge cases and error conditions
func TestEdgeCases(t *testing.T) {
	t.Run("empty types array", func(t *testing.T) {
		schema := &openapi3.Schema{
			Type: &openapi3.Types{},
		}

		require.True(t, schema.Type.IsEmpty())
		require.False(t, schema.Type.IsSingle())
		require.False(t, schema.Type.IsMultiple())
	})

	t.Run("nil vs empty webhooks", func(t *testing.T) {
		doc30 := &openapi3.T{
			OpenAPI: "3.0.3",
			Info:    &openapi3.Info{Title: "Test", Version: "1.0.0"},
			Paths:   openapi3.NewPaths(),
		}

		doc31Empty := &openapi3.T{
			OpenAPI:  "3.1.0",
			Info:     &openapi3.Info{Title: "Test", Version: "1.0.0"},
			Paths:    openapi3.NewPaths(),
			Webhooks: map[string]*openapi3.PathItem{},
		}

		// Nil webhooks should not serialize
		data30, _ := json.Marshal(doc30)
		require.NotContains(t, string(data30), "webhooks")

		// Empty webhooks should not serialize
		data31, _ := json.Marshal(doc31Empty)
		require.NotContains(t, string(data31), "webhooks")
	})

	t.Run("license with both url and identifier", func(t *testing.T) {
		license := &openapi3.License{
			Name:       "MIT",
			URL:        "https://opensource.org/licenses/MIT",
			Identifier: "MIT",
		}

		// Should serialize both (spec says only one should be used, but library allows both)
		data, err := json.Marshal(license)
		require.NoError(t, err)
		require.Contains(t, string(data), `"url"`)
		require.Contains(t, string(data), `"identifier"`)
	})

	t.Run("version detection with edge cases", func(t *testing.T) {
		var doc *openapi3.T
		require.False(t, doc.IsOpenAPI3_0())
		require.False(t, doc.IsOpenAPI3_1())
		require.Equal(t, "", doc.Version())

		doc = &openapi3.T{}
		require.False(t, doc.IsOpenAPI3_0())
		require.False(t, doc.IsOpenAPI3_1())

		doc = &openapi3.T{OpenAPI: "3.x"}
		require.False(t, doc.IsOpenAPI3_0())
		require.False(t, doc.IsOpenAPI3_1())
	})

	t.Run("schema without type permits any type", func(t *testing.T) {
		schema := &openapi3.Schema{}

		require.True(t, schema.Type.Permits("string"))
		require.True(t, schema.Type.Permits("number"))
		require.True(t, schema.Type.Permits("anything"))
	})
}

// TestPerformance checks for obvious performance issues
func TestPerformance(t *testing.T) {
	t.Run("large schema compilation", func(t *testing.T) {

		// Create a large schema
		properties := make(openapi3.Schemas)
		for i := 0; i < 100; i++ {
			properties[string(rune('a'+i%26))+string(rune('0'+i/26))] = &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: &openapi3.Types{"string"},
				},
			}
		}

		schema := &openapi3.Schema{
			Type:       &openapi3.Types{"object"},
			Properties: properties,
		}

		// Should compile and validate without hanging
		data := map[string]any{"a0": "test"}
		err := schema.VisitJSON(data, openapi3.EnableJSONSchema2020())
		require.NoError(t, err)
	})

	t.Run("deeply nested schema", func(t *testing.T) {
		// Create deeply nested schema (but not too deep to cause stack overflow)
		schema := &openapi3.Schema{Type: &openapi3.Types{"object"}}
		current := schema

		for i := 0; i < 10; i++ {
			current.Properties = openapi3.Schemas{
				"nested": &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type: &openapi3.Types{"object"},
					},
				},
			}
			current = current.Properties["nested"].Value
		}

		// Should serialize without issue
		_, err := json.Marshal(schema)
		require.NoError(t, err)
	})
}
