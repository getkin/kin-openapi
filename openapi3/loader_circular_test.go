package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadCircular(t *testing.T) {
	loader := NewLoader()
	loader.IsExternalRefsAllowed = true
	_, err := loader.LoadFromFile("testdata/circularRef2/circular2.yaml")
	require.NoError(t, err)
}
