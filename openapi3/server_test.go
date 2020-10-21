package openapi3

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServerParamNames(t *testing.T) {
	server := &Server{
		URL: "http://{x}.{y}.example.com",
	}
	values, err := server.ParameterNames()
	require.NoError(t, err)
	require.Exactly(t, []string{"x", "y"}, values)
}

func TestServerParamValuesWithPath(t *testing.T) {
	server := &Server{
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
	server := &Server{
		URL: "https://{arg0}.{arg1}.example.com/",
	}
	for input, expected := range map[string]*serverMatch{
		"https://domain0.domain1.example.com/": newServerMatch("/", "domain0", "domain1"),
	} {
		t.Run(input, testServerParamValues(t, server, input, expected))
	}
}

func validServer() *Server {
	return &Server{
		URL: "http://my.cool.website",
	}
}

func invalidServer() *Server {
	return &Server{}
}

func TestServerValidation(t *testing.T) {
	tests := []struct {
		name          string
		input         *Server
		expectedError error
	}{
		{
			"when no URL is provided",
			invalidServer(),
			errors.New("value of url must be a non-empty JSON string"),
		},
		{
			"when a URL is provided",
			validServer(),
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := context.Background()
			validationErr := test.input.Validate(c)

			require.Equal(t, test.expectedError, validationErr, "expected errors (or lack of) to match")
		})
	}
}

func testServerParamValues(t *testing.T, server *Server, input string, expected *serverMatch) func(*testing.T) {
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
