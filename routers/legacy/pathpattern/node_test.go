package pathpattern

import (
	"testing"
)

func TestPatterns(t *testing.T) {
	DefaultOptions.SupportRegExp = true
	rootNode := &Node{}
	add := func(path, value string) {
		rootNode.MustAdd(path, value, nil)
	}
	add("GET /abc", "GET METHOD")
	add("POST /abc", "POST METHOD")

	add("/abc", "SIMPLE")
	add("/abc/fixedString", "FIXED STRING")
	add("/abc/{param}", "FILE")
	add("/abc/{param*}", "DEEP FILE")
	add("/abc/{fileName|(.*)\\.jpeg}", "JPEG")
	add("/abc/{fileName|some_prefix_(.*)\\.jpeg}", "PREFIXED JPEG")
	add("/root/{path*}", "DIRECTORY")
	add("/impossible_route", "IMPOSSIBLE")

	add(PathFromHost("www.nike.com", true), "WWW-HOST")
	add(PathFromHost("{other}.nike.com", true), "OTHER-HOST")

	expect := func(uri string, expected string, expectedArgs ...string) {
		actually := "not found"
		node, actualArgs := rootNode.Match(uri)
		if node != nil {
			if s, ok := node.Value.(string); ok {
				actually = s
			}
		}
		if actually != expected {
			t.Fatalf("Wrong path!\nInput: %s\nExpected: %q\nActually: %q\nTree:\n%s\n\n", uri, expected, actually, rootNode.String())
			return
		}
		if !argsEqual(expectedArgs, actualArgs) {
			t.Fatalf("Wrong variable values!\nInput: %s\nExpected: %q\nActually: %q\nTree:\n%s\n\n", uri, expectedArgs, actualArgs, rootNode.String())
			return
		}
	}
	expect("", "not found")
	expect("/", "not found")

	expect("GET /abc", "GET METHOD")
	expect("GET /abc/", "GET METHOD")
	expect("POST /abc", "POST METHOD")

	expect("/url_without_handler", "not found")
	expect("/abc", "SIMPLE")
	expect("/abc/fixedString", "FIXED STRING")
	expect("/abc/09az", "FILE", "09az")
	expect("/abc/09az/1/2/3", "DEEP FILE", "09az/1/2/3")
	expect("/abc/09az/1/2/3/", "DEEP FILE", "09az/1/2/3")
	expect("/abc/someFile.jpeg", "JPEG", "someFile")
	expect("/abc/someFile.old.jpeg", "JPEG", "someFile.old")
	expect("/abc/some_prefix_someFile.jpeg", "PREFIXED JPEG", "someFile")

	expect("/root", "DIRECTORY", "")
	expect("/root/", "DIRECTORY", "")
	expect("/root/a/b/c", "DIRECTORY", "a/b/c")

	expect(PathFromHost("www.nike.com", true), "WWW-HOST")
	expect(PathFromHost("example.nike.com", true), "OTHER-HOST", "example")
	expect(PathFromHost("subdomain.example.nike.com", true), "not found")
}

func argsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, ai := range a {
		if ai != b[i] {
			return false
		}
	}
	return true
}
