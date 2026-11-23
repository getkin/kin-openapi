package openapi2

import (
	"encoding/json"
	"maps"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// T is the root of an OpenAPI v2 document
type T struct {
	Extensions map[string]any `json:"-" yaml:"-"`
	// PathsExtensions holds the x- fields of the Paths Object, alongside its routes.
	PathsExtensions map[string]any `json:"-" yaml:"-"`

	Swagger             string                     `json:"swagger,case:strict" yaml:"swagger"` // required
	Info                openapi3.Info              `json:"info,case:strict" yaml:"info"`       // required
	ExternalDocs        *openapi3.ExternalDocs     `json:"externalDocs,omitempty,case:strict" yaml:"externalDocs,omitempty"`
	Schemes             []string                   `json:"schemes,omitempty,case:strict" yaml:"schemes,omitempty"`
	Consumes            []string                   `json:"consumes,omitempty,case:strict" yaml:"consumes,omitempty"`
	Produces            []string                   `json:"produces,omitempty,case:strict" yaml:"produces,omitempty"`
	Host                string                     `json:"host,omitempty,case:strict" yaml:"host,omitempty"`
	BasePath            string                     `json:"basePath,omitempty,case:strict" yaml:"basePath,omitempty"`
	Paths               map[string]*PathItem       `json:"paths,omitempty,case:strict" yaml:"paths,omitempty"`
	Definitions         map[string]*SchemaRef      `json:"definitions,omitempty,case:strict" yaml:"definitions,omitempty"`
	Parameters          map[string]*Parameter      `json:"parameters,omitempty,case:strict" yaml:"parameters,omitempty"`
	Responses           map[string]*Response       `json:"responses,omitempty,case:strict" yaml:"responses,omitempty"`
	SecurityDefinitions map[string]*SecurityScheme `json:"securityDefinitions,omitempty,case:strict" yaml:"securityDefinitions,omitempty"`
	Security            SecurityRequirements       `json:"security,omitempty,case:strict" yaml:"security,omitempty"`
	Tags                openapi3.Tags              `json:"tags,omitempty,case:strict" yaml:"tags,omitempty"`
}

// MarshalJSON returns the JSON encoding of T.
func (doc T) MarshalJSON() ([]byte, error) {
	m := make(map[string]any, 15+len(doc.Extensions))
	maps.Copy(m, doc.Extensions)
	m["swagger"] = doc.Swagger
	m["info"] = doc.Info
	if x := doc.ExternalDocs; x != nil {
		m["externalDocs"] = x
	}
	if x := doc.Schemes; len(x) != 0 {
		m["schemes"] = x
	}
	if x := doc.Consumes; len(x) != 0 {
		m["consumes"] = x
	}
	if x := doc.Produces; len(x) != 0 {
		m["produces"] = x
	}
	if x := doc.Host; x != "" {
		m["host"] = x
	}
	if x := doc.BasePath; x != "" {
		m["basePath"] = x
	}
	if len(doc.PathsExtensions) != 0 {
		m["paths"] = mapWithExtensions[PathItem]{m: doc.Paths, Extensions: doc.PathsExtensions}
	} else if len(doc.Paths) != 0 {
		m["paths"] = doc.Paths
	}
	if x := doc.Definitions; len(x) != 0 {
		m["definitions"] = x
	}
	if x := doc.Parameters; len(x) != 0 {
		m["parameters"] = x
	}
	if x := doc.Responses; len(x) != 0 {
		m["responses"] = x
	}
	if x := doc.SecurityDefinitions; len(x) != 0 {
		m["securityDefinitions"] = x
	}
	if x := doc.Security; len(x) != 0 {
		m["security"] = x
	}
	if x := doc.Tags; len(x) != 0 {
		m["tags"] = x
	}
	return json.Marshal(m)
}

// UnmarshalJSON sets T to a copy of data.
func (doc *T) UnmarshalJSON(data []byte) error {
	type TBis T
	var x TBis
	// The extension pass already reads every Paths key. Ordinary documents
	// can keep decoding their routes straight into the public map.
	extensionsErr := unmarshalJSON(data, &x.Extensions)
	var fieldsErr error
	if pathsHaveExtensions(x.Extensions) {
		wire := struct {
			*TBis
			Paths *pathsObject `json:"paths,case:strict" yaml:"paths"`
		}{TBis: &x}
		fieldsErr = unmarshalJSON(data, &wire)
		if wire.Paths != nil {
			x.Paths = wire.Paths.m
			x.PathsExtensions = wire.Paths.Extensions
		}
	} else {
		fieldsErr = unmarshalJSON(data, &x)
	}
	if fieldsErr != nil {
		return unmarshalError(fieldsErr)
	}
	if extensionsErr != nil {
		return unmarshalError(extensionsErr)
	}
	delete(x.Extensions, "swagger")
	delete(x.Extensions, "info")
	delete(x.Extensions, "externalDocs")
	delete(x.Extensions, "schemes")
	delete(x.Extensions, "consumes")
	delete(x.Extensions, "produces")
	delete(x.Extensions, "host")
	delete(x.Extensions, "basePath")
	delete(x.Extensions, "paths")
	delete(x.Extensions, "definitions")
	delete(x.Extensions, "parameters")
	delete(x.Extensions, "responses")
	delete(x.Extensions, "securityDefinitions")
	delete(x.Extensions, "security")
	delete(x.Extensions, "tags")
	if len(x.Extensions) == 0 {
		x.Extensions = nil
	}
	*doc = T(x)
	return nil
}

func pathsHaveExtensions(fields map[string]any) bool {
	for name, value := range fields {
		// encoding/json v1 also matches the paths field case-insensitively.
		if strings.EqualFold(name, "paths") {
			paths, _ := value.(map[string]any)
			for key := range paths {
				if strings.HasPrefix(key, "x-") {
					return true
				}
			}
		}
	}
	return false
}

func (doc *T) AddOperation(path string, method string, operation *Operation) {
	if doc.Paths == nil {
		doc.Paths = make(map[string]*PathItem)
	}
	pathItem := doc.Paths[path]
	if pathItem == nil {
		pathItem = &PathItem{}
		doc.Paths[path] = pathItem
	}
	pathItem.SetOperation(method, operation)
}
