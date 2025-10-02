package openapi3

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

var ctx = context.Background()

func TestDocumentVersionDetection(t *testing.T) {
	t.Run("IsOpenAPI3_0", func(t *testing.T) {
		doc := &T{OpenAPI: "3.0.0"}
		require.True(t, doc.IsOpenAPI3_0())
		require.False(t, doc.IsOpenAPI3_1())

		doc = &T{OpenAPI: "3.0.3"}
		require.True(t, doc.IsOpenAPI3_0())
		require.False(t, doc.IsOpenAPI3_1())

		doc = &T{OpenAPI: "3.0.1"}
		require.True(t, doc.IsOpenAPI3_0())
	})

	t.Run("IsOpenAPI3_1", func(t *testing.T) {
		doc := &T{OpenAPI: "3.1.0"}
		require.True(t, doc.IsOpenAPI3_1())
		require.False(t, doc.IsOpenAPI3_0())

		doc = &T{OpenAPI: "3.1.1"}
		require.True(t, doc.IsOpenAPI3_1())
		require.False(t, doc.IsOpenAPI3_0())
	})

	t.Run("Version", func(t *testing.T) {
		doc := &T{OpenAPI: "3.0.3"}
		require.Equal(t, "3.0", doc.Version())

		doc = &T{OpenAPI: "3.1.0"}
		require.Equal(t, "3.1", doc.Version())

		doc = &T{OpenAPI: "3.1"}
		require.Equal(t, "3.1", doc.Version())
	})

	t.Run("nil or empty document", func(t *testing.T) {
		var doc *T
		require.False(t, doc.IsOpenAPI3_0())
		require.False(t, doc.IsOpenAPI3_1())
		require.Equal(t, "", doc.Version())

		doc = &T{}
		require.False(t, doc.IsOpenAPI3_0())
		require.False(t, doc.IsOpenAPI3_1())
		require.Equal(t, "", doc.Version())
	})
}

func TestWebhooksField(t *testing.T) {
	t.Run("serialize webhooks in OpenAPI 3.1", func(t *testing.T) {
		doc := &T{
			OpenAPI: "3.1.0",
			Info: &Info{
				Title:   "Test API",
				Version: "1.0.0",
			},
			Paths: NewPaths(),
			Webhooks: map[string]*PathItem{
				"newPet": {
					Post: &Operation{
						Summary: "New pet webhook",
						Responses: NewResponses(
							WithStatus(200, &ResponseRef{
								Value: &Response{
									Description: Ptr("Success"),
								},
							}),
						),
					},
				},
			},
		}

		data, err := json.Marshal(doc)
		require.NoError(t, err)

		// Should contain webhooks
		require.Contains(t, string(data), `"webhooks"`)
		require.Contains(t, string(data), `"newPet"`)
	})

	t.Run("deserialize webhooks from OpenAPI 3.1", func(t *testing.T) {
		jsonData := []byte(`{
			"openapi": "3.1.0",
			"info": {
				"title": "Test API",
				"version": "1.0.0"
			},
			"paths": {},
			"webhooks": {
				"newPet": {
					"post": {
						"summary": "New pet webhook",
						"responses": {
							"200": {
								"description": "Success"
							}
						}
					}
				}
			}
		}`)

		var doc T
		err := json.Unmarshal(jsonData, &doc)
		require.NoError(t, err)

		require.True(t, doc.IsOpenAPI3_1())
		require.NotNil(t, doc.Webhooks)
		require.Contains(t, doc.Webhooks, "newPet")
		require.NotNil(t, doc.Webhooks["newPet"].Post)
		require.Equal(t, "New pet webhook", doc.Webhooks["newPet"].Post.Summary)
	})

	t.Run("OpenAPI 3.0 without webhooks", func(t *testing.T) {
		jsonData := []byte(`{
			"openapi": "3.0.3",
			"info": {
				"title": "Test API",
				"version": "1.0.0"
			},
			"paths": {}
		}`)

		var doc T
		err := json.Unmarshal(jsonData, &doc)
		require.NoError(t, err)

		require.True(t, doc.IsOpenAPI3_0())
		require.Nil(t, doc.Webhooks)
	})

	t.Run("validate webhooks", func(t *testing.T) {
		doc := &T{
			OpenAPI: "3.1.0",
			Info: &Info{
				Title:   "Test API",
				Version: "1.0.0",
			},
			Paths: NewPaths(),
			Webhooks: map[string]*PathItem{
				"validWebhook": {
					Post: &Operation{
						Responses: NewResponses(
							WithStatus(200, &ResponseRef{
								Value: &Response{
									Description: Ptr("Success"),
								},
							}),
						),
					},
				},
			},
		}

		// Should validate successfully
		err := doc.Validate(ctx)
		require.NoError(t, err)
	})

	t.Run("validate fails with nil webhook", func(t *testing.T) {
		doc := &T{
			OpenAPI: "3.1.0",
			Info: &Info{
				Title:   "Test API",
				Version: "1.0.0",
			},
			Paths: NewPaths(),
			Webhooks: map[string]*PathItem{
				"invalidWebhook": nil,
			},
		}

		err := doc.Validate(ctx)
		require.Error(t, err)
		require.Contains(t, err.Error(), "webhook")
		require.Contains(t, err.Error(), "invalidWebhook")
	})
}

func TestJSONLookupWithWebhooks(t *testing.T) {
	doc := &T{
		OpenAPI: "3.1.0",
		Info: &Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Paths: NewPaths(),
		Webhooks: map[string]*PathItem{
			"test": {
				Post: &Operation{
					Summary: "Test webhook",
				},
			},
		},
	}

	result, err := doc.JSONLookup("webhooks")
	require.NoError(t, err)
	require.NotNil(t, result)

	webhooks, ok := result.(map[string]*PathItem)
	require.True(t, ok)
	require.Contains(t, webhooks, "test")
}

func TestVersionBasedBehavior(t *testing.T) {
	t.Run("detect and handle OpenAPI 3.0", func(t *testing.T) {
		doc := &T{
			OpenAPI: "3.0.3",
			Info: &Info{
				Title:   "Test API",
				Version: "1.0.0",
			},
			Paths: NewPaths(),
		}

		if doc.IsOpenAPI3_0() {
			// OpenAPI 3.0 specific logic
			require.Nil(t, doc.Webhooks)
		}
	})

	t.Run("detect and handle OpenAPI 3.1", func(t *testing.T) {
		doc := &T{
			OpenAPI: "3.1.0",
			Info: &Info{
				Title:   "Test API",
				Version: "1.0.0",
			},
			Paths: NewPaths(),
			Webhooks: map[string]*PathItem{
				"test": {
					Post: &Operation{
						Summary: "Test",
						Responses: NewResponses(
							WithStatus(200, &ResponseRef{
								Value: &Response{
									Description: Ptr("OK"),
								},
							}),
						),
					},
				},
			},
		}

		if doc.IsOpenAPI3_1() {
			// OpenAPI 3.1 specific logic
			require.NotNil(t, doc.Webhooks)
			require.Contains(t, doc.Webhooks, "test")
		}
	})
}

func TestMigrationScenario(t *testing.T) {
	t.Run("upgrade document from 3.0 to 3.1", func(t *testing.T) {
		// Start with 3.0 document
		doc := &T{
			OpenAPI: "3.0.3",
			Info: &Info{
				Title:   "Test API",
				Version: "1.0.0",
			},
			Paths: NewPaths(),
		}

		require.True(t, doc.IsOpenAPI3_0())
		require.Nil(t, doc.Webhooks)

		// Upgrade to 3.1
		doc.OpenAPI = "3.1.0"

		// Add 3.1 features
		doc.Webhooks = map[string]*PathItem{
			"newEvent": {
				Post: &Operation{
					Summary: "New event notification",
					Responses: NewResponses(
						WithStatus(200, &ResponseRef{
							Value: &Response{
								Description: Ptr("Processed"),
							},
						}),
					),
				},
			},
		}

		require.True(t, doc.IsOpenAPI3_1())
		require.NotNil(t, doc.Webhooks)

		// Validate the upgraded document
		err := doc.Validate(ctx)
		require.NoError(t, err)
	})
}