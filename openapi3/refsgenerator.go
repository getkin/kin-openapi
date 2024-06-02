//go:build ignore
// +build ignore

// The program generates refs.go, invoke `go generate ./...` to run.
package main

import (
	_ "embed"
	"os"
	"text/template"
)

//go:embed refs.tmpl
var tmplData string

func main() {
	file, err := os.Create("refs.go")
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()

	packageTemplate := template.Must(template.New("openapi3-refs").Parse(tmplData))

	if err := packageTemplate.Execute(file, struct {
		Package string
		Types   []string
	}{
		Package: os.Getenv("GOPACKAGE"), // set by the go:generate directive
		Types: []string{
			"Callback",
			"Example",
			"Header",
			"Link",
			"Parameter",
			"RequestBody",
			"Response",
			"Schema",
			"SecurityScheme",
		},
	}); err != nil {
		panic(err)
	}
}
