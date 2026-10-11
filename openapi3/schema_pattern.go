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
// settings may be nil, as when validating the schema itself.
func (schema *Schema) compilePattern(settings *schemaValidationSettings, c RegexCompilerFunc) (cp RegexMatcher, err error) {
	pattern := schema.Pattern
	if c != nil {
		cp, err = c(pattern)
	} else {
		cp, err = regexp.Compile(intoGoRegexp(pattern))
	}
	if err != nil {
		schemaErr := settings.newSchemaError(schema, nil).
			withField("pattern").
			withReason(fmt.Sprintf("cannot compile pattern %q: %v", pattern, err)).
			withOrigin(err)
		// A failed compile can yield a typed nil, which no call site's nil check catches.
		cp = nil
		err = newSchemaPatternRegexError(pattern, schemaErr, schema.Origin)
		return
	}

	compiledPatterns.Store(pattern, cp)
	return
}
