//go:build goexperiment.jsonv2

package openapi2_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi2"
)

// TestFieldsIgnoreKeysOfAnotherCase covers keys that differ from a known
// field only in case: the json:",case:strict" tags make encoding/json keep
// them as extensions rather than fill the field.
func TestFieldsIgnoreKeysOfAnotherCase(t *testing.T) {
	t.Parallel()

	var doc openapi2.T
	err := json.Unmarshal([]byte(`{"swagger":"2.0","Host":"example.com","info":{"Title":"t"}}`), &doc)
	require.NoError(t, err)
	require.Equal(t, "2.0", doc.Swagger)
	require.Empty(t, doc.Host)
	require.Empty(t, doc.Info.Title)
	require.Equal(t, map[string]any{"Host": "example.com"}, doc.Extensions)
	require.Equal(t, map[string]any{"Title": "t"}, doc.Info.Extensions)
}
