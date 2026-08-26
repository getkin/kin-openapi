package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
	goyaml "go.yaml.in/yaml/v3"
)

// YAML resolves more scalar forms than JSON does, so a value reaching an
// any-typed field must not carry a type that depends on its notation.
func TestNativeScalars_AnyValuesAreJSONShaped(t *testing.T) {
	const src = `
type: object
x-dec: 42
x-hex: 0x2A
x-underscore: 4_2
x-float: 1.5
x-str: "42"
x-bool: true
x-nested: {n: 7, list: [1, 2]}
example: 42
default: 0x10
`
	var node goyaml.Node
	require.NoError(t, goyaml.Unmarshal([]byte(src), &node))
	stripTimestamps(&node)

	var s Schema
	require.NoError(t, node.Content[0].Decode(&s))

	// Every integer notation lands as float64, as it would through JSON.
	for _, k := range []string{"x-dec", "x-hex", "x-underscore", "x-float"} {
		require.IsType(t, float64(0), s.Extensions[k], "%s", k)
	}
	require.EqualValues(t, 42, s.Extensions["x-dec"])
	require.EqualValues(t, 42, s.Extensions["x-hex"], "hex resolves to its value, not its text")
	require.EqualValues(t, 42, s.Extensions["x-underscore"])

	// Other types are untouched.
	require.Equal(t, "42", s.Extensions["x-str"])
	require.Equal(t, true, s.Extensions["x-bool"])

	// Nested maps and lists too.
	n := s.Extensions["x-nested"].(map[string]any)
	require.IsType(t, float64(0), n["n"])
	require.IsType(t, float64(0), n["list"].([]any)[0])

	// The any-typed struct fields the decoder fills directly.
	require.IsType(t, float64(0), s.Example)
	require.IsType(t, float64(0), s.Default)
	require.EqualValues(t, 16, s.Default)
}

// A declared integer field keeps its own type and full range: normalising
// those too would cost precision beyond 2^53, which the previous decode path
// did not.
func TestNativeScalars_DeclaredIntegerFieldsKeepPrecision(t *testing.T) {
	const src = "type: integer\nmaxLength: 9007199254740993\n"
	var node goyaml.Node
	require.NoError(t, goyaml.Unmarshal([]byte(src), &node))

	var s Schema
	require.NoError(t, node.Content[0].Decode(&s))
	require.NotNil(t, s.MaxLength)
	require.Equal(t, uint64(9007199254740993), *s.MaxLength, "must not round-trip through float64")
}

// Date-shaped scalars stay strings; an explicit tag still asks for a time.
func TestNativeScalars_Timestamps(t *testing.T) {
	const src = "example: 2020-06-11T16:32:50Z\n"
	var node goyaml.Node
	require.NoError(t, goyaml.Unmarshal([]byte(src), &node))
	stripTimestamps(&node)

	var s Schema
	require.NoError(t, node.Content[0].Decode(&s))
	require.IsType(t, "", s.Example, "a date-shaped example stays a string")
}
