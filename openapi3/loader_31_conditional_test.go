package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveConditionalSchemaRefs(t *testing.T) {
	loader := NewLoader()
	doc, err := loader.LoadFromFile("testdata/schema31_conditional.yml")
	require.NoError(t, err)

	err = doc.Validate(loader.Context)
	require.NoError(t, err)

	// Verify if/then/else refs are resolved
	conditional := doc.Components.Schemas["ConditionalField"].Value
	require.NotNil(t, conditional.If)
	require.NotNil(t, conditional.If.Value)
	require.True(t, conditional.If.Value.Type.Is("string"))

	require.NotNil(t, conditional.Then)
	require.NotNil(t, conditional.Then.Value)
	require.Equal(t, uint64(3), conditional.Then.Value.MinLength)

	require.NotNil(t, conditional.Else)
	require.NotNil(t, conditional.Else.Value)
	require.True(t, conditional.Else.Value.Type.Is("number"))

	// Verify dependentRequired is loaded
	payment := doc.Components.Schemas["PaymentInfo"].Value
	require.Equal(t, map[string][]string{
		"creditCard": {"billingAddress"},
	}, payment.DependentRequired)
}
