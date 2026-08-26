package openapi3

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	goyaml "go.yaml.in/yaml/v3"
)

const nativeSrc = `"200":
  description: ok
  x-tags:
    - alpha
    - beta
  content:
    application/json:
      x-media: 1
"404":
  $ref: '#/components/responses/NotFound'
  summary: missing
x-collection: top
`

// A document decoded from the node must equal the same document decoded as
// JSON: the two paths are interchangeable for content.
func TestNativeStock_MatchesJSONPath(t *testing.T) {
	var viaJSON Responses
	jsonBytes, err := yamlToJSON(nativeSrc)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(jsonBytes, &viaJSON))

	var viaNode Responses
	require.NoError(t, goyaml.Unmarshal([]byte(nativeSrc), &viaNode))

	want, err := json.Marshal(&viaJSON)
	require.NoError(t, err)
	got, err := json.Marshal(&viaNode)
	require.NoError(t, err)
	require.JSONEq(t, string(want), string(got))
}

// Origins must reach nested collections, not only the top level.
func TestNativeStock_OriginsAtDepth(t *testing.T) {
	defer func(v bool) { originEnabledVar = v }(originEnabledVar)
	originEnabledVar = true

	var r Responses
	require.NoError(t, goyaml.Unmarshal([]byte(nativeSrc), &r))
	mt := r.Value("200").Value.Content["application/json"]
	require.NotNil(t, mt)
	require.NotNil(t, mt.Origin, "nested media type should carry an origin")
	require.NotNil(t, mt.Origin.Key)
	require.Equal(t, "application/json", mt.Origin.Key.Name)
	require.Equal(t, 7, mt.Origin.Key.Line)
}

func yamlToJSON(src string) ([]byte, error) {
	var v any
	if err := goyaml.Unmarshal([]byte(src), &v); err != nil {
		return nil, err
	}
	return json.Marshal(v)
}
