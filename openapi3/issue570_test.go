package openapi3_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func TestIssue570(t *testing.T) {
	loader := openapi3.NewLoader()
	_, err := loader.LoadFromFile("testdata/issue570.json")
	require.NoError(t, err)
}
