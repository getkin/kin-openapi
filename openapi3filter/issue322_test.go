//go:build go1.27 && goexperiment.jsonv2

// Build-tagged go1.27 because duplicate names are rejected only with encoding/json/v2.

package openapi3filter

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestIssue322(t *testing.T) {
	t.Parallel()

	schema := openapi3.NewObjectSchema().WithProperty("role", openapi3.NewStringSchema().WithEnum("user"))
	body := &openapi3.RequestBody{Content: openapi3.NewContentWithJSONSchema(schema)}
	for _, tc := range []struct {
		name    string
		body    string
		wantErr bool
	}{
		{name: "unique names", body: `{"role":"user","items":[{"a":1},{"a":2}]}`},
		{name: "duplicate name", body: `{"role":"admin","role":"user"}`, wantErr: true},
		{name: "duplicate escaped name", body: `{"role":"admin","\u0072ole":"user"}`, wantErr: true},
		{name: "duplicate in array item", body: `{"items":[{"a":1,"a":2}]}`, wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "/test", strings.NewReader(tc.body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			err = ValidateRequestBody(t.Context(), &RequestValidationInput{Request: req}, body)
			if tc.wantErr {
				require.ErrorContains(t, err, "duplicate object member name")
			} else {
				require.NoError(t, err)
			}
		})
	}
}
