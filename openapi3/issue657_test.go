package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestOneOf_Warning_Errors(t *testing.T) {
	t.Parallel()

	loader := openapi3.NewLoader()
	spec := `
components:
  schemas:
    Something:
      type: object
      properties:
        field:
          title: Some field
          oneOf:
            - title: First rule
              type: string
              minLength: 10
              maxLength: 10
            - title: Second rule
              type: string
              minLength: 15
              maxLength: 15
`[1:]

	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)

	tests := [...]struct {
		name     string
		value    string
		checkErr require.ErrorAssertionFunc
	}{
		{
			name:     "valid value",
			value:    "ABCDE01234",
			checkErr: require.NoError,
		},
		{
			name:     "valid value",
			value:    "ABCDE0123456789",
			checkErr: require.NoError,
		},
		{
			name:  "no valid value",
			value: "ABCDE",
			checkErr: func(t require.TestingT, err error, i ...interface{}) {
				require.ErrorContains(t, err, "doesn't match schema due to: minimum string length is 10")

				wErr := &openapi3.MultiError{}
				require.ErrorAs(t, err, wErr)

				require.Len(t, *wErr, 2)

				require.Equal(t, "minimum string length is 10", (*wErr)[0].(*openapi3.SchemaError).Reason)
				require.Equal(t, "minimum string length is 15", (*wErr)[1].(*openapi3.SchemaError).Reason)
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err = doc.Components.Schemas["Something"].Value.Properties["field"].Value.VisitJSON(test.value)

			test.checkErr(t, err)
		})
	}
}
