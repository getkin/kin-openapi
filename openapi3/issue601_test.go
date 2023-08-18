package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIssue601(t *testing.T) {
	// Document is invalid: first validation error returned is because
	//     schema:
	//       example: {key: value}
	// is not how schema examples are defined (but how components' examples are defined. Components are maps.)
	// Correct code should be:
	//     schema: {example: value}
	sl := NewLoader()
	doc, err := sl.LoadFromFile("testdata/lxkns.yaml")
	require.NoError(t, err)

	err = doc.Validate(sl.Context)
	require.ErrorContains(t, err, `invalid components: schema "DiscoveryResult": invalid example: Error at "/type": property "type" is missing`)
	require.ErrorContains(t, err, `| Error at "/nsid": property "nsid" is missing`)

	err = doc.Validate(sl.Context, DisableExamplesValidation())
	require.NoError(t, err)

	// Now let's remove all the invalid parts
	for _, schema := range doc.Components.Schemas {
		schema.Value.Example = nil
	}

	err = doc.Validate(sl.Context)
	require.NoError(t, err)
}
