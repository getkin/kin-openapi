package openapi2_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi2"
)

func TestResponsesUnmarshalErrorHasKey(t *testing.T) {
	var doc openapi2.T
	err := json.Unmarshal([]byte(`{"paths": {"/": {"get": {"responses": {"200": {"description": "OK"}, "404": {"description": 42}}}}}}`), &doc)
	if openapi2.UsingJSONv2 {
		require.ErrorContains(t, err, `.responses.404.description `)
	} else {
		require.ErrorContains(t, err, `response "404": `)
	}
}
