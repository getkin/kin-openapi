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
	require.NoError(t, prepareScalars(&node))

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

// Date-shaped scalars stay strings, tagged or not: JSON has no timestamp, and
// a time.Time would marshal to text that reads back as a string.
func TestNativeScalars_Timestamps(t *testing.T) {
	const src = "example: 2020-06-11T16:32:50Z\n"
	var node goyaml.Node
	require.NoError(t, goyaml.Unmarshal([]byte(src), &node))
	require.NoError(t, prepareScalars(&node))

	var s Schema
	require.NoError(t, node.Content[0].Decode(&s))
	require.IsType(t, "", s.Example, "a date-shaped example stays a string")

	const tagged = "example: !!timestamp 2020-06-11T16:32:50Z\n"
	var taggedNode goyaml.Node
	require.NoError(t, goyaml.Unmarshal([]byte(tagged), &taggedNode))
	require.NoError(t, prepareScalars(&taggedNode))

	var ts Schema
	require.NoError(t, taggedNode.Content[0].Decode(&ts))
	require.IsType(t, "", ts.Example, "a tagged one too, the json path doing the same")
}

// decodeNative decodes src the way the loader's native path does.
func decodeNative(t *testing.T, src string, v any) error {
	t.Helper()

	var node goyaml.Node
	require.NoError(t, goyaml.Unmarshal([]byte(src), &node))
	if err := prepareScalars(&node); err != nil {
		return err
	}
	return node.Content[0].Decode(v)
}

// A string cannot hold nil, so the decoder drops a null from a list of them
// and the list comes back shorter than the document. The json path keeps the
// place of the null as an empty string, and so does this.
func TestNativeScalars_NullKeepsItsPlaceInAListOfStrings(t *testing.T) {
	var s Schema
	require.NoError(t, decodeNative(t, "required: [a, null, b]\ntype: [string, null]\n", &s))
	require.Equal(t, []string{"a", "", "b"}, s.Required)
	require.Equal(t, Types{"string", ""}, *s.Type)

	var op Operation
	require.NoError(t, decodeNative(t, "tags: [a, null]\n", &op))
	require.Equal(t, []string{"a", ""}, op.Tags)

	var v ServerVariable
	require.NoError(t, decodeNative(t, "default: d\nenum: [a, null]\n", &v))
	require.Equal(t, []string{"a", ""}, v.Enum)

	var sr SecurityRequirement
	require.NoError(t, decodeNative(t, "scheme: [read, null]\n", &sr))
	require.Equal(t, []string{"read", ""}, sr["scheme"])
}

// A scalar's text is not a string where the scalar resolved to something else:
// `title: 1` is a number, which the json path refuses rather than reading as
// "1".
func TestNativeScalars_StringFieldsTakeStringsOnly(t *testing.T) {
	for _, test := range []struct {
		name, src, want string
	}{
		{"an integer", "title: 1\nversion: \"1\"\n", `cannot unmarshal !!int into string field "title"`},
		{"a float", "title: t\nversion: 1.0\n", `cannot unmarshal !!float into string field "version"`},
		{"a boolean", "title: true\nversion: \"1\"\n", `cannot unmarshal !!bool into string field "title"`},
	} {
		t.Run(test.name, func(t *testing.T) {
			var info Info
			require.ErrorContains(t, decodeNative(t, test.src, &info), test.want)
		})
	}

	var s Schema
	require.ErrorContains(t, decodeNative(t, "required: [1]\n", &s),
		`cannot unmarshal !!int into string field "required"`)
	require.ErrorContains(t, decodeNative(t, "discriminator: {propertyName: p, mapping: {a: 1}}\n", &s),
		`cannot unmarshal !!int into string field "a"`)

	// A null is the empty string, as it is on the json path, and a date-shaped
	// scalar is a string in this document's data model.
	var info Info
	require.NoError(t, decodeNative(t, "title: null\nversion: 2020-01-01\n", &info))
	require.Equal(t, "", info.Title)
	require.Equal(t, "2020-01-01", info.Version)
}

// YAML allows any scalar as a key and a document's names are strings, so a
// scalar key is spelled out. A null key spells nothing, and the json path has
// no key to decode it into, so the document is refused.
func TestNativeScalars_NullKeyNamesNothing(t *testing.T) {
	var extension, example Schema
	require.ErrorContains(t, decodeNative(t, "x-k: {null: a}\n", &extension), "mapping key is null")
	require.ErrorContains(t, decodeNative(t, "example: {null: a}\n", &example), "mapping key is null")

	// Every other scalar has a name.
	var named Schema
	require.NoError(t, decodeNative(t, "x-k: {1: a, true: b, 1.5: c, 0x2A: d}\n", &named))
	require.Equal(t, map[string]any{"1": "a", "true": "b", "1.5": "c", "42": "d"}, named.Extensions["x-k"])
}

// A requirement written as an alias is the mapping it points at, so a null in
// its scopes keeps its place there too.
func TestNativeScalars_NullKeepsItsPlaceThroughAnAlias(t *testing.T) {
	var doc T
	require.NoError(t, decodeNative(t, "x-s: &s {scheme: [read, null]}\nsecurity: [*s]\n", &doc))
	require.Equal(t, []string{"read", ""}, doc.Security[0]["scheme"])
}
