package openapi3_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func TestIssue697(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile("testdata/issue697.yml")
	require.NoError(t, err)
	err = doc.Validate(loader.Context)
	require.NoError(t, err)
}
