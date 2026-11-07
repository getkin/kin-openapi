package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestUnmarshalErrorNamesTheField(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name string
		spec string
		want string
	}{
		{
			name: "a scalar of the wrong type",
			spec: `
openapi: 3.0.1
info: {version: v1, title: t}
paths: {}
components:
  schemas:
    S:
      type: object
      minItems: not-a-number
`,
			want: "line 8: cannot construct !!str `not-a-n...` into uint64 in /components/schemas/S/minItems",
		},
		{
			name: "an item of a sequence, under a key holding a tilde",
			spec: `
openapi: 3.0.1
info: {version: v1, title: t}
paths: {}
components:
  schemas:
    a~b:
      type: object
      required:
        - ok
        - {not: a string}
`,
			want: "line 10: cannot construct !!map into string in /components/schemas/a~0b/required/1",
		},
		{
			name: "a path, whose slash is escaped",
			spec: `
openapi: 3.0.1
info: {version: v1, title: t}
paths:
  /p:
    get:
      responses:
        '200':
          description: ''
      parameters: 3
`,
			want: "line 9: cannot construct !!int `3` into openapi3.Parameters in /paths/~1p/get/parameters",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			loader := openapi3.NewLoader()
			_, err := loader.LoadFromData([]byte(tt.spec[1:]))
			require.ErrorContains(t, err, tt.want)
		})
	}
}
