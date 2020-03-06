package openapi3

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSecurityRequirementsEncoding(t *testing.T) {

	tests := []struct {
		requirements *SecurityRequirements
		json         string
	}{
		{
			requirements: NewSecurityRequirements(),
			json:         `[]`,
		},
		{
			requirements: NewSecurityRequirements().With(NewSecurityRequirement()),
			json:         `[{}]`,
		},
	}

	for _, test := range tests {

		b, err := json.Marshal(test.requirements)
		require.NoError(t, err)
		require.Equal(t, test.json, string(b), "incorrect requirements encoding")
	}
}

func TestSecurityRequirementEncoding(t *testing.T) {

	tests := []struct {
		requirement SecurityRequirement
		json        string
	}{
		{
			requirement: NewSecurityRequirement(),
			json:        `{}`,
		},
		{
			requirement: NewSecurityRequirement().Authenticate("provider"),
			json:        `{"provider":[]}`,
		},
		{
			requirement: NewSecurityRequirement().Authenticate("provider", "scope1"),
			json:        `{"provider":["scope1"]}`,
		},
		{
			requirement: NewSecurityRequirement().Authenticate("provider", "scope1", "scope2"),
			json:        `{"provider":["scope1","scope2"]}`,
		},
	}

	for _, test := range tests {

		b, err := json.Marshal(test.requirement)
		require.NoError(t, err)
		require.Equal(t, test.json, string(b), "incorrect requirements encoding")
	}
}
