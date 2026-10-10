//go:build go1.27 && goexperiment.jsonv2

// Build-tagged go1.27 because decodeJSONSlow only exists with json/v2.

package openapi3filter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestDecodeJSON checks that decodeJSON returns what decodeJSONSlow returns,
// errors included.
func TestDecodeJSON(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct{ name, body string }{
		{"empty", ``},
		{"null", `null`},
		{"values", `{"a":[1,-2.5e3,"x\n",true,false,null,{}],"b":{"c":[]}}`},
		{"trailing data", `{} trailing`},
		{"byte order mark", "\xef\xbb\xbf{}"},
		{"invalid UTF-8", "{\"\xff\":\"\xfe\"}"},
		{"duplicate name", `{"a":1,"a":2}`},
		{"duplicate invalid UTF-8 names", "{\"\xff\":1,\"\xfe\":2}"},
		{"duplicate lone surrogate names", `{"\ud800":1,"\udc00":2}`},
		{"truncated after duplicate", `{"a":1,"a":2,`},
		{"syntax error before duplicate", `{"a":1,,"a":2}`},
		{"too deep", strings.Repeat(`[`, 10001)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			want, wantErr := decodeJSONSlow(strings.NewReader(tc.body))
			got, err := decodeJSON(strings.NewReader(tc.body))
			require.Equal(t, wantErr, err)
			require.Equal(t, want, got)
		})
	}
}
