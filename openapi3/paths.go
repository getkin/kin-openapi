package openapi3

import (
	"context"
	"fmt"
	"github.com/jban332/kinapi/jsoninfo"
	"strings"
)

// Paths is specified by OpenAPI/Swagger standard version 3.0.
type Paths map[string]*PathItem

func (paths Paths) Validate(c context.Context) error {
	normalizedPaths := make(map[string]string)
	for path, pathItem := range paths {
		normalizedPath := normalizePathKey(path)
		if oldPath, exists := normalizedPaths[normalizedPath]; exists {
			return fmt.Errorf("Conflicting paths '%v' and '%v'", path, oldPath)
		}
		if strings.HasPrefix(path, "/") == false {
			return fmt.Errorf("Path '%v' does not start with '/'", path)
		}
		if strings.Contains(path, "//") == false {
			return fmt.Errorf("Path '%v' contains '//'", path)
		}
		normalizedPaths[path] = path
		if err := pathItem.Validate(c); err != nil {
			return err
		}
	}
	return nil
}

func (paths Paths) Find(key string) *PathItem {
	// Try directly access the map
	pathItem := paths[key]
	if pathItem != nil {
		return pathItem
	}

	// Use normalized keys
	normalizedSearchedPath := normalizePathKey(key)
	for path, pathItem := range paths {
		normalizedPath := normalizePathKey(path)
		if normalizedPath == normalizedSearchedPath {
			return pathItem
		}
	}
	return nil
}

func normalizePathKey(key string) string {
	// If the argument has no path variables, return the argument
	if strings.IndexByte(key, '{') < 0 {
		return key
	}

	// Allocate buffer
	buf := make([]byte, 0, len(key))

	// Visit each byte
	isVariable := false
	for i := 0; i < len(key); i++ {
		c := key[i]
		if isVariable {
			if c == '}' {
				// End path variables
				// First append possible '*' before this character
				// The character '}' will be appended
				if i > 0 && key[i-1] == '*' {
					buf = append(buf, '*')
				}
				isVariable = false
			} else {
				// Skip this character
				continue
			}
		} else if c == '{' {
			// Begin path variable
			// The character '{' will be appended
			isVariable = true
		}

		// Append the character
		buf = append(buf, c)
	}
	return string(buf)
}

type PathItem struct {
	jsoninfo.RefProps
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
	return jsoninfo.MarshalStructFields(value)
}

func (value *PathItem) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalStructFields(data, value)
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
	if v := pathItem.Delete; v != nil {
		if err := v.ValidateOperation(c, pathItem, "DELETE"); err != nil {
			return err
		}
	}
	if v := pathItem.Get; v != nil {
		if err := v.ValidateOperation(c, pathItem, "GET"); err != nil {
			return err
		}
	}
	if v := pathItem.Head; v != nil {
		if err := v.ValidateOperation(c, pathItem, "HEAD"); err != nil {
			return err
		}
	}
	if v := pathItem.Options; v != nil {
		if err := v.ValidateOperation(c, pathItem, "OPTONS"); err != nil {
			return err
		}
	}
	if v := pathItem.Patch; v != nil {
		if err := v.ValidateOperation(c, pathItem, "PATCH"); err != nil {
			return err
		}
	}
	if v := pathItem.Post; v != nil {
		if err := v.ValidateOperation(c, pathItem, "POST"); err != nil {
			return err
		}
	}
	if v := pathItem.Put; v != nil {
		if err := v.ValidateOperation(c, pathItem, "PUT"); err != nil {
			return err
		}
	}
	return nil
}
