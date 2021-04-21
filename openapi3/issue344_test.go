package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue344(t *testing.T) {
	sl := NewSwaggerLoader()
	sl.IsExternalRefsAllowed = true

	doc, err := sl.LoadSwaggerFromFile("testdata/spec.yaml")
	require.NoError(t, err)

	err = doc.Validate(sl.Context)
	require.NoError(t, err)
}
