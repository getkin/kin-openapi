package openapi3

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIssue430(t *testing.T) {
	schema := NewOneOfSchema(
		NewStringSchema().WithFormat("ipv4"),
		NewStringSchema().WithFormat("ipv6"),
	)

	err := schema.Validate(context.Background())
	require.NoError(t, err)

	data := map[string]bool{
		"127.0.1.1": true,

		// https://stackoverflow.com/a/48519490/1418165

		// v4
		"192.168.0.1": true,
		// "192.168.0.1:80" doesn't parse per net.ParseIP()

		// v6
		"::FFFF:C0A8:1":                        false,
		"::FFFF:C0A8:0001":                     false,
		"0000:0000:0000:0000:0000:FFFF:C0A8:1": false,
		// "::FFFF:C0A8:1%1" doesn't parse per net.ParseIP()
		"::FFFF:192.168.0.1": false,
		// "[::FFFF:C0A8:1]:80" doesn't parse per net.ParseIP()
		// "[::FFFF:C0A8:1%1]:80" doesn't parse per net.ParseIP()
	}

	for datum := range data {
		err = schema.VisitJSON(datum)
		require.Error(t, err, ErrOneOfConflict.Error())
	}

	for datum, isV4 := range data {
		err = schema.VisitJSON(datum, SetOpenAPIMinorVersion(1))
		require.NoError(t, err)
		if isV4 {
			require.Nil(t, validateIPv4(datum), "%q should be IPv4", datum)
			require.NotNil(t, validateIPv6(datum), "%q should not be IPv6", datum)
		} else {
			require.NotNil(t, validateIPv4(datum), "%q should not be IPv4", datum)
			require.Nil(t, validateIPv6(datum), "%q should be IPv6", datum)
		}
	}
}

func TestFormatCallback_WrapError(t *testing.T) {
	var errSomething = errors.New("something error")

	DefineStringFormatCallback("foobar", func(value string) error {
		return errSomething
	})

	s := &Schema{Format: "foobar"}
	err := s.VisitJSONString("blablabla")

	assert.ErrorIs(t, err, errSomething)

	delete(SchemaStringFormats, "foobar")
}

func TestReversePathInMessageSchemaError(t *testing.T) {
	SchemaErrorDetailsDisabled = true

	const spc = `
components:
  schemas:
    Something:
      type: object
      properties:
        ip:
          type: string
          format: ipv4
`
	l := NewLoader()

	doc, err := l.LoadFromData([]byte(spc))
	require.NoError(t, err)

	err = doc.Components.Schemas["Something"].Value.VisitJSON(map[string]interface{}{
		`ip`: `123.0.0.11111`,
	}, SetOpenAPIMinorVersion(1))

	// assert, do not require to ensure SchemaErrorDetailsDisabled can be set to false
	assert.ErrorContains(t, err, `Error at "/ip"`)

	SchemaErrorDetailsDisabled = false
}

func TestUuidFormat(t *testing.T) {

	type testCase struct {
		name    string
		value   string
		wantErr bool
	}

	DefineStringFormat("uuid", FormatOfStringForUUIDOfRFC4122)
	defer RestoreDefaultStringFormats()
	testCases := []testCase{
		{
			name:    "invalid",
			value:   "foo",
			wantErr: true,
		},
		{
			name:    "uuid v1",
			value:   "77e66540-ca29-11ed-afa1-0242ac120002",
			wantErr: false,
		},
		{
			name:    "uuid v4",
			value:   "00f4d301-b9f4-4366-8907-2b5a03430aa1",
			wantErr: false,
		},
		{
			name:    "uuid nil",
			value:   "00000000-0000-0000-0000-000000000000",
			wantErr: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := NewUUIDSchema().VisitJSON(tc.value)
			var schemaError = &SchemaError{}
			if tc.wantErr {
				require.Error(t, err)
				require.ErrorAs(t, err, &schemaError)

				require.NotZero(t, schemaError.Reason)
				require.NotContains(t, schemaError.Reason, fmt.Sprint(tc.value))
			} else {
				require.Nil(t, err)
			}
		})
	}
}
func TestStringFormatsStartingWithOpenAPIMinorVersion(t *testing.T) {
	DefineStringFormat("test", "test0", FromOpenAPIMinorVersion(0))
	defer RestoreDefaultStringFormats()
	for i := uint64(0); i < 10; i++ {
		if assert.Contains(t, SchemaStringFormats, "test") &&
			assert.NotNilf(t, SchemaStringFormats["test"].get(i), "%d", i) {
			if assert.NotNilf(t, SchemaStringFormats["test"].get(i).regexp, "%d", i) {
				assert.Equalf(t, "test0", SchemaStringFormats["test"].get(i).regexp.String(), "%d", i)
			}
			assert.Nilf(t, SchemaStringFormats["test"].get(i).callback, "%d", i)
		}
	}

	DefineStringFormat("test", "test1", FromOpenAPIMinorVersion(1))
	for i := uint64(0); i < 10; i++ {
		if assert.Contains(t, SchemaStringFormats, "test") &&
			assert.NotNilf(t, SchemaStringFormats["test"].get(i), "%d", i) {
			if assert.NotNilf(t, SchemaStringFormats["test"].get(i).regexp, "%d", i) {
				var regexpString string
				switch {
				case i == 0:
					regexpString = "test0"
				case i > 0:
					regexpString = "test1"
				}
				assert.Equalf(t, regexpString, SchemaStringFormats["test"].get(i).regexp.String(), "%d", i)
			}
			assert.Nilf(t, SchemaStringFormats["test"].get(i).callback, "%d", i)
		}
	}

	DefineStringFormat("test", "test5", FromOpenAPIMinorVersion(5))
	for i := uint64(0); i < 10; i++ {
		if assert.Contains(t, SchemaStringFormats, "test") &&
			assert.NotNilf(t, SchemaStringFormats["test"].get(i), "%d", i) {
			if assert.NotNilf(t, SchemaStringFormats["test"].get(i).regexp, "%d", i) {
				var regexpString string
				switch {
				case i == 0:
					regexpString = "test0"
				case i >= 1 && i < 5:
					regexpString = "test1"
				case i >= 5:
					regexpString = "test5"
				}
				assert.Equalf(t, regexpString, SchemaStringFormats["test"].get(i).regexp.String(), "%d", i)
			}
			assert.Nilf(t, SchemaStringFormats["test"].get(i).callback, "%d", i)
		}
	}

	DefineStringFormat("test", "test2", FromOpenAPIMinorVersion(2))
	for i := uint64(0); i < 10; i++ {
		if assert.Contains(t, SchemaStringFormats, "test") &&
			assert.NotNilf(t, SchemaStringFormats["test"].get(i), "%d", i) {
			if assert.NotNilf(t, SchemaStringFormats["test"].get(i).regexp, "%d", i) {
				var regexpString string
				switch {
				case i == 0:
					regexpString = "test0"
				case i == 1:
					regexpString = "test1"
				case i >= 2:
					regexpString = "test2"
				}
				assert.Equalf(t, regexpString, SchemaStringFormats["test"].get(i).regexp.String(), "%d", i)
			}
			assert.Nilf(t, SchemaStringFormats["test"].get(i).callback, "%d", i)
		}
	}

	DefineStringFormat("test", "test4", FromOpenAPIMinorVersion(4))
	for i := uint64(0); i < 10; i++ {
		if assert.Contains(t, SchemaStringFormats, "test") &&
			assert.NotNilf(t, SchemaStringFormats["test"].get(i), "%d", i) {
			if assert.NotNilf(t, SchemaStringFormats["test"].get(i).regexp, "%d", i) {
				var regexpString string
				switch {
				case i == 0:
					regexpString = "test0"
				case i == 1:
					regexpString = "test1"
				case i >= 2 && i < 4:
					regexpString = "test2"
				case i >= 4:
					regexpString = "test4"
				}
				assert.Equalf(t, regexpString, SchemaStringFormats["test"].get(i).regexp.String(), "%d", i)
			}
			assert.Nilf(t, SchemaStringFormats["test"].get(i).callback, "%d", i)
		}
	}

	DefineStringFormat("test", "test3", FromOpenAPIMinorVersion(3))
	for i := uint64(0); i < 10; i++ {
		if assert.Contains(t, SchemaStringFormats, "test") &&
			assert.NotNilf(t, SchemaStringFormats["test"].get(i), "%d", i) {
			if assert.NotNilf(t, SchemaStringFormats["test"].get(i).regexp, "%d", i) {
				var regexpString string
				switch {
				case i == 0:
					regexpString = "test0"
				case i == 1:
					regexpString = "test1"
				case i == 2:
					regexpString = "test2"
				case i >= 3:
					regexpString = "test3"
				}
				assert.Equalf(t, regexpString, SchemaStringFormats["test"].get(i).regexp.String(), "%d", i)
			}
			assert.Nilf(t, SchemaStringFormats["test"].get(i).callback, "%d", i)
		}
	}

	DefineStringFormat("test", "test7", FromOpenAPIMinorVersion(7))
	for i := uint64(0); i < 10; i++ {
		if assert.Contains(t, SchemaStringFormats, "test") &&
			assert.NotNilf(t, SchemaStringFormats["test"].get(i), "%d", i) {
			if assert.NotNilf(t, SchemaStringFormats["test"].get(i).regexp, "%d", i) {
				var regexpString string
				switch {
				case i == 0:
					regexpString = "test0"
				case i == 1:
					regexpString = "test1"
				case i == 2:
					regexpString = "test2"
				case i >= 3 && i < 7:
					regexpString = "test3"
				case i >= 7:
					regexpString = "test7"
				}
				assert.Equalf(t, regexpString, SchemaStringFormats["test"].get(i).regexp.String(), "%d", i)
			}
			assert.Nilf(t, SchemaStringFormats["test"].get(i).callback, "%d", i)
		}
	}

	DefineStringFormat("test", "testnew")
	for i := uint64(0); i < 10; i++ {
		if assert.Contains(t, SchemaStringFormats, "test") &&
			assert.NotNilf(t, SchemaStringFormats["test"].get(i), "%d", i) {
			if assert.NotNilf(t, SchemaStringFormats["test"].get(i).regexp, "%d", i) {
				assert.Equalf(t, "testnew", SchemaStringFormats["test"].get(i).regexp.String(), "%d", i)
			}
			assert.Nilf(t, SchemaStringFormats["test"].get(i).callback, "%d", i)
		}
	}
}

func createCallBackError(minorVersion uint) error {
	return fmt.Errorf("%d", minorVersion)
}
func createCallBack(callbackError error) FormatCallback {
	return func(name string) error {
		return callbackError
	}
}

func TestStringFormatsCallbackStartingWithOpenAPIMinorVersion(t *testing.T) {
	defer RestoreDefaultStringFormats()
	callbackError0 := createCallBackError(0)
	DefineStringFormatCallback("testCallback", createCallBack(callbackError0))
	for i := uint64(0); i < 10; i++ {
		if assert.NotNilf(t, SchemaStringFormats["testCallback"], "%d", i) &&
			assert.NotNilf(t, SchemaStringFormats["testCallback"].get(i), "%d", i) {
			assert.Emptyf(t, SchemaStringFormats["testCallback"].get(i).regexp, "%d", i)
			assert.Equal(t, callbackError0, SchemaStringFormats["testCallback"].get(i).callback("ignored"), "%d", i)
		}
	}

	callbackError1 := createCallBackError(1)
	DefineStringFormatCallback("testCallback", createCallBack(callbackError1), FromOpenAPIMinorVersion(1))
	for i := uint64(0); i < 10; i++ {
		if assert.NotNilf(t, SchemaStringFormats["testCallback"], "%d", i) &&
			assert.NotNilf(t, SchemaStringFormats["testCallback"].get(i), "%d", i) {
			assert.Emptyf(t, SchemaStringFormats["testCallback"].get(i).regexp, "%d", i)
			var err error
			switch {
			case i == 0:
				err = callbackError0
			case i > 0:
				err = callbackError1
			}
			assert.Equal(t, err, SchemaStringFormats["testCallback"].get(i).callback("ignored"), "%d", i)
		}
	}
	callbackError5 := createCallBackError(5)
	DefineStringFormatCallback("testCallback", createCallBack(callbackError5), FromOpenAPIMinorVersion(5))
	assert.Equal(t, 6, len(SchemaStringFormats["testCallback"].versionedFormats))
	for i := uint64(0); i < 10; i++ {
		if assert.NotNilf(t, SchemaStringFormats["testCallback"], "%d", i) &&
			assert.NotNilf(t, SchemaStringFormats["testCallback"].get(i), "%d", i) {
			assert.Emptyf(t, SchemaStringFormats["testCallback"].get(i).regexp, "%d", i)
			var err error
			switch {
			case i == 0:
				err = callbackError0
			case i > 0 && i < 5:
				err = callbackError1
			case i >= 5:
				err = callbackError5
			}
			assert.Equal(t, err, SchemaStringFormats["testCallback"].get(i).callback("ignored"), "%d", i)
		}
	}
	callbackError3 := createCallBackError(3)
	DefineStringFormatCallback("testCallback", createCallBack(callbackError3), FromOpenAPIMinorVersion(3))
	assert.Equal(t, 6, len(SchemaStringFormats["testCallback"].versionedFormats))
	for i := uint64(0); i < 10; i++ {
		if assert.NotNilf(t, SchemaStringFormats["testCallback"], "%d", i) &&
			assert.NotNilf(t, SchemaStringFormats["testCallback"].get(i), "%d", i) {
			assert.Emptyf(t, SchemaStringFormats["testCallback"].get(i).regexp, "%d", i)
			var err error
			switch {
			case i == 0:
				err = callbackError0
			case i > 0 && i < 3:
				err = callbackError1
			case i >= 3:
				err = callbackError3
			}
			assert.Equal(t, err, SchemaStringFormats["testCallback"].get(i).callback("ignored"), "%d", i)
		}
	}
	callbackError4 := createCallBackError(4)
	DefineStringFormatCallback("testCallback", createCallBack(callbackError4), FromOpenAPIMinorVersion(4))
	assert.Equal(t, 6, len(SchemaStringFormats["testCallback"].versionedFormats))
	for i := uint64(0); i < 10; i++ {
		if assert.NotNilf(t, SchemaStringFormats["testCallback"], "%d", i) &&
			assert.NotNilf(t, SchemaStringFormats["testCallback"].get(i), "%d", i) {
			assert.Emptyf(t, SchemaStringFormats["testCallback"].get(i).regexp, "%d", i)
			var err error
			switch {
			case i == 0:
				err = callbackError0
			case i > 0 && i < 3:
				err = callbackError1
			case i == 3:
				err = callbackError3
			case i >= 4:
				err = callbackError4
			}
			assert.Equal(t, err, SchemaStringFormats["testCallback"].get(i).callback("ignored"), "%d", i)
		}
	}
	callbackError99 := createCallBackError(99)
	DefineStringFormatCallback("testCallback", createCallBack(callbackError99))
	assert.Equal(t, 6, len(SchemaStringFormats["testCallback"].versionedFormats))
	for i := uint64(0); i < 10; i++ {
		if assert.NotNilf(t, SchemaStringFormats["testCallback"], "%d", i) &&
			assert.NotNilf(t, SchemaStringFormats["testCallback"].get(i), "%d", i) {
			assert.Emptyf(t, SchemaStringFormats["testCallback"].get(i).regexp, "%d", i)
			assert.Equal(t, callbackError99, SchemaStringFormats["testCallback"].get(i).callback("ignored"), "%d", i)
		}
	}
}
