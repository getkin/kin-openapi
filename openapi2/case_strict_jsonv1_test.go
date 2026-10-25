//go:build !goexperiment.jsonv2

package openapi2_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi2"
)

// TestFieldsIgnoreKeysOfAnotherCase records how encoding/json v1, which
// ignores case:strict, matches keys to fields regardless of case: such keys
// fill the field and are also kept as extensions.
func TestFieldsIgnoreKeysOfAnotherCase(t *testing.T) {
	t.Parallel()

	var doc openapi2.T
	err := json.Unmarshal([]byte(`{"swagger":"2.0","Host":"example.com","info":{"Title":"t"}}`), &doc)
	require.NoError(t, err)
	require.Equal(t, "2.0", doc.Swagger)
	require.Equal(t, "example.com", doc.Host)
	require.Equal(t, "t", doc.Info.Title)
	require.Equal(t, map[string]any{"Host": "example.com"}, doc.Extensions)
	require.Equal(t, map[string]any{"Title": "t"}, doc.Info.Extensions)
}
