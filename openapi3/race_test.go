package openapi3_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestRaceyPatternSchema(t *testing.T) {
	schema := openapi3.Schema{
		Pattern: "^test|for|race|condition$",
		Type:    "string",
	}

	err := schema.Validate(context.Background())
	require.NoError(t, err)

	visit := func() {
		err := schema.VisitJSONString("test")
		require.NoError(t, err)
	}

	go visit()
	visit()
}
