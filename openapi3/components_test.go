package openapi3

import (
	"strings"
	"testing"
)

func TestValidateIdentifier(t *testing.T) {
	tt := []struct {
		name       string
		identifier string
		isValid    bool
		errMsg     string
	}{
		{name: "valid", identifier: "User-12_3.4", isValid: true, errMsg: ""},
		{name: "valid with []", identifier: "User[Admin]", isValid: true, errMsg: ""},
		{name: "invalid", identifier: "User Admin", isValid: false, errMsg: ""},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateIdentifier(tc.identifier)
			if err == nil && !tc.isValid {
				t.Errorf("test '%s' should have failed but didn't", tc.name)
			}
			if err != nil && tc.isValid {
				t.Errorf("test '%s' should have succeeded but failed with error:\n %v", tc.name, err)
			}

			if err != nil && !tc.isValid {
				if !strings.Contains(err.Error(), tc.errMsg) {
					t.Errorf("expected test '%s' to fail with\n'%s'\nbut got\n'%s'\n", tc.name, tc.errMsg, err.Error())
				}
			}
		})
	}
}
