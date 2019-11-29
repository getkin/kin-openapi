package openapi3_test

import (
	"context"
	"errors"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func TestServerParamValuesWithPath(t *testing.T) {
	server := &openapi3.Server{
		URL: "http://{arg0}.{arg1}.example.com/a/{arg3}-version/{arg4}c{arg5}",
	}
	for input, expected := range map[string]*serverMatch{
		"http://x.example.com/a/b":                                    nil,
		"http://x.y.example.com/":                                     nil,
		"http://x.y.example.com/a/":                                   nil,
		"http://x.y.example.com/a/c":                                  nil,
		"http://baddomain.com/.example.com/a/1.0.0-version/c/d":       nil,
		"http://baddomain.com/.example.com/a/1.0.0/2/2.0.0-version/c": nil,
		"http://x.y.example.com/a/b-version/prefixedc":                newServerMatch("/", "x", "y", "b", "prefixed", ""),
		"http://x.y.example.com/a/b-version/c":                        newServerMatch("/", "x", "y", "b", "", ""),
		"http://x.y.example.com/a/b-version/c/":                       newServerMatch("/", "x", "y", "b", "", ""),
		"http://x.y.example.com/a/b-version/c/d":                      newServerMatch("/d", "x", "y", "b", "", ""),
		"http://domain0.domain1.example.com/a/b-version/c/d":          newServerMatch("/d", "domain0", "domain1", "b", "", ""),
		"http://domain0.domain1.example.com/a/1.0.0-version/c/d":      newServerMatch("/d", "domain0", "domain1", "1.0.0", "", ""),
	} {
		t.Run(input, testServerParamValues(t, server, input, expected))
	}
}

func TestServerParamValuesNoPath(t *testing.T) {
	server := &openapi3.Server{
		URL: "https://{arg0}.{arg1}.example.com/",
	}
	for input, expected := range map[string]*serverMatch{
		"https://domain0.domain1.example.com/": newServerMatch("/", "domain0", "domain1"),
	} {
		t.Run(input, testServerParamValues(t, server, input, expected))
	}
}

func validServerWithParameterInURL() *openapi3.Server {
	return &openapi3.Server{
		URL: "http://{Param1}.my.cool.{Param2}.website",
	}
}

func validServerWithoutParameterInURL() *openapi3.Server {
	return &openapi3.Server{
		URL: "http://my.cool.website",
	}
}

func invalidServerWithoutUrl() *openapi3.Server {
	return &openapi3.Server{}
}

func invalidServerWithErrorParameterDefinedInUrl() *openapi3.Server {
	return &openapi3.Server{
		URL: "http://{my.cool.website",
	}
}

func TestServerValidation(t *testing.T) {
	tests := []struct {
		name                       string
		input                      *openapi3.Server
		expectedError              error
		expectedVariableNamesInURL []string
	}{
		{
			"when no URL is provided",
			invalidServerWithoutUrl(),
			errors.New("Variable 'URL' must be a non-empty JSON string"),
			nil,
		},
		{
			"when URL is provided with error parameter definition",
			invalidServerWithErrorParameterDefinedInUrl(),
			errors.New("URL 'http://{my.cool.website' Missing '}'"),
			nil,
		},
		{
			"when a URL is provided with parameter in 'url'",
			validServerWithParameterInURL(),
			nil,
			[]string{"Param1", "Param2"},
		},
		{
			"when a URL is provided without parameter in 'url'",
			validServerWithoutParameterInURL(),
			nil,
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := context.Background()
			validationErr := test.input.Validate(c)

			require.Equal(t, test.expectedError, validationErr, "expected errors (or lack of) to match")
			require.Equal(t, test.expectedVariableNamesInURL, test.input.VariableNamesInURL)
		})
	}
}

func testServerParamValues(t *testing.T, server *openapi3.Server, input string, expected *serverMatch) func(*testing.T) {
	return func(t *testing.T) {
		args, remaining, ok := server.MatchRawURL(input)
		if expected == nil {
			require.False(t, ok)
			return
		}
		require.True(t, ok)

		actual := &serverMatch{
			Remaining: remaining,
			Args:      args,
		}
		require.Equal(t, expected, actual)
	}
}

type serverMatch struct {
	Remaining string
	Args      []string
}

func newServerMatch(remaining string, args ...string) *serverMatch {
	return &serverMatch{
		Remaining: remaining,
		Args:      args,
	}
}
