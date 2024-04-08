package openapi3

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/go-openapi/jsonpointer"
)

// T is the root of an OpenAPI v3 document
// See https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.3.md#openapi-object
type T struct {
	Extensions map[string]interface{} `json:"-" yaml:"-"`

	OpenAPI      string               `json:"openapi" yaml:"openapi"` // Required
	Components   *Components          `json:"components,omitempty" yaml:"components,omitempty"`
	Info         *Info                `json:"info" yaml:"info"`   // Required
	Paths        *Paths               `json:"paths" yaml:"paths"` // Required
	Security     SecurityRequirements `json:"security,omitempty" yaml:"security,omitempty"`
	Servers      Servers              `json:"servers,omitempty" yaml:"servers,omitempty"`
	Tags         Tags                 `json:"tags,omitempty" yaml:"tags,omitempty"`
	ExternalDocs *ExternalDocs        `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`

	visited visitedComponent
	url     *url.URL
}

// MatchesSchemaInRootDocument returns if the given schema is identical
// to a schema defined in the root document's '#/components/schemas'.
// It returns a reference to the schema in the form
// '#/components/schemas/NameXXX'
//
// Of course given it a schema from the root document will always match.
//
// https://swagger.io/docs/specification/using-ref/#syntax
//
// Case 1: Directly via
//
//	../openapi.yaml#/components/schemas/Record
//
// Case 2: Or indirectly by using a $ref which matches a schema
// in the root document's '#/components/schemas' using the same
// $ref.
//
// In schemas/record.yaml
//
//	$ref: ./record.yaml
//
// In openapi.yaml
//
//	  components:
//	    schemas:
//		  Record:
//		    $ref: schemas/record.yaml
func (doc *T) MatchesSchemaInRootDocument(sch *SchemaRef) (string, bool) {
	// Case 1:
	// Something like: ../another-folder/document.json#/myElement
	if isRemoteReference(sch.Ref) && isSchemaReference(sch.Ref) {
		// Determine if it is *this* root doc.
		if sch.referencesRootDocument(doc) {
			_, name, _ := strings.Cut(sch.Ref, "#/components/schemas")

			return path.Join("#/components/schemas", name), true
		}
	}

	// If there are no schemas defined in the root document return early.
	if doc.Components == nil || doc.Components.Schemas == nil {
		return "", false
	}

	// Case 2:
	// Something like: ../openapi.yaml#/components/schemas/myElement
	for name, s := range doc.Components.Schemas {
		// Must be a reference to a YAML file.
		if !isWholeDocumentReference(s.Ref) {
			continue
		}

		// Is the schema a ref to the same resource.
		if !sch.refersToSameDocument(s) {
			continue
		}

		// Transform the remote ref to the equivalent schema in the root document.
		return path.Join("#/components/schemas", name), true
	}

	return "", false
}

// isElementReference takes a $ref value and checks if it references a specific element.
func isElementReference(ref string) bool {
	return ref != "" && !isWholeDocumentReference(ref)
}

// isSchemaReference takes a $ref value and checks if it references a schema element.
func isSchemaReference(ref string) bool {
	return isElementReference(ref) && strings.Contains(ref, "#/components/schemas")
}

// isWholeDocumentReference takes a $ref value and checks if it is whole document reference.
func isWholeDocumentReference(ref string) bool {
	return ref != "" && !strings.ContainsAny(ref, "#")
}

// isRemoteReference takes a $ref value and checks if it is remote reference.
func isRemoteReference(ref string) bool {
	return ref != "" && !strings.HasPrefix(ref, "#") && !isURLReference(ref)
}

// isURLReference takes a $ref value and checks if it is URL reference.
func isURLReference(ref string) bool {
	return strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") || strings.HasPrefix(ref, "//")
}

var _ jsonpointer.JSONPointable = (*T)(nil)

// JSONLookup implements https://pkg.go.dev/github.com/go-openapi/jsonpointer#JSONPointable
func (doc *T) JSONLookup(token string) (interface{}, error) {
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
func (doc *T) MarshalYAML() (interface{}, error) {
	m := make(map[string]interface{}, 4+len(doc.Extensions))
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

	return validateExtensions(ctx, doc.Extensions)
}
