package openapi3

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue430(t *testing.T) {
	schema := NewOneOfSchema(
		NewStringSchema().WithFormat("ipv4"),
		NewStringSchema().WithFormat("ipv6"),
	)

	err := schema.Validate(context.Background())
	require.NoError(t, err)

	err = schema.VisitJSON("127.0.1.1")
	require.NoError(t, err)
}
