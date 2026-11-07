package yamlconv_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/internal/yamlconv"
)

// A scalar whose resolved type JSON cannot name arrives as the text the
// document wrote it with, so the document survives the conversion and keeps
// the member names it spells.
func TestUnmarshalRetagsResolvedScalars(t *testing.T) {
	for _, tt := range []struct {
		name string
		yaml string
		json string
	}{
		{
			name: "a date-shaped scalar keeps its text",
			yaml: "x-date: 2020-06-11\n",
			json: `{"x-date":"2020-06-11"}`,
		},
		{
			name: "and so does a date-time one",
			yaml: "x-ts: 2020-06-11T16:32:50-03:00\n",
			json: `{"x-ts":"2020-06-11T16:32:50-03:00"}`,
		},
		{
			name: "a tagged timestamp asks for a time and gets one",
			yaml: "x-date: !!timestamp 2020-06-11\n",
			json: `{"x-date":"2020-06-11T00:00:00Z"}`,
		},
		{
			name: "a key is named by the text that spells it",
			yaml: "x-keys: {18_24: a, 1.5: b, 200: c, 0x2A: d, .0: e}\n",
			json: `{"x-keys":{"18_24":"a","1.5":"b","200":"c","0x2A":"d",".0":"e"}}`,
		},
		{
			name: "at any depth",
			yaml: "x-deep: {a: {200: b}}\n",
			json: `{"x-deep":{"a":{"200":"b"}}}`,
		},
		{
			name: "a tagged key is named by its text as well",
			yaml: "x-keys: {!!timestamp 2020-06-11: a}\n",
			json: `{"x-keys":{"2020-06-11":"a"}}`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var got map[string]any
			require.NoError(t, yamlconv.Unmarshal([]byte(tt.yaml), &got))

			data, err := json.Marshal(got)
			require.NoError(t, err)
			require.JSONEq(t, tt.json, string(data))
		})
	}
}
