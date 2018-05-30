package openapi3

import (
	"context"
	"fmt"
	"github.com/ronniedada/kin-openapi/jsoninfo"
)

type PathItem struct {
	ExtensionProps
	Summary     string     `json:"summary,omitempty"`
	Description string     `json:"description,omitempty"`
	Delete      *Operation `json:"delete,omitempty"`
	Get         *Operation `json:"get,omitempty"`
	Head        *Operation `json:"head,omitempty"`
	Options     *Operation `json:"options,omitempty"`
	Patch       *Operation `json:"patch,omitempty"`
	Post        *Operation `json:"post,omitempty"`
	Put         *Operation `json:"put,omitempty"`
	Trace       *Operation `json:"trace,omitempty"`
	Servers     Servers    `json:"servers,omitempty"`
	Parameters  Parameters `json:"parameters,omitempty"`
}

func (value *PathItem) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalStrictStruct(value)
}

func (value *PathItem) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalStrictStruct(data, value)
}

func (pathItem *PathItem) Operations() map[string]*Operation {
	operations := make(map[string]*Operation, 4)
	if v := pathItem.Delete; v != nil {
		operations["DELETE"] = v
	}
	if v := pathItem.Get; v != nil {
		operations["GET"] = v
	}
	if v := pathItem.Head; v != nil {
		operations["HEAD"] = v
	}
	if v := pathItem.Options; v != nil {
		operations["OPTIONS"] = v
	}
	if v := pathItem.Patch; v != nil {
		operations["PATCH"] = v
	}
	if v := pathItem.Post; v != nil {
		operations["POST"] = v
	}
	if v := pathItem.Put; v != nil {
		operations["PUT"] = v
	}
	return operations
}

func (pathItem *PathItem) GetOperation(method string) *Operation {
	switch method {
	case "DELETE":
		return pathItem.Delete
	case "GET":
		return pathItem.Get
	case "HEAD":
		return pathItem.Head
	case "OPTIONS":
		return pathItem.Options
	case "PATCH":
		return pathItem.Patch
	case "POST":
		return pathItem.Post
	case "PUT":
		return pathItem.Put
	case "TRACE":
		return pathItem.Trace
	default:
		panic(fmt.Errorf("Unsupported HTTP method '%s'", method))
	}
}

func (pathItem *PathItem) SetOperation(method string, operation *Operation) {
	switch method {
	case "DELETE":
		pathItem.Delete = operation
	case "GET":
		pathItem.Get = operation
	case "HEAD":
		pathItem.Head = operation
	case "OPTIONS":
		pathItem.Options = operation
	case "PATCH":
		pathItem.Patch = operation
	case "POST":
		pathItem.Post = operation
	case "PUT":
		pathItem.Put = operation
	case "TRACE":
		pathItem.Trace = operation
	default:
		panic(fmt.Errorf("Unsupported HTTP method '%s'", method))
	}
}

func (pathItem *PathItem) Validate(c context.Context) error {
	for method, operation := range pathItem.Operations() {
		err := operation.ValidateOperation(c, pathItem, method)
		if err != nil {
			return err
		}
	}
	return nil
}
