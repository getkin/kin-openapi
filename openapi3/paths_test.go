package openapi3

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPathValidate(t *testing.T) {
	const emptyPathSpec = `
openapi: "3.0.0"
info:
  version: 1.0.0
  title: Swagger Petstore
  license:
    name: MIT
servers:
  - url: http://petstore.swagger.io/v1
paths:
  /pets:
`
	doc, err := NewLoader().LoadFromData([]byte(emptyPathSpec))
	require.NoError(t, err)
	err = doc.Paths.Validate(context.Background())
	require.NoError(t, err)
}
