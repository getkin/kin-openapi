package openapi3_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func TestIssue753(t *testing.T) {
	loader := openapi3.NewLoader()

	doc, err := loader.LoadFromFile("testdata/issue753.yml")
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	require.NotNil(t, doc.
		Paths.Value("/test1").
		Post.Callbacks["callback1"].Value.
		Value("{$request.body#/callback}").
		Post.RequestBody.Value.
		Content["application/json"].
		Schema.Value)
	require.NotNil(t, doc.
		Paths.Value("/test2").
		Post.Callbacks["callback2"].Value.
		Value("{$request.body#/callback}").
		Post.RequestBody.Value.
		Content["application/json"].
		Schema.Value)
}
