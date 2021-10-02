package openapi3_test

import (
	"context"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func TestIssue430(t *testing.T) {
	schema := openapi3.NewOneOfSchema(
		openapi3.NewStringSchema().WithFormat("ipv4"),
		openapi3.NewStringSchema().WithFormat("ipv6"),
	)

	err := schema.Validate(context.Background())
	require.NoError(t, err)

	err = schema.VisitJSON("127.0.1.1")
	require.Error(t, err, openapi3.ErrOneOfConflict.Error())

	openapi3.DefineIPv4Format()
	openapi3.DefineIPv6Format()

	err = schema.VisitJSON("127.0.1.1")
	require.NoError(t, err)
}
