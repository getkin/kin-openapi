package openapi3

import (
	"context"
	"github.com/jban332/kin-openapi/jsoninfo"
)

type Swagger struct {
	jsoninfo.ExtensionProps
	OpenAPI      string               `json:"openapi"` // Required
	Info         Info                 `json:"info"`    // Required
	Servers      Servers              `json:"servers,omitempty"`
	Paths        Paths                `json:"paths,omitempty"`
	Components   Components           `json:"components,omitempty"`
	Security     SecurityRequirements `json:"security,omitempty"`
	ExternalDocs *ExternalDocs        `json:"externalDocs,omitempty"`

	// To prevent infinite recursion
	visitedSchemas map[*Schema]struct{}
}

func (value *Swagger) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalStructFields(value)
}

func (value *Swagger) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalStructFields(data, value)
}

func (value *Swagger) ResolveRefs() error {
	resolver := &resolver{
		Swagger: value,
	}
	return resolver.ResolveRefs()
}

func (swagger *Swagger) AddOperation(path string, method string, operation *Operation) {
	paths := swagger.Paths
	if paths == nil {
		paths = make(Paths)
		swagger.Paths = paths
	}
	pathItem := paths[path]
	if pathItem == nil {
		pathItem = &PathItem{}
		paths[path] = pathItem
	}
	pathItem.SetOperation(method, operation)
}

func (swagger *Swagger) AddServer(server *Server) {
	swagger.Servers = append(swagger.Servers, server)
}

func (swagger *Swagger) Validate(c context.Context) error {
	if err := swagger.Components.Validate(c); err != nil {
		return err
	}
	if v := swagger.Security; v != nil {
		if err := v.Validate(c); err != nil {
			return err
		}
	}
	if paths := swagger.Paths; paths != nil {
		if err := paths.Validate(c); err != nil {
			return err
		}
	}
	if v := swagger.Servers; v != nil {
		if err := v.Validate(c); err != nil {
			return err
		}
	}
	if v := swagger.Paths; v != nil {
		if err := v.Validate(c); err != nil {
			return err
		}
	}
	return nil
}
