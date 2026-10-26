//go:build ignore

// The program generates native_yaml_shadow.go, invoke `go generate ./...` to run.
package main

import (
	"bytes"
	_ "embed"
	"go/format"
	"os"
	"text/template"
)

//go:embed nativeyaml.tmpl
var tmplData string

type shadowType struct {
	Name string
	Recv string
}

func main() {
	// The types whose YAML form is a mapping of declared fields plus
	// extensions. The $ref wrappers are generated from refs.tmpl instead.
	// Hand-written in native_yaml_special.go: the maplike collections, the
	// union-typed values, Schema, which may be written as a boolean, Operation,
	// which has to tell an omitted responses
	// from an explicitly null one, and T, which has no parent to take its
	// Key from.
	types := []shadowType{
		{"Components", "components"},
		{"Contact", "contact"},
		{"Discriminator", "discriminator"},
		{"Encoding", "encoding"},
		{"Example", "example"},
		{"ExternalDocs", "e"},
		{"Info", "info"},
		{"License", "license"},
		{"Link", "link"},
		{"MediaType", "mediaType"},
		{"OAuthFlow", "flow"},
		{"OAuthFlows", "flows"},
		{"Parameter", "parameter"},
		{"PathItem", "pathItem"},
		{"RequestBody", "requestBody"},
		{"Response", "response"},
		{"SecurityScheme", "ss"},
		{"Server", "server"},
		{"ServerVariable", "serverVariable"},
		{"Tag", "t"},
		{"XML", "xml"},
	}

	tmpl := template.Must(template.New("nativeyaml").Parse(tmplData))
	buf := new(bytes.Buffer)
	if err := tmpl.Execute(buf, struct {
		Package string
		Types   []shadowType
	}{
		Package: os.Getenv("GOPACKAGE"), // set by the go:generate directive
		Types:   types,
	}); err != nil {
		panic(err)
	}
	src, err := format.Source(buf.Bytes())
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile("native_yaml_shadow.go", src, 0o644); err != nil {
		panic(err)
	}
}
