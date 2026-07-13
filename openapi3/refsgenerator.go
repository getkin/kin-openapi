//go:build ignore

// The program generates refs.go, invoke `go generate ./...` to run.
package main

import (
	"bytes"
	_ "embed"
	"go/format"
	"os"
	"text/template"
)

//go:embed refs.tmpl
var tmplData string

//go:embed refs_test.tmpl
var tmplTestData string

func main() {
	generateTemplate("refs", tmplData)
	generateTemplate("refs_test", tmplTestData)
}

func generateTemplate(filename string, tmpl string) {
	packageTemplate := template.Must(template.New("openapi3-" + filename).Parse(tmpl))

	type componentType struct {
		Name           string
		CollectionName string
	}

	var output bytes.Buffer
	if err := packageTemplate.Execute(&output, struct {
		Package string
		Types   []componentType
	}{
		Package: os.Getenv("GOPACKAGE"), // set by the go:generate directive
		Types: []componentType{
			{Name: "Callback", CollectionName: "callbacks"},
			{Name: "Example", CollectionName: "examples"},
			{Name: "Header", CollectionName: "headers"},
			{Name: "Link", CollectionName: "links"},
			{Name: "Parameter", CollectionName: "parameters"},
			{Name: "RequestBody", CollectionName: "requestBodies"},
			{Name: "Response", CollectionName: "responses"},
			{Name: "Schema", CollectionName: "schemas"},
			{Name: "SecurityScheme", CollectionName: "securitySchemes"},
		},
	}); err != nil {
		panic(err)
	}

	formatted, err := format.Source(output.Bytes())
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(filename+".go", formatted, 0o644); err != nil {
		panic(err)
	}
}
