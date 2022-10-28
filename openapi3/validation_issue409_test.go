package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestIssue409PatternIgnored(t *testing.T) {
	l := openapi3.NewLoader()
	s, err := l.LoadFromFile("testdata/issue409.yml")
	require.NoError(t, err)

	err = s.Validate(l.Context, openapi3.DisableSchemaPatternValidation())
	assert.NoError(t, err)
}

func TestIssue409PatternNotIgnored(t *testing.T) {
	l := openapi3.NewLoader()
	s, err := l.LoadFromFile("testdata/issue409.yml")
	require.NoError(t, err)

	err = s.Validate(l.Context)
	assert.Error(t, err)
}

func TestIssue409HygienicUseOfCtx(t *testing.T) {
	l := openapi3.NewLoader()
	doc, err := l.LoadFromFile("testdata/issue409.yml")
	require.NoError(t, err)

	err = doc.Validate(l.Context, openapi3.DisableSchemaPatternValidation())
	assert.NoError(t, err)
	err = doc.Validate(l.Context)
	assert.Error(t, err)

	// and the other way

	l = openapi3.NewLoader()
	doc, err = l.LoadFromFile("testdata/issue409.yml")
	require.NoError(t, err)

	err = doc.Validate(l.Context)
	assert.Error(t, err)
	err = doc.Validate(l.Context, openapi3.DisableSchemaPatternValidation())
	assert.NoError(t, err)
}
