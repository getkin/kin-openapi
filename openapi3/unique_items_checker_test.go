package openapi3_test

import (
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func TestRegisterArrayUniqueItemsChecker(t *testing.T) {
	var (
		schema = openapi3.Schema{
			Type:        "array",
			UniqueItems: true,
			Items:       openapi3.NewStringSchema().NewRef(),
		}
		val = []interface{}{"1", "2", "3"}
	)

	// Fist checked by predefined function
	err := schema.VisitJSON(val)
	require.NoError(t, err)

	// Register a function will always return false when check if a
	// slice has unique items, then use a slice indeed has unique
	// items to verify that check unique items will failed.
	openapi3.RegisterArrayUniqueItemsChecker(func(items []interface{}) bool {
		return false
	})
	defer openapi3.RegisterArrayUniqueItemsChecker(nil) // Reset for other tests

	err = schema.VisitJSON(val)
	require.Error(t, err)
	require.True(t, strings.HasPrefix(err.Error(), "Duplicate items found"))
}
