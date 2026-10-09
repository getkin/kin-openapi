package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServerParamNames(t *testing.T) {
	t.Parallel()

	server := &Server{
		URL: "http://{x}.{y}.example.com",
	}
	values, err := server.ParameterNames()
	require.NoError(t, err)
	require.Exactly(t, []string{"x", "y"}, values)
}

func TestServerParamValuesWithPath(t *testing.T) {
	t.Parallel()

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
		t.Run(input, testServerParamValues(server, input, expected))
	}
}

func TestServerParamValuesNoPath(t *testing.T) {
	t.Parallel()

	server := &Server{
		URL: "https://{arg0}.{arg1}.example.com/",
	}
	for input, expected := range map[string]*serverMatch{
		"https://domain0.domain1.example.com/": newServerMatch("/", "domain0", "domain1"),
	} {
		t.Run(input, testServerParamValues(server, input, expected))
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
	t.Parallel()

	tests := []struct {
		name             string
		input            *Server
		expectedErrorMsg string // empty = expect no error
	}{
		{
			"when no URL is provided",
			invalidServer(),
			"value of url must be a non-empty string",
		},
		{
			"when a URL is provided",
			validServer(),
			"",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := t.Context()
			validationErr := test.input.Validate(c)

			if test.expectedErrorMsg == "" {
				require.NoError(t, validationErr)
			} else {
				require.EqualError(t, validationErr, test.expectedErrorMsg)
			}
		})
	}
}

// A variable that Server.Variables declares but Server.URL never references
// is unused, not undeclared: the two defects call for opposite edits, so
// mislabelling one sends the reader to fix the wrong half of the spec.
func TestServerValidationUnusedVariables(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		input            *Server
		expectedErrorMsg string // empty = expect no error
	}{
		{
			"when a declared variable is missing from the URL",
			&Server{
				URL: "https://api.example.com/v1",
				Variables: ServerVariables{
					"region": &ServerVariable{Default: "us-east"},
				},
			},
			"server has unused variable region",
		},
		{
			"when the URL references no variables at all",
			&Server{
				URL: "https://{x}.example.com",
				Variables: ServerVariables{
					"x": &ServerVariable{Default: "www"},
					"y": &ServerVariable{Default: "com"},
				},
			},
			"server has unused variable y",
		},
		{
			"when only surplus variables are declared",
			&Server{
				URL: "https://api.example.com/{x}",
				Variables: ServerVariables{
					"x": &ServerVariable{Default: "www"},
					"y": &ServerVariable{Default: "com"},
					"z": &ServerVariable{Default: "net"},
				},
			},
			"server has unused variable y",
		},
		{
			"an undeclared reference outranks an unused variable",
			&Server{
				URL: "https://{y}.example.com",
				Variables: ServerVariables{
					"x": &ServerVariable{Default: "www"},
				},
			},
			"server has undeclared variables",
		},
		{
			"a repeated reference is declared once",
			&Server{
				URL: "https://{x}.example.com/{x}",
				Variables: ServerVariables{
					"x": &ServerVariable{Default: "www"},
				},
			},
			"",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.input.Validate(t.Context())

			if test.expectedErrorMsg == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, test.expectedErrorMsg)
			}
		})
	}
}

func testServerParamValues(server *Server, input string, expected *serverMatch) func(*testing.T) {
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

func TestServersBasePath(t *testing.T) {
	t.Parallel()

	for _, testcase := range []struct {
		title    string
		servers  Servers
		expected string
	}{
		{
			title:    "empty servers",
			servers:  nil,
			expected: "/",
		},
		{
			title:    "URL set, missing trailing slash",
			servers:  Servers{&Server{URL: "https://example.com"}},
			expected: "/",
		},
		{
			title:    "URL set, with trailing slash",
			servers:  Servers{&Server{URL: "https://example.com/"}},
			expected: "/",
		},
		{
			title:    "URL set",
			servers:  Servers{&Server{URL: "https://example.com/b/l/a"}},
			expected: "/b/l/a",
		},
		{
			title: "URL set with variables",
			servers: Servers{&Server{
				URL: "{scheme}://example.com/b/l/a",
				Variables: map[string]*ServerVariable{
					"scheme": {
						Enum:    []string{"http", "https"},
						Default: "https",
					},
				},
			}},
			expected: "/b/l/a",
		},
		{
			title: "URL set with variables in path",
			servers: Servers{&Server{
				URL: "http://example.com/b/{var1}/a",
				Variables: map[string]*ServerVariable{
					"var1": {
						Default: "lllll",
					},
				},
			}},
			expected: "/b/lllll/a",
		},
		{
			title: "URLs set with variables in path",
			servers: Servers{
				&Server{
					URL: "http://example.com/b/{var2}/a",
					Variables: map[string]*ServerVariable{
						"var2": {
							Default: "LLLLL",
						},
					},
				},
				&Server{
					URL: "https://example.com/b/{var1}/a",
					Variables: map[string]*ServerVariable{
						"var1": {
							Default: "lllll",
						},
					},
				},
			},
			expected: "/b/LLLLL/a",
		},
	} {
		t.Run(testcase.title, func(t *testing.T) {
			err := testcase.servers.Validate(t.Context())
			require.NoError(t, err)

			got, err := testcase.servers.BasePath()
			require.NoError(t, err)
			require.Exactly(t, testcase.expected, got)
		})
	}
}
