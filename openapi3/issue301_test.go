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

	transCallbacks := doc.Paths["/trans"].Post.Callbacks["transactionCallback"].Value
	require.Equal(t, "object", (*transCallbacks)["http://notificationServer.com?transactionId={$request.body#/id}&email={$request.body#/email}"].Post.RequestBody.
		Value.Content["application/json"].Schema.
		Value.Type)

	otherCallbacks := doc.Paths["/other"].Post.Callbacks["myEvent"].Value
	require.Equal(t, "boolean", (*otherCallbacks)["{$request.query.queryUrl}"].Post.RequestBody.
		Value.Content["application/json"].Schema.
		Value.Type)
}
