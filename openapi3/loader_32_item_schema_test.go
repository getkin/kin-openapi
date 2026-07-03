package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestOpenAPI32MediaTypeItemSchema(t *testing.T) {
	doc, err := openapi3.NewLoader().LoadFromData([]byte(`
openapi: 3.2.0
info:
  title: SSE API
  version: 1.0.0
paths:
  /events:
    get:
      responses:
        "200":
          description: event stream
          content:
            text/event-stream:
              itemSchema:
                $ref: "#/components/schemas/SseEvent"
components:
  schemas:
    SseEvent:
      type: object
      required:
        - data
      properties:
        data:
          type: string
        event:
          type: string
        id:
          type: string
        retry:
          type: integer
          minimum: 0
`))
	require.NoError(t, err)

	media := doc.Paths.Value("/events").Get.Responses.Status(200).Value.Content.Get("text/event-stream")
	require.NotNil(t, media)
	require.NotNil(t, media.ItemSchema)
	require.Equal(t, "#/components/schemas/SseEvent", media.ItemSchema.Ref)
	require.NotNil(t, media.ItemSchema.Value)
	require.True(t, media.ItemSchema.Value.Type.Is(openapi3.TypeObject))

	err = doc.Validate(t.Context())
	require.NoError(t, err)
}

func TestWalkSchemasVisitsMediaTypeItemSchema(t *testing.T) {
	itemSchema := &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{openapi3.TypeObject}}}
	doc := &openapi3.T{
		OpenAPI: "3.2.0",
		Info:    &openapi3.Info{Title: "SSE API", Version: "1.0.0"},
		Paths: openapi3.NewPaths(openapi3.WithPath("/events", &openapi3.PathItem{
			Get: &openapi3.Operation{
				Responses: openapi3.NewResponses(openapi3.WithStatus(200, &openapi3.ResponseRef{
					Value: &openapi3.Response{
						Description: openapi3.Ptr("event stream"),
						Content: openapi3.Content{
							"text/event-stream": {ItemSchema: itemSchema},
						},
					},
				})),
			},
		})),
	}

	var got []string
	err := doc.WalkSchemas(func(jsonPointer string, schema *openapi3.SchemaRef) error {
		got = append(got, jsonPointer)
		return nil
	})
	require.NoError(t, err)
	require.Contains(t, got, "/paths/~1events/get/responses/200/content/text~1event-stream/itemSchema")
}
