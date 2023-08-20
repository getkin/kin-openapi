package openapi3_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestRaceyPatternSchemaValidateHindersIt(t *testing.T) {
	schema := openapi3.NewStringSchema().WithPattern("^test|for|race|condition$")

	err := schema.Validate(context.Background())
	require.NoError(t, err)

	visit := func() {
		err := schema.VisitJSONString("test")
		require.NoError(t, err)
	}

	go visit()
	visit()
}

func TestRaceyPatternSchemaForIssue775(t *testing.T) {
	schema := openapi3.NewStringSchema().WithPattern("^test|for|race|condition$")

	// err := schema.Validate(context.Background())
	// require.NoError(t, err)

	visit := func() {
		err := schema.VisitJSONString("test")
		require.NoError(t, err)
	}

	go visit()
	visit()
}
