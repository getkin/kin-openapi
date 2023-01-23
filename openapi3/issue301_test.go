package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue301(t *testing.T) {
	sl := NewLoader()
	sl.IsExternalRefsAllowed = true

	doc, err := sl.LoadFromFile("testdata/callbacks.yml")
	require.NoError(t, err)

	err = doc.Validate(sl.Context)
	require.NoError(t, err)

	require.Equal(t, "object", doc.
		Paths.Value("/trans").
		Post.Callbacks["transactionCallback"].Value.
		Value("http://notificationServer.com?transactionId={$request.body#/id}&email={$request.body#/email}").
		Post.RequestBody.Value.
		Content["application/json"].Schema.Value.
		Type)

	require.Equal(t, "boolean", doc.
		Paths.Value("/other").
		Post.Callbacks["myEvent"].Value.
		Value("{$request.query.queryUrl}").
		Post.RequestBody.Value.
		Content["application/json"].Schema.Value.
		Type)
}
