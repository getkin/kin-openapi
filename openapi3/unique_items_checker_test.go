package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestRegisterArrayUniqueItemsChecker(t *testing.T) {
	var (
		schema = openapi3.Schema{
			Type:        &openapi3.Types{"array"},
			UniqueItems: true,
			Items:       openapi3.NewStringSchema().NewRef(),
		}
		val = []any{"1", "2", "3"}
	)

	// Fist checked by predefined function
	err := schema.VisitJSON(val)
	require.NoError(t, err)

	// Register a function will always return false when check if a
	// slice has unique items, then use a slice indeed has unique
	// items to verify that check unique items will failed.
	openapi3.RegisterArrayUniqueItemsChecker(func(items []any) bool {
		return false
	})
	defer openapi3.RegisterArrayUniqueItemsChecker(nil) // Reset for other tests

	err = schema.VisitJSON(val)
	require.Error(t, err)
	require.ErrorContains(t, err, "duplicate items found")
}
