package openapi3

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue513OKWithExtension(t *testing.T) {
	spec := `
openapi: "3.0.3"
info:
  title: 'My app'
  version: 1.0.0
  description: 'An API'

paths:
  /v1/operation:
    delete:
      summary: Delete something
      responses:
        200:
          description: Success
        default:
          description: '* **400** - Bad Request'
          x-my-extension: {val: ue}
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
components:
  schemas:
    Error:
      type: object
      description: An error response body.
      properties:
        message:
          description: A detailed message describing the error.
          type: string
`[1:]
	sl := NewLoader()
	doc, err := sl.LoadFromData([]byte(spec))
	require.NoError(t, err)
	err = doc.Validate(sl.Context)
	require.NoError(t, err)
	data, err := json.Marshal(doc)
	require.NoError(t, err)
	require.Contains(t, string(data), `x-my-extension`)
}

func TestIssue513KOHasExtraFieldSchema(t *testing.T) {
	spec := `
openapi: "3.0.3"
info:
  title: 'My app'
  version: 1.0.0
  description: 'An API'

paths:
  /v1/operation:
    delete:
      summary: Delete something
      responses:
        200:
          description: Success
        default:
          description: '* **400** - Bad Request'
          x-my-extension: {val: ue}
          # Notice here schema is invalid. It should instead be:
          # content:
          #   application/json:
          #     schema:
          #       $ref: '#/components/schemas/Error'
          schema:
            $ref: '#/components/schemas/Error'
components:
  schemas:
    Error:
      type: object
      description: An error response body.
      properties:
        message:
          description: A detailed message describing the error.
          type: string
`[1:]
	sl := NewLoader()
	doc, err := sl.LoadFromData([]byte(spec))
	require.NoError(t, err)
	require.Contains(t, doc.Paths["/v1/operation"].Delete.Responses["default"].Value.Extensions, `x-my-extension`)
	err = doc.Validate(sl.Context)
	require.ErrorContains(t, err, `extra sibling fields: [schema]`)
}

func TestIssue513KOMixesRefAlongWithOtherFields(t *testing.T) {
	spec := `
openapi: "3.0.3"
info:
  title: 'My app'
  version: 1.0.0
  description: 'An API'

paths:
  /v1/operation:
    delete:
      summary: Delete something
      responses:
        200:
          description: A sibling field that the spec says is ignored
          $ref: '#/components/responses/SomeResponseBody'
components:
  responses:
    SomeResponseBody:
      description: Success
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
  schemas:
    Error:
      type: object
      description: An error response body.
      properties:
        message:
          description: A detailed message describing the error.
          type: string
`[1:]
	sl := NewLoader()
	doc, err := sl.LoadFromData([]byte(spec))
	require.NoError(t, err)
	err = doc.Validate(sl.Context)
	require.ErrorContains(t, err, `extra sibling fields: [description]`)
}

func TestIssue513KOMixesRefAlongWithOtherFieldsAllowed(t *testing.T) {
	spec := `
openapi: "3.0.3"
info:
  title: 'My app'
  version: 1.0.0
  description: 'An API'

paths:
  /v1/operation:
    delete:
      summary: Delete something
      responses:
        200:
          description: A sibling field that the spec says is ignored
          $ref: '#/components/responses/SomeResponseBody'
components:
  responses:
    SomeResponseBody:
      description: Success
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/Error'
  schemas:
    Error:
      type: object
      description: An error response body.
      properties:
        message:
          description: A detailed message describing the error.
          type: string
`[1:]
	sl := NewLoader()
	doc, err := sl.LoadFromData([]byte(spec))
	require.NoError(t, err)
	err = doc.Validate(sl.Context, AllowExtraSiblingFields("description"))
	require.NoError(t, err)
}
