package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestValidateNullEntries(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name, spec, err string
	}{
		{
			name: "server",
			spec: "servers:\n- \n",
			err:  `invalid servers: value of url must be a non-empty string`,
		},
		{
			name: "server variable",
			spec: "servers:\n- url: 'http://x/{v}'\n  variables:\n    v:\n",
			err:  `invalid servers: field default is required in {}`,
		},
		{
			name: "tag",
			spec: "tags:\n- \n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			loader := openapi3.NewLoader()
			doc, err := loader.LoadFromData([]byte("openapi: 3.0.0\ninfo: {title: t, version: '1'}\npaths: {}\n" + tc.spec))
			require.NoError(t, err)
			err = doc.Validate(loader.Context)
			if tc.err == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.err)
			}
		})
	}

	t.Run("component", func(t *testing.T) {
		t.Parallel()

		components := openapi3.Components{Links: openapi3.Links{"a": nil}}
		err := components.Validate(t.Context())
		require.EqualError(t, err, `link "a": found unresolved ref: ""`)
	})
}
