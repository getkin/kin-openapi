package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue542(t *testing.T) {
	spec := []byte(`
openapi: '3.0.0'
info:
  version: '1.0.0'
  title: Swagger Petstore
  license:
    name: MIT
servers:
- url: http://petstore.swagger.io/v1
paths: {}
components:
  schemas:
    Cat:
      anyOf:
      - $ref: '#/components/schemas/Kitten'
      - type: object
    Kitten:
      type: string
`[1:])

	sl := NewLoader()

	doc, err := sl.LoadFromData(spec)
	require.NoError(t, err)

	err = doc.Validate(sl.Context)
	require.NoError(t, err)
}
