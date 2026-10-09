package openapi3

import (
	"maps"
	"testing"
)

// KeepGlobalFormats restores SchemaStringFormats, SchemaNumberFormats and SchemaIntegerFormats
// when t completes, so that the format validators a test defines do not leak into other tests.
func KeepGlobalFormats(t testing.TB) {
	t.Helper()
	stringFormats := maps.Clone(SchemaStringFormats)
	numberFormats := maps.Clone(SchemaNumberFormats)
	integerFormats := maps.Clone(SchemaIntegerFormats)
	t.Cleanup(func() {
		SchemaStringFormats = stringFormats
		SchemaNumberFormats = numberFormats
		SchemaIntegerFormats = integerFormats
	})
}
