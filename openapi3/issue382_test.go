package openapi3_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func TestOverridingGlobalParametersValidation(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile("testdata/Test_param_override.yml")
	require.NoError(t, err)
	err = doc.Validate(loader.Context)
	require.NoError(t, err)
}
