package openapi3_test

import (
	"context"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
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
