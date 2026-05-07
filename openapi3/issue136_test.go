package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestIssue136(t *testing.T) {
	specf := func(dflt string) string {
		return `
openapi: 3.0.2
info:
  title: "Hello World REST APIs"
  version: "1.0"
paths: {}
components:
  schemas:
    SomeSchema:
      type: string
      default: ` + dflt + `
`
	}

	for _, testcase := range []struct {
		dflt, err string
	}{
		{
			dflt: `"foo"`,
			err:  "",
		},
		{
			dflt: `1`,
			err:  "invalid components: invalid schema default: value must be a string",
		},
	} {
		t.Run(testcase.dflt, func(t *testing.T) {
			spec := specf(testcase.dflt)

			sl := openapi3.NewLoader()

			doc, err := sl.LoadFromData([]byte(spec))
			require.NoError(t, err)

			err = doc.Validate(sl.Context)
			if testcase.err == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err, testcase.err)
			}
		})
	}
}
