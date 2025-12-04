package openapi2

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnmarshalError(t *testing.T) {
	{
		v2 := []byte(`
openapi: '2.0'
info:
  version: '1.10'
  title: title
paths:
  "/ping":
    post:
      consumes:
      - multipart/form-data
      parameters:
        name: file  # <-- Missing dash
        in: formData
        description: file
        required: true
        type: file
      responses:
        '200':
          description: OK
`[1:])

		var doc T
		err := unmarshal(v2, &doc)
		require.ErrorContains(t, err, `json: cannot unmarshal object into field Operation.parameters of type openapi2.Parameters`)
	}

	v2 := []byte(`
openapi: '2.0'
info:
  version: '1.10'
  title: title
paths:
  "/ping":
    post:
      consumes:
      - multipart/form-data
      parameters:
      - name: file  # <--
        in: formData
        description: file
        required: true
        type: file
      responses:
        '200':
          description: OK
`[1:])

	var doc T
	err := unmarshal(v2, &doc)
	require.NoError(t, err)
}
