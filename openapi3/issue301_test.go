package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestIssue301(t *testing.T) {
	sl := openapi3.NewLoader()
	sl.IsExternalRefsAllowed = true

	doc, err := sl.LoadFromFile("testdata/callbacks.yml")
	require.NoError(t, err)

	err = doc.Validate(sl.Context)
	require.NoError(t, err)

	require.Equal(t, &openapi3.Types{"object"}, doc.
		Paths.Value("/trans").
		Post.Callbacks["transactionCallback"].Value.
		Value("http://notificationServer.com?transactionId={$request.body#/id}&email={$request.body#/email}").
		Post.RequestBody.Value.
		Content["application/json"].Schema.Value.
		Type)

	require.Equal(t, &openapi3.Types{"boolean"}, doc.
		Paths.Value("/other").
		Post.Callbacks["myEvent"].Value.
		Value("{$request.query.queryUrl}").
		Post.RequestBody.Value.
		Content["application/json"].Schema.Value.
		Type)
}
