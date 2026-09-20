package openapi3_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestWebhooksField(t *testing.T) {
	t.Run("serialize webhooks in OpenAPI 3.1", func(t *testing.T) {
		doc := &openapi3.T{
			OpenAPI: "3.1.0",
			Info: &openapi3.Info{
				Title:   "Test API",
				Version: "1.0.0",
			},
			Paths: openapi3.NewPaths(),
			Webhooks: map[string]*openapi3.PathItem{
				"newPet": {
					Post: &openapi3.Operation{
						Summary: "New pet webhook",
						Responses: openapi3.NewResponses(
							openapi3.WithStatus(200, &openapi3.ResponseRef{
								Value: &openapi3.Response{
									Description: openapi3.Ptr("Success"),
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

		var doc openapi3.T
		err := json.Unmarshal(jsonData, &doc)
		require.NoError(t, err)

		require.True(t, doc.IsOpenAPI31OrLater())
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

		var doc openapi3.T
		err := json.Unmarshal(jsonData, &doc)
		require.NoError(t, err)

		require.True(t, doc.IsOpenAPI30())
		require.Nil(t, doc.Webhooks)
	})

	t.Run("validate webhooks", func(t *testing.T) {
		doc := &openapi3.T{
			OpenAPI: "3.1.0",
			Info: &openapi3.Info{
				Title:   "Test API",
				Version: "1.0.0",
			},
			Paths: openapi3.NewPaths(),
			Webhooks: map[string]*openapi3.PathItem{
				"validWebhook": {
					Post: &openapi3.Operation{
						Responses: openapi3.NewResponses(
							openapi3.WithStatus(200, &openapi3.ResponseRef{
								Value: &openapi3.Response{
									Description: openapi3.Ptr("Success"),
								},
							}),
						),
					},
				},
			},
		}

		// Should validate successfully
		err := doc.Validate(t.Context())
		require.NoError(t, err)
	})

	t.Run("validate fails with nil webhook", func(t *testing.T) {
		doc := &openapi3.T{
			OpenAPI: "3.1.0",
			Info: &openapi3.Info{
				Title:   "Test API",
				Version: "1.0.0",
			},
			Paths: openapi3.NewPaths(),
			Webhooks: map[string]*openapi3.PathItem{
				"invalidWebhook": nil,
			},
		}

		err := doc.Validate(t.Context())
		require.Error(t, err)
		require.ErrorContains(t, err, "webhook")
		require.ErrorContains(t, err, "invalidWebhook")
	})
}

func TestJSONLookupWithWebhooks(t *testing.T) {
	doc := &openapi3.T{
		OpenAPI: "3.1.0",
		Info: &openapi3.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Paths: openapi3.NewPaths(),
		Webhooks: map[string]*openapi3.PathItem{
			"test": {
				Post: &openapi3.Operation{
					Summary: "Test webhook",
				},
			},
		},
	}

	result, err := doc.JSONLookup("webhooks")
	require.NoError(t, err)
	require.NotNil(t, result)

	webhooks, ok := result.(map[string]*openapi3.PathItem)
	require.True(t, ok)
	require.Contains(t, webhooks, "test")
}

func TestVersionBasedBehavior(t *testing.T) {
	t.Run("detect and handle OpenAPI 3.0", func(t *testing.T) {
		doc := &openapi3.T{
			OpenAPI: "3.0.3",
			Info: &openapi3.Info{
				Title:   "Test API",
				Version: "1.0.0",
			},
			Paths: openapi3.NewPaths(),
		}

		if doc.IsOpenAPI30() {
			// OpenAPI 3.0 specific logic
			require.Nil(t, doc.Webhooks)
		}
	})

	t.Run("detect and handle OpenAPI 3.1", func(t *testing.T) {
		doc := &openapi3.T{
			OpenAPI: "3.1.0",
			Info: &openapi3.Info{
				Title:   "Test API",
				Version: "1.0.0",
			},
			Paths: openapi3.NewPaths(),
			Webhooks: map[string]*openapi3.PathItem{
				"test": {
					Post: &openapi3.Operation{
						Summary: "Test",
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

		if doc.IsOpenAPI31OrLater() {
			// OpenAPI 3.1 specific logic
			require.NotNil(t, doc.Webhooks)
			require.Contains(t, doc.Webhooks, "test")
		}
	})
}

// TestValidateOpenAPIVersionField covers the `openapi` field's own validation:
// which version strings T.Validate accepts, and which typed error it reports
// for the ones it doesn't.
func TestValidateOpenAPIVersionField(t *testing.T) {
	newDoc := func(version string) *openapi3.T {
		return &openapi3.T{
			OpenAPI: version,
			Info:    &openapi3.Info{Title: "Test API", Version: "1.0.0"},
			Paths:   openapi3.NewPaths(),
		}
	}

	t.Run("accepted versions", func(t *testing.T) {
		// Every parseable 3.x is accepted, including minors and patch
		// levels released after this library.
		for _, version := range []string{
			"3", "3.0", "3.0.4", "3.0.99",
			"3.1", "3.1.2",
			"3.2", "3.2.1",
			"3.3.0", "3.10.0", "3.99",
		} {
			t.Run(version, func(t *testing.T) {
				require.NoError(t, newDoc(version).Validate(t.Context()))
			})
		}
	})

	t.Run("empty version is reported as required", func(t *testing.T) {
		err := newDoc("").Validate(t.Context())
		require.Error(t, err)

		var required *openapi3.OpenAPIVersionRequired
		require.ErrorAs(t, err, &required)

		var unsupported *openapi3.OpenAPIVersionUnsupportedError
		require.False(t, errors.As(err, &unsupported))
	})

	t.Run("unsupported versions", func(t *testing.T) {
		// Unparseable strings and non-3.x majors alike: previously these
		// validated silently (with 3.0-era strictness), now they're named.
		for _, version := range []string{
			"4.0.0", "2.0", "garbage", "v3.1", "3.", "3.x", "3.1.0.2", "3.1-rc", "3.1.0-rc.1",
		} {
			t.Run(version, func(t *testing.T) {
				err := newDoc(version).Validate(t.Context())
				require.Error(t, err)

				var unsupported *openapi3.OpenAPIVersionUnsupportedError
				require.ErrorAs(t, err, &unsupported)
				require.Equal(t, version, unsupported.Value)
				require.ErrorContains(t, err, version)
				require.Equal(t, "openapi-version-unsupported", unsupported.Code())

				var required *openapi3.OpenAPIVersionRequired
				require.False(t, errors.As(err, &required))
			})
		}
	})

	t.Run("multi error keeps validating past an unsupported version", func(t *testing.T) {
		doc := newDoc("4.0.0")
		doc.Info = &openapi3.Info{Title: "Test API"} // missing info.version

		err := doc.Validate(t.Context(), openapi3.EnableMultiError())
		require.Error(t, err)

		var me openapi3.MultiError
		require.ErrorAs(t, err, &me)
		require.Len(t, me, 2)

		var unsupported *openapi3.OpenAPIVersionUnsupportedError
		require.ErrorAs(t, err, &unsupported)
		var infoVersion *openapi3.InfoVersionRequired
		require.ErrorAs(t, err, &infoVersion)
	})
}

// TestValidateFutureMinorVersions is the headline of the numeric-comparison
// fix: a document declaring a 3.x version this library predates is validated
// as 3.1+, so 3.1-only constructs are allowed rather than rejected.
func TestValidateFutureMinorVersions(t *testing.T) {
	for _, version := range []string{"3.2.1", "3.3.0", "3.10.0"} {
		t.Run(version, func(t *testing.T) {
			doc := &openapi3.T{
				OpenAPI: version,
				Info: &openapi3.Info{
					Title:   "Test API",
					Summary: "A summary", // OpenAPI >=3.1 only
					Version: "1.0.0",
				},
				Paths: openapi3.NewPaths(),
				Webhooks: map[string]*openapi3.PathItem{ // OpenAPI >=3.1 only
					"newPet": {
						Post: &openapi3.Operation{
							Responses: openapi3.NewResponses(
								openapi3.WithStatus(200, &openapi3.ResponseRef{
									Value: &openapi3.Response{
										Description: openapi3.Ptr("Success"),
									},
								}),
							),
						},
					},
				},
			}

			require.True(t, doc.IsOpenAPI31OrLater())
			require.NoError(t, doc.Validate(t.Context()))
		})
	}
}

func TestMigrationScenario(t *testing.T) {
	t.Run("upgrade document from 3.0 to 3.1", func(t *testing.T) {
		// Start with 3.0 document
		doc := &openapi3.T{
			OpenAPI: "3.0.3",
			Info: &openapi3.Info{
				Title:   "Test API",
				Version: "1.0.0",
			},
			Paths: openapi3.NewPaths(),
		}

		require.True(t, doc.IsOpenAPI30())
		require.Nil(t, doc.Webhooks)

		// Upgrade to 3.1
		doc.OpenAPI = "3.1.0"

		// Add 3.1 features
		doc.Webhooks = map[string]*openapi3.PathItem{
			"newEvent": {
				Post: &openapi3.Operation{
					Summary: "New event notification",
					Responses: openapi3.NewResponses(
						openapi3.WithStatus(200, &openapi3.ResponseRef{
							Value: &openapi3.Response{
								Description: openapi3.Ptr("Processed"),
							},
						}),
					),
				},
			},
		}

		require.True(t, doc.IsOpenAPI31OrLater())
		require.NotNil(t, doc.Webhooks)

		// Validate the upgraded document
		err := doc.Validate(t.Context())
		require.NoError(t, err)
	})
}
