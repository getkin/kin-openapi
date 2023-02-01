package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue759(t *testing.T) {
	spec := []byte(`
openapi: 3.0.0
info:
  title: title
  description: description
  version: 0.0.0
paths:
  /slash:
    get:
      responses:
        "200":
          # Ref should point to a response, not a schema
          $ref: "#/components/schemas/UserStruct"
components:
  schemas:
    UserStruct:
      type: object
`[1:])

	loader := NewLoader()

	doc, err := loader.LoadFromData(spec)
	require.Nil(t, doc)
	require.EqualError(t, err, `bad data in "#/components/schemas/UserStruct" (expecting ref to response object)`)
}
