package openapi3

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// FuzzServerMatchRawURL matches arbitrary URLs against arbitrary server URL
// templates. It never panics, and a match accounts for the whole input: the
// template with its variables replaced by the returned values, followed by
// the remaining path, is the input again.
func FuzzServerMatchRawURL(f *testing.F) {
	f.Add("http://example.com/api/{v}", "http://example.com/api/v1/items")
	f.Add("{scheme}://{host}:{port}/base", "https://h:8443/base/x")
	f.Add("/", "/items")
	f.Add("", "")
	f.Add("{a}{b}", "xy")
	f.Add("/v{n}/", "/v2/")

	f.Fuzz(func(t *testing.T, template, input string) {
		if len(template) > 1<<10 || len(input) > 4<<10 {
			t.Skip()
		}
		params, remaining, ok := Server{URL: template}.MatchRawURL(input)
		if !ok {
			require.Nil(t, params)
			return
		}
		require.True(t, strings.HasPrefix(remaining, "/"), "remaining %q", remaining)

		// Rebuild the input from the template the way MatchRawURL walked it.
		var rebuilt strings.Builder
		pattern := template
		for len(pattern) > 0 {
			if len(pattern) == 1 && pattern[0] == '/' {
				break
			}
			if pattern[0] == '{' {
				end := strings.IndexByte(pattern, '}')
				require.GreaterOrEqual(t, end, 0)
				require.NotEmpty(t, params)
				rebuilt.WriteString(params[0])
				params = params[1:]
				pattern = pattern[end+1:]
				continue
			}
			rebuilt.WriteByte(pattern[0])
			pattern = pattern[1:]
		}
		require.Empty(t, params)
		rebuilt.WriteString(remaining)

		// An input that ends where the template does matches with "/" left.
		if got := rebuilt.String(); got != input {
			require.Equal(t, input+"/", got, "template %q input %q", template, input)
		}
	})
}
