package openapi3_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func issue1359Spec(version string) []byte {
	return []byte(`
openapi: "` + version + `"
info:
  title: Ref Sibling Marshal
  version: "1.0"
paths: {}
components:
  schemas:
    uuid:
      type: string
      format: uuid
    Thing:
      type: object
      required: [id]
      properties:
        id:
          $ref: '#/components/schemas/uuid'
          readOnly: true
          description: The thing identifier.
          x-order: 1
          unknownKeyword: dropped
        other:
          $ref: '#/components/schemas/uuid'
`)
}

func TestIssue1359(t *testing.T) {
	for _, version := range []string{"3.0.3", "3.1.0"} {
		t.Run(version, func(t *testing.T) {
			ctx := context.Background()
			loader := openapi3.NewLoader()
			doc, err := loader.LoadFromData(issue1359Spec(version))
			require.NoError(t, err)

			data, err := json.Marshal(doc)
			require.NoError(t, err)

			var raw struct {
				Components struct {
					Schemas map[string]struct {
						Properties map[string]map[string]any `json:"properties"`
					} `json:"schemas"`
				} `json:"components"`
			}
			require.NoError(t, json.Unmarshal(data, &raw))
			props := raw.Components.Schemas["Thing"].Properties
			require.Equal(t, map[string]any{
				"$ref":        "#/components/schemas/uuid",
				"readOnly":    true,
				"description": "The thing identifier.",
				"x-order":     float64(1),
			}, props["id"])
			require.Equal(t, map[string]any{
				"$ref": "#/components/schemas/uuid",
			}, props["other"])

			yml, err := doc.MarshalYAML()
			require.NoError(t, err)
			ymlData, err := json.Marshal(yml)
			require.NoError(t, err)
			require.JSONEq(t, string(data), string(ymlData))

			reloaded, err := openapi3.NewLoader().LoadFromData(data)
			require.NoError(t, err)
			id := reloaded.Components.Schemas["Thing"].Value.Properties["id"]
			require.Equal(t, "#/components/schemas/uuid", id.Ref)
			require.Equal(t, map[string]any{"x-order": float64(1)}, id.Extensions)

			if version == "3.1.0" {
				require.NoError(t, reloaded.Validate(ctx))
				require.True(t, id.Value.ReadOnly)
				require.Equal(t, "The thing identifier.", id.Value.Description)
			} else {
				// OpenAPI 3.0: siblings are serialized as in the source but
				// still ignored when resolving the $ref.
				require.False(t, id.Value.ReadOnly)
				require.Empty(t, id.Value.Description)
			}
		})
	}
}
