//go:build goexperiment.jsonv2

package openapi3filter_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestIssue1111Decoding checks that an application decoding a validated body
// sees the value that was validated when its struct tags use case:strict,
// which encoding/json honors whenever it is implemented with encoding/json/v2
// (GOEXPERIMENT=jsonv2, the default from Go 1.27).
func TestIssue1111Decoding(t *testing.T) {
	for name, expected := range map[string]string{
		"good":         "aniki",
		"bad":          "beniki",
		"evil":         "aniki",
		"chaotic evil": "",
	} {
		var val struct {
			Value string `json:"value,case:strict"`
		}
		err := json.Unmarshal([]byte(issue1111Bodies[name]), &val)
		require.NoError(t, err)
		require.Equal(t, expected, val.Value, name)
	}
}
