package openapi3

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPathsMustStartWithSlash(t *testing.T) {
	spec := `
openapi: "3.0"
info:
  version: "1.0"
  title: sample
basePath: /adc/v1
paths:
  PATH:
    get:
      responses:
        200:
          description: description
`

	for path, expectedErr := range map[string]string{
		"foo/bar":  "invalid paths: path \"foo/bar\" does not start with a forward slash (/)",
		"/foo/bar": "",
	} {
		loader := NewSwaggerLoader()
		doc, err := loader.LoadSwaggerFromData([]byte(strings.Replace(spec, "PATH", path, 1)))
		require.NoError(t, err)
		err = doc.Validate(loader.Context)
		if expectedErr != "" {
			require.EqualError(t, err, expectedErr)
		} else {
			require.NoError(t, err)
		}
	}
}
