package openapi3

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/go-openapi/jsonpointer"
)

// T is the root of an OpenAPI v3 document
// See https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.3.md#openapi-object
// and https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.1.0.md#openapi-object
type T struct {
	Extensions map[string]any `json:"-" yaml:"-"`

	OpenAPI      string               `json:"openapi" yaml:"openapi"` // Required
	Components   *Components          `json:"components,omitempty" yaml:"components,omitempty"`
	Info         *Info                `json:"info" yaml:"info"`   // Required
	Paths        *Paths               `json:"paths" yaml:"paths"` // Required
	Security     SecurityRequirements `json:"security,omitempty" yaml:"security,omitempty"`
	Servers      Servers              `json:"servers,omitempty" yaml:"servers,omitempty"`
	Tags         Tags                 `json:"tags,omitempty" yaml:"tags,omitempty"`
	ExternalDocs *ExternalDocs        `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`

	// OpenAPI 3.1.x specific fields
	// Webhooks are a new feature in OpenAPI 3.1 that allow APIs to define callback operations
	Webhooks map[string]*PathItem `json:"webhooks,omitempty" yaml:"webhooks,omitempty"`

	// JSONSchemaDialect allows specifying the default JSON Schema dialect for Schema Objects
	// See https://spec.openapis.org/oas/v3.1.0#schema-object
	JSONSchemaDialect string `json:"jsonSchemaDialect,omitempty" yaml:"jsonSchemaDialect,omitempty"`

	visited visitedComponent
	url     *url.URL
}

var _ jsonpointer.JSONPointable = (*T)(nil)

// IsOpenAPI3_0 returns true if the document is OpenAPI 3.0.x
func (doc *T) IsOpenAPI3_0() bool {
	return doc.Version() == "3.0"
}

// IsOpenAPI3_1 returns true if the document is OpenAPI 3.1.x
func (doc *T) IsOpenAPI3_1() bool {
	return doc.Version() == "3.1"
}

// Version returns the major.minor version of the OpenAPI document
func (doc *T) Version() string {
	if doc == nil || doc.OpenAPI == "" {
		return ""
	}
	// Extract major.minor (e.g., "3.0" from "3.0.3")
	if len(doc.OpenAPI) >= 3 {
		return doc.OpenAPI[0:3]
	}
	return doc.OpenAPI
}

// JSONLookup implements https://pkg.go.dev/github.com/go-openapi/jsonpointer#JSONPointable
func (doc *T) JSONLookup(token string) (any, error) {
	switch token {
	case "openapi":
		return doc.OpenAPI, nil
	case "components":
		return doc.Components, nil
	case "info":
		return doc.Info, nil
	case "paths":
		return doc.Paths, nil
	case "security":
		return doc.Security, nil
	case "servers":
		return doc.Servers, nil
	case "tags":
		return doc.Tags, nil
	case "externalDocs":
		return doc.ExternalDocs, nil
	case "webhooks":
		return doc.Webhooks, nil
	case "jsonSchemaDialect":
		return doc.JSONSchemaDialect, nil
	}

	v, _, err := jsonpointer.GetForToken(doc.Extensions, token)
	return v, err
}

// MarshalJSON returns the JSON encoding of T.
func (doc *T) MarshalJSON() ([]byte, error) {
	x, err := doc.MarshalYAML()
	if err != nil {
		return nil, err
	}
	return json.Marshal(x)
}

// MarshalYAML returns the YAML encoding of T.
func (doc *T) MarshalYAML() (any, error) {
	if doc == nil {
		return nil, nil
	}
	m := make(map[string]any, 4+len(doc.Extensions))
	for k, v := range doc.Extensions {
		m[k] = v
	}
	m["openapi"] = doc.OpenAPI
	if x := doc.Components; x != nil {
		m["components"] = x
	}
	m["info"] = doc.Info
	m["paths"] = doc.Paths
	if x := doc.Security; len(x) != 0 {
		m["security"] = x
	}
	if x := doc.Servers; len(x) != 0 {
		m["servers"] = x
	}
	if x := doc.Tags; len(x) != 0 {
		m["tags"] = x
	}
	if x := doc.ExternalDocs; x != nil {
		m["externalDocs"] = x
	}
	// OpenAPI 3.1 fields
	if x := doc.Webhooks; len(x) != 0 {
		m["webhooks"] = x
	}
	if x := doc.JSONSchemaDialect; x != "" {
		m["jsonSchemaDialect"] = x
	}
	return m, nil
}

// UnmarshalJSON sets T to a copy of data.
func (doc *T) UnmarshalJSON(data []byte) error {
	type TBis T
	var x TBis
	if err := json.Unmarshal(data, &x); err != nil {
		return unmarshalError(err)
	}
	_ = json.Unmarshal(data, &x.Extensions)
	delete(x.Extensions, "openapi")
	delete(x.Extensions, "components")
	delete(x.Extensions, "info")
	delete(x.Extensions, "paths")
	delete(x.Extensions, "security")
	delete(x.Extensions, "servers")
	delete(x.Extensions, "tags")
	delete(x.Extensions, "externalDocs")
	// OpenAPI 3.1 fields
	delete(x.Extensions, "webhooks")
	delete(x.Extensions, "jsonSchemaDialect")
	if len(x.Extensions) == 0 {
		x.Extensions = nil
	}
	*doc = T(x)
	return nil
}

func (doc *T) AddOperation(path string, method string, operation *Operation) {
	if doc.Paths == nil {
		doc.Paths = NewPaths()
	}
	pathItem := doc.Paths.Value(path)
	if pathItem == nil {
		pathItem = &PathItem{}
		doc.Paths.Set(path, pathItem)
	}
	pathItem.SetOperation(method, operation)
}

func (doc *T) AddServer(server *Server) {
	doc.Servers = append(doc.Servers, server)
}

func (doc *T) AddServers(servers ...*Server) {
	doc.Servers = append(doc.Servers, servers...)
}

// Validate returns an error if T does not comply with the OpenAPI spec.
// Validations Options can be provided to modify the validation behavior.
func (doc *T) Validate(ctx context.Context, opts ...ValidationOption) error {
	ctx = WithValidationOptions(ctx, opts...)

	if doc.OpenAPI == "" {
		return errors.New("value of openapi must be a non-empty string")
	}

	var wrap func(error) error

	wrap = func(e error) error { return fmt.Errorf("invalid components: %w", e) }
	if v := doc.Components; v != nil {
		if err := v.Validate(ctx); err != nil {
			return wrap(err)
		}
	}

	wrap = func(e error) error { return fmt.Errorf("invalid info: %w", e) }
	if v := doc.Info; v != nil {
		if err := v.Validate(ctx); err != nil {
			return wrap(err)
		}
	} else {
		return wrap(errors.New("must be an object"))
	}

	wrap = func(e error) error { return fmt.Errorf("invalid paths: %w", e) }
	if v := doc.Paths; v != nil {
		if err := v.Validate(ctx); err != nil {
			return wrap(err)
		}
	} else {
		return wrap(errors.New("must be an object"))
	}

	wrap = func(e error) error { return fmt.Errorf("invalid security: %w", e) }
	if v := doc.Security; v != nil {
		if err := v.Validate(ctx); err != nil {
			return wrap(err)
		}
	}

	wrap = func(e error) error { return fmt.Errorf("invalid servers: %w", e) }
	if v := doc.Servers; v != nil {
		if err := v.Validate(ctx); err != nil {
			return wrap(err)
		}
	}

	wrap = func(e error) error { return fmt.Errorf("invalid tags: %w", e) }
	if v := doc.Tags; v != nil {
		if err := v.Validate(ctx); err != nil {
			return wrap(err)
		}
	}

	wrap = func(e error) error { return fmt.Errorf("invalid external docs: %w", e) }
	if v := doc.ExternalDocs; v != nil {
		if err := v.Validate(ctx); err != nil {
			return wrap(err)
		}
	}

	// OpenAPI 3.1 webhooks validation
	if doc.Webhooks != nil {
		wrap = func(e error) error { return fmt.Errorf("invalid webhooks: %w", e) }
		for name, pathItem := range doc.Webhooks {
			if pathItem == nil {
				return wrap(fmt.Errorf("webhook %q is nil", name))
			}
			if err := pathItem.Validate(ctx); err != nil {
				return wrap(fmt.Errorf("webhook %q: %w", name, err))
			}
		}
	}

	return validateExtensions(ctx, doc.Extensions)
}
