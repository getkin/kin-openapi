package openapi2conv

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

// TestFromV3CustomMethods checks that converting an OpenAPI 3.2 document using
// methods Swagger 2.0 cannot express returns an error instead of panic-ing in
// openapi2's SetOperation.
func TestFromV3CustomMethods(t *testing.T) {
	const specFmt = `
openapi: 3.2.0
info:
  title: MyAPI
  version: "0.1"
paths:
  /pets:
    %s
components: {}
`

	for _, tc := range []struct {
		name    string
		snippet string
		method  string
	}{
		{
			name: "query field",
			snippet: `query:
      responses:
        "200":
          description: OK`,
			method: openapi3.MethodQuery,
		},
		{
			name: "additionalOperations",
			snippet: `additionalOperations:
      COPY:
        responses:
          "200":
            description: OK`,
			method: "COPY",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			loader := openapi3.NewLoader()
			doc3, err := loader.LoadFromData([]byte(fmt.Sprintf(specFmt, tc.snippet)))
			require.NoError(t, err)
			require.NoError(t, doc3.Validate(loader.Context))

			require.NotPanics(t, func() {
				_, err = FromV3(doc3)
			})
			require.ErrorContains(t, err, tc.method)
			require.ErrorContains(t, err, "OpenAPI 2.0")

			pathItem := doc3.Paths.Value("/pets")
			require.NotPanics(t, func() {
				_, err = FromV3PathItem(doc3, pathItem)
			})
			require.ErrorContains(t, err, tc.method)
		})
	}
}
