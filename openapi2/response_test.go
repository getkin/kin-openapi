package openapi2_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi2"
)

func TestResponsesUnmarshalErrorHasKey(t *testing.T) {
	var responses openapi2.Responses
	err := json.Unmarshal([]byte(`{"200": {"description": "OK"}, "404": {"description": 42}}`), &responses)
	require.ErrorContains(t, err, `response "404": `)
}
