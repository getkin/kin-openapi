package openapi2conv_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
)

func TestToV3ResponseExamples(t *testing.T) {
	for _, tt := range []struct {
		name     string
		response string
		produces []string
		want     string
	}{
		{
			name:     "schema and multiple media types",
			response: `{"description":"response","schema":{"type":"object"},"examples":{"application/json":{"status":"ok"},"application/xml":"<status>ok</status>"}}`,
			produces: []string{"application/json", "application/xml", "text/plain"},
			want:     `{"description":"response","content":{"application/json":{"schema":{"type":"object"},"example":{"status":"ok"}},"application/xml":{"schema":{"type":"object"},"example":"<status>ok</status>"},"text/plain":{"schema":{"type":"object"}}}}`,
		},
		{
			name:     "without schema",
			response: `{"description":"response","examples":{"text/plain":"hello"}}`,
			produces: []string{"text/plain"},
			want:     `{"description":"response","content":{"text/plain":{"example":"hello"}}}`,
		},
		{
			name:     "false example",
			response: `{"description":"response","schema":{"type":"boolean"},"examples":{"application/json":false}}`,
			want:     `{"description":"response","content":{"application/json":{"schema":{"type":"boolean"},"example":false}}}`,
		},
		{
			name:     "zero example",
			response: `{"description":"response","schema":{"type":"integer"},"examples":{"application/json":0}}`,
			want:     `{"description":"response","content":{"application/json":{"schema":{"type":"integer"},"example":0}}}`,
		},
		{
			name:     "empty string example",
			response: `{"description":"response","examples":{"text/plain":""}}`,
			produces: []string{"text/plain"},
			want:     `{"description":"response","content":{"text/plain":{"example":""}}}`,
		},
		{
			name:     "schema without examples",
			response: `{"description":"response","schema":{"type":"string"}}`,
			want:     `{"description":"response","content":{"application/json":{"schema":{"type":"string"}}}}`,
		},
		{
			name:     "without schema or examples",
			response: `{"description":"response"}`,
			want:     `{"description":"response"}`,
		},
		{
			name:     "reference",
			response: `{"$ref":"#/responses/Example"}`,
			want:     `{"$ref":"#/components/responses/Example"}`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var response openapi2.Response
			require.NoError(t, json.Unmarshal([]byte(tt.response), &response))
			converted, err := openapi2conv.ToV3Response(&response, tt.produces)
			require.NoError(t, err)
			data, err := json.Marshal(converted)
			require.NoError(t, err)
			require.JSONEq(t, tt.want, string(data))
		})
	}
}
