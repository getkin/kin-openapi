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

	type componentType struct {
		Name           string
		CollectionName string
	}

	if err := packageTemplate.Execute(file, struct {
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
}
