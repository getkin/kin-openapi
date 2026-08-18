package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

// TestIssue1244UUIDRFC9562 documents which UUIDs each predefined format accepts.
// RFC 9562 obsoletes RFC 4122 and adds versions 6, 7 and 8 as well as the Max
// UUID, none of which are matched by FormatOfStringForUUIDOfRFC4122. The RFC 4122
// format is left untouched so existing users keep the stricter behaviour.
// See https://github.com/getkin/kin-openapi/issues/1244.
func TestIssue1244UUIDRFC9562(t *testing.T) {
	rfc4122 := openapi3.NewRegexpFormatValidator(openapi3.FormatOfStringForUUIDOfRFC4122)
	rfc9562 := openapi3.NewRegexpFormatValidator(openapi3.FormatOfStringForUUIDOfRFC9562)

	for _, tc := range []struct {
		name      string
		value     string
		okRFC4122 bool
		okRFC9562 bool
	}{
		// Covered by both formats.
		{"v1", "77e66540-ca29-11ed-afa1-0242ac120002", true, true},
		{"v4", "00f4d301-b9f4-4366-8907-2b5a03430aa1", true, true},
		{"v5", "630eb68f-e0fa-5ecc-887a-7c7a62614681", true, true},
		{"nil UUID", "00000000-0000-0000-0000-000000000000", true, true},

		// Covered by RFC 9562 only.
		{"v6", "1ec9414c-232a-6b00-b3c8-9e6bdeced846", false, true},
		{"v7", "017f22e2-79b0-7cc3-98c4-dc0c0c07398f", false, true},
		{"v8", "2489e9ad-2ee2-8e00-8ec9-32d5f69181c0", false, true},
		{"max UUID", "ffffffff-ffff-ffff-ffff-ffffffffffff", false, true},

		// Covered by neither format.
		{"not a uuid", "foo", false, false},
		{"version 0", "00f4d301-b9f4-0366-8907-2b5a03430aa1", false, false},
		{"version 9", "00f4d301-b9f4-9366-8907-2b5a03430aa1", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.okRFC4122 {
				require.NoError(t, rfc4122.Validate(tc.value))
			} else {
				require.Error(t, rfc4122.Validate(tc.value))
			}
			if tc.okRFC9562 {
				require.NoError(t, rfc9562.Validate(tc.value))
			} else {
				require.Error(t, rfc9562.Validate(tc.value))
			}
		})
	}
}
