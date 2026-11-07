package openapi2

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnmarshalError(t *testing.T) {
	t.Parallel()

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
		if UsingJSONv2 {
			require.ErrorContains(t, err, `json: cannot unmarshal object into Go struct field T.paths.~1ping.post.parameters of type openapi2.Parameters`)
		} else {
			require.ErrorContains(t, err, `json: cannot unmarshal object into field Operation.parameters of type openapi2.Parameters`)
		}
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

func TestMayBeJSON(t *testing.T) {
	t.Parallel()

	for data, expected := range map[string]bool{
		``:               false,
		" \t\r\n":        false,
		`swagger: '2.0'`: false,
		`# comment`:      false,
		"\ufeff{}":       false,
		` {}`:            true,
		"\n[]":           true,
		`"s"`:            true,
		`-1`:             true,
		`---`:            true,
		`42`:             true,
		`true`:           true,
		`false`:          true,
		`null`:           true,
	} {
		require.Equal(t, expected, mayBeJSON([]byte(data)), "%q", data)
	}
}

func TestUnmarshalErrorWrapsTheDecodersTried(t *testing.T) {
	t.Parallel()

	var doc T
	err := unmarshal([]byte("swagger: '2.0'\ninfo: 42\n"), &doc)
	require.ErrorContains(t, err, "failed to unmarshal data: yaml error: ")
	msg := err.Error()
	require.NotContains(t, msg, "json error")
	require.Error(t, errors.Unwrap(err))

	err = unmarshal([]byte(`{"swagger": "2.0", "info": 42}`), &doc)
	require.ErrorContains(t, err, "failed to unmarshal data: json error: ")
	require.ErrorContains(t, err, ", yaml error: ")
	var wrapped interface{ Unwrap() []error }
	require.ErrorAs(t, err, &wrapped)
	require.Len(t, wrapped.Unwrap(), 2)
}
