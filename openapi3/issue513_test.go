package openapi3

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtraSiblingsInRemoteRef(t *testing.T) {
	spec := []byte(`
openapi: 3.0.1
servers:
- url: http://localhost:5000
info:
  version: v1
  title: Products api
  contact:
    name: me
    email: me@github.com
  description: This is a sample
paths:
  /categories:
    get:
      summary: Provides the available categories for the store
      operationId: list-categories
      responses:
        '200':
          description: this is a desc
          content:
            application/json:
              schema:
                $ref: http://schemas.sentex.io/store/categories.json
`[1:])

	// When that site fails to respond:
	// see https://github.com/getkin/kin-openapi/issues/495

	// http://schemas.sentex.io/store/categories.json
	// {
	//   "$id": "http://schemas.sentex.io/store/categories.json",
	//   "$schema": "http://json-schema.org/draft-07/schema#",
	//   "description": "array of category strings",
	//   "type": "array",
	//   "items": {
	//     "allOf": [
	//       {
	//         "$ref": "http://schemas.sentex.io/store/category.json"
	//       }
	//     ]
	//   }
	// }

	// http://schemas.sentex.io/store/category.json
	// {
	//   "$id": "http://schemas.sentex.io/store/category.json",
	//   "$schema": "http://json-schema.org/draft-07/schema#",
	//   "description": "category name for products",
	//   "type": "string",
	//   "pattern": "^[A-Za-z0-9\\-]+$",
	//   "minimum": 1,
	//   "maximum": 30
	// }

	sl := NewLoader()
	sl.IsExternalRefsAllowed = true

	doc, err := sl.LoadFromData(spec)
	require.NoError(t, err)

	err = doc.Validate(sl.Context, AllowExtraSiblingFields("$id", "$schema"))
	require.NoError(t, err)
}

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
	require.Contains(t, doc.Paths.Value("/v1/operation").Delete.Responses.Default().Value.Extensions, `x-my-extension`)
	err = doc.Validate(sl.Context)
	require.ErrorContains(t, err, `extra sibling fields: [schema]`)
}

func TestIssue513KOMixesRefAlongWithOtherFieldsDisallowed(t *testing.T) {
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
