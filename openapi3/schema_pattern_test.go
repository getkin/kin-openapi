package openapi3

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPattern(t *testing.T) {
	_, err := regexp.Compile("^[a-zA-Z\\u0080-\\u024F\\s\\/\\-\\)\\(\\`\\.\\\"\\']+$")
	require.EqualError(t, err, "error parsing regexp: invalid escape sequence: `\\u`")

	_, err = regexp.Compile(`^[a-zA-Z\x{0080}-\x{024F}]+$`)
	require.NoError(t, err)

	require.Equal(t, `^[a-zA-Z\x{0080}-\x{024F}]+$`, intoGoRegexp(`^[a-zA-Z\u0080-\u024F]+$`))
	require.Equal(t, `^[6789a-zA-Z\x{0080}-\x{024F}]+$`, intoGoRegexp(`^[6789a-zA-Z\u0080-\u024F]+$`))
}

func TestSchemaPatternCache(t *testing.T) {

	var schema Schema

	schema.Pattern = `^[a-zA-Z\x{0080}-\x{024F}]+$`
	cp, err := schema.compilePattern(nil)
	require.NoError(t, err)
	require.NotNil(t, cp)

	// The compiled pattern should be cached
	v, ok := compiledPatterns.Load(schema.Pattern)
	require.True(t, ok)
	require.Equal(t, cp, v)
}
