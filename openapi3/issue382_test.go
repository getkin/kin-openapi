package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOverridingGlobalParametersValidation(t *testing.T) {
	loader := NewLoader()
	doc, err := loader.LoadFromFile("testdata/Test_param_override.yml")
	require.NoError(t, err)
	err = doc.Validate(loader.Context)
	require.NoError(t, err)
}
