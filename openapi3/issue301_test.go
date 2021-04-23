package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue301(t *testing.T) {
	sl := NewSwaggerLoader()
	sl.IsExternalRefsAllowed = true

	doc, err := sl.LoadSwaggerFromFile("testdata/callbacks.yml")
	require.NoError(t, err)

	err = doc.Validate(sl.Context)
	require.NoError(t, err)

	transCallbacks := doc.Paths["/trans"].Post.Callbacks["transactionCallback"].Value
	require.Equal(t, "object", (*transCallbacks)["http://notificationServer.com?transactionId={$request.body#/id}&email={$request.body#/email}"].Post.RequestBody.
		Value.Content["application/json"].Schema.
		Value.Type)

	otherCallbacks := doc.Paths["/other"].Post.Callbacks["myCallback"].Value
	require.Equal(t, "boolean", (*otherCallbacks)["{$request.query.queryUrl}"]) //.Post.RequestBody.
	// Value.Content["application/json"].Schema.
	// Value.Type)
}
