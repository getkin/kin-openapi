package openapi3_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

// YAML has infinity and NaN and JSON does not, so a document carrying one
// cannot be written back. It is refused where it is read, naming the line,
// rather than at the marshal that would otherwise fail later.
func TestNonFiniteNumbersAreRefused(t *testing.T) {
	t.Parallel()

	const header = "openapi: 3.1.0\ninfo: {title: t, version: \"1\"}\npaths: {}\n"

	for _, tt := range []struct {
		name, doc, want string
	}{
		{name: "schema bound", doc: "components:\n schemas:\n  S: {type: number, maximum: .inf}\n", want: "line 6: .inf"},
		{name: "negative bound", doc: "components:\n schemas:\n  S: {type: number, minimum: -.inf}\n", want: "line 6: -.inf"},
		{name: "extension", doc: "x-ratio: .NaN\n", want: "line 4: .NaN"},
		{name: "inside a sequence", doc: "x-list: [1, +.inf]\n", want: "line 4: +.inf"},
		{name: "explicitly tagged", doc: "x-ratio: !!float .inf\n", want: "line 4: .inf"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			loader := openapi3.NewLoader()
			_, err := loader.LoadFromData([]byte(header + tt.doc))
			require.Error(t, err)
			require.ErrorContains(t, err, tt.want)
			require.ErrorContains(t, err, noJSONRepresentation)
		})
	}
}

const noJSONRepresentation = "has no JSON representation"

// A float JSON can represent, and a quoted one that is a string, both load.
func TestFiniteNumbersAndQuotedInfinityLoad(t *testing.T) {
	t.Parallel()

	const spec = "openapi: 3.1.0\ninfo: {title: t, version: \"1\"}\npaths: {}\nx-big: 1e308\nx-text: '.inf'\n"

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)
	require.InDelta(t, 1e308, doc.Extensions["x-big"], 0)
	require.Equal(t, ".inf", doc.Extensions["x-text"])

	out, err := doc.MarshalJSON()
	require.NoError(t, err)
	require.Contains(t, string(out), fmt.Sprintf("%v", 1e308))
}
