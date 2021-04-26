package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue341(t *testing.T) {
	sl := NewLoader()
	sl.IsExternalRefsAllowed = true
	doc, err := sl.LoadFromFile("testdata/main.yaml")
	require.NoError(t, err)

	err = doc.Validate(sl.Context)
	require.NoError(t, err)

	err = sl.ResolveRefsIn(doc, nil)
	require.NoError(t, err)

	bs, err := doc.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, []byte(`{"components":{},"info":{"title":"test file","version":"n/a"},"openapi":"3.0.0","paths":{"/testpath":{"get":{"responses":{"200":{"$ref":"#/components/responses/testpath_200_response"}}}}}}`), bs)

	require.Equal(t, "string", doc.Paths["/testpath"].Get.Responses["200"].Value.Content["application/json"].Schema.Value.Type)
}
