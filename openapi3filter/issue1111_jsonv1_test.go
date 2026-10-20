//go:build !goexperiment.jsonv2

package openapi3filter_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestIssue1111Decoding records how an application decoding a validated body
// with encoding/json v1 sees another value than the one validated: v1 ignores
// case:strict and matches keys to fields regardless of case, the last one winning.
func TestIssue1111Decoding(t *testing.T) {
	t.Parallel()

	for name, expected := range map[string]string{
		"good":         "aniki",
		"bad":          "beniki",
		"evil":         "beniki",
		"chaotic evil": "chaos",
	} {
		var val struct {
			Value string `json:"value,case:strict"`
		}
		err := json.Unmarshal([]byte(issue1111Bodies[name]), &val)
		require.NoError(t, err)
		require.Equal(t, expected, val.Value, name)
	}
}
