package openapi3

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type versionExample struct {
	Title         string
	Values        []Version
	ExpectedValid bool
}

var versionExamples = []versionExample{
	{
		Title: "Nominal",
		Values: []Version{
			"3.0.1",
			"3.0.0",
			"3.1.0",
			"3.299.0",
		},
		ExpectedValid: true,
	},
	{
		Title: "Invalid",
		Values: []Version{
			"3.0.1.2",
			"3.0",
			"3",
			"3.0.1-pre1",
			"2.0.0",
			"2.0",
			"4.0.0",
		},
		ExpectedValid: false,
	},
}

func TestVersions(t *testing.T) {
	for _, example := range versionExamples {
		t.Run(example.Title, testVersion(t, example))
	}
}
func testVersion(t *testing.T, e versionExample) func(*testing.T) {
	testCtx := context.Background()
	return func(t *testing.T) {
		for _, value := range e.Values {
			if e.ExpectedValid {
				assert.NoErrorf(t, value.Validate(testCtx), "valid value: %v", value)
			} else {
				assert.Errorf(t, value.Validate(testCtx), "invalid value: %v", value)
			}
		}
	}
}

type minorVersionExampleValues struct {
	InputVersion         Version
	ExpectedMinorVersion uint64
}

type minorVersionExample struct {
	Title  string
	Values []minorVersionExampleValues
}

var minorVersionExamples = []minorVersionExample{
	{
		Title: "Nominal",
		Values: []minorVersionExampleValues{
			{InputVersion: "3.0.0", ExpectedMinorVersion: 0},
			{InputVersion: "3.1.0", ExpectedMinorVersion: 1},
			{InputVersion: "3.2.noway", ExpectedMinorVersion: 2},
			{InputVersion: "2.0.0", ExpectedMinorVersion: 0},
			{InputVersion: "4.0.0", ExpectedMinorVersion: 0},
		},
	},
	{
		Title: "Invalid",
		Values: []minorVersionExampleValues{
			{InputVersion: "xyz", ExpectedMinorVersion: 0},
			{InputVersion: "x.y.z", ExpectedMinorVersion: 0},
		},
	},
}

func TestMinorVersions(t *testing.T) {
	for _, example := range minorVersionExamples {
		t.Run(example.Title, testMinorVersion(t, example))
	}
}
func testMinorVersion(t *testing.T, e minorVersionExample) func(*testing.T) {
	return func(t *testing.T) {
		for _, value := range e.Values {
			assert.Equal(t, value.ExpectedMinorVersion, value.InputVersion.Minor(), value.InputVersion)
		}
	}
}
