package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCrashOnLoad(t *testing.T) {
	loader := NewLoader()
	_, err := loader.LoadFromFile("testdata/issue794.yml")
	require.NoError(t, err)
}
