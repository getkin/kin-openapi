package openapi3_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func TestValidateInfo_SummaryIn30(t *testing.T) {
	spec := []byte(`
openapi: '3'
paths: {}
info:
  title: An API
  version: 1.2.3.4
  summary: bla
`)

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(spec)
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.ErrorContains(t, err, "invalid info")
	require.ErrorContains(t, err, "field summary")
}

func TestValidateInfo_SummaryIn31(t *testing.T) {
	spec := []byte(`
openapi: '3.1'
paths: {}
info:
  title: An API
  version: 1.2.3.4
  summary: bla
`)

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(spec)
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)
}
