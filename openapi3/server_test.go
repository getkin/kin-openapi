package openapi3_test

import (
	"github.com/ronniedada/kin-openapi/openapi3"
	"github.com/ronniedada/kin-test/jsontest"
	"testing"
)

type ServerMatch struct {
	Remaining string
	Args      []string
}

func NewServerMatch(remaining string, args ...string) *ServerMatch {
	return &ServerMatch{
		Remaining: remaining,
		Args:      args,
	}
}

func TestServerParamNames(t *testing.T) {
	server := &openapi3.Server{
		URL: "http://{x}.{y}.example.com",
	}
	values, err := server.ParameterNames()
	jsontest.ExpectWithErr(t, values, err).Value([]interface{}{
		"x",
		"y",
	})
}

func TestServerParamValues(t *testing.T) {
	var server openapi3.Server
	expect := func(input string, expected *ServerMatch) {
		args, remaining, ok := server.MatchRawURL(input)
		if expected == nil {
			if ok {
				t.Fatalf("Should not have matched!\nPattern: %s\nInput: %s",
					server.URL,
					input)

			}
			return
		}
		actual := &ServerMatch{
			Remaining: remaining,
			Args:      args,
		}
		if !ok {
			t.Fatalf("Should have matched!\nPattern: %s\nInput: %s",
				server.URL,
				input)
		}
		jsontest.Expect(t, actual).Value(expected)
	}
	server.URL = "http://{arg0}.{arg1}.example.com/a/b"
	expect("http://x.example.com/a/b", nil)
	expect("http://x.y.example.com/", nil)
	expect("http://x.y.example.com/a/", nil)
	expect("http://x.y.example.com/a/b", NewServerMatch("/", "x", "y"))
	expect("http://x.y.example.com/a/b/", NewServerMatch("/", "x", "y"))
	expect("http://x.y.example.com/a/b/c", NewServerMatch("/c", "x", "y"))
	expect("http://domain0.domain1.example.com/a/b/c", NewServerMatch("/c", "domain0", "domain1"))
	server.URL = "https://{arg0}.{arg1}.example.com/"
	expect("https://domain0.domain1.example.com/", NewServerMatch("/", "domain0", "domain1"))
}
