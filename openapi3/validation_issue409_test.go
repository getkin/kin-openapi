package openapi3_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIssue409(t *testing.T) {
	l := openapi3.NewLoader()
	s, err := l.LoadFromFile("testdata/issue409.yml")
	require.NoError(t, err)

	err = s.Validate(l.Context, openapi3.DisableSchemaPatternValidation())
	assert.NoError(t, err)
}
