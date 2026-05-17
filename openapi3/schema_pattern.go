package openapi3

import (
	"fmt"
	"regexp"
)

var patRewriteCodepoints = regexp.MustCompile(`(?P<replaced_with_slash_x>\\u)(?P<code>[0-9A-F]{4})`)

// See https://pkg.go.dev/regexp/syntax
func intoGoRegexp(re string) string {
	return patRewriteCodepoints.ReplaceAllString(re, `\x{${code}}`)
}

// NOTE: racey WRT [writes to schema.Pattern] vs [reads schema.Pattern then writes to compiledPatterns]
func (schema *Schema) compilePattern(c RegexCompilerFunc) (cp RegexMatcher, err error) {
	pattern := schema.Pattern
	if c != nil {
		cp, err = c(pattern)
	} else {
		cp, err = regexp.Compile(intoGoRegexp(pattern))
	}
	if err != nil {
		schemaErr := &SchemaError{
			Schema:      schema,
			SchemaField: "pattern",
			Origin:      err,
			Reason:      fmt.Sprintf("cannot compile pattern %q: %v", pattern, err),
		}
		// Wrap in a typed cluster (#1187 follow-on) so consumers can
		// detect the regex-compile failure specifically via errors.As
		// against SchemaPatternRegexError. errors.As against the
		// legacy SchemaError still works via the Unwrap chain,
		// preserving backward compatibility.
		err = newSchemaPatternRegexError(pattern, schemaErr, schema.Origin)
		return
	}

	compiledPatterns.Store(pattern, cp)
	return
}
