package openapi3

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExternalDocs_Validate(t *testing.T) {
	tests := []struct {
		name        string
		extDocs     *ExternalDocs
		expectedErr string
	}{
		{
			name:        "url is missing",
			extDocs:     &ExternalDocs{},
			expectedErr: "url is required",
		},
		{
			name:        "url is incorrect",
			extDocs:     &ExternalDocs{URL: "ht tps://example.com"},
			expectedErr: `url is incorrect: parse "ht tps://example.com": first path segment in URL cannot contain colon`,
		},
		{
			name:    "ok",
			extDocs: &ExternalDocs{URL: "https://example.com"},
		},
	}
	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			err := tt.extDocs.Validate(context.Background())
			if tt.expectedErr != "" {
				require.EqualError(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
