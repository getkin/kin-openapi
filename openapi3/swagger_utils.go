package openapi3

import (
	"reflect"
	"strings"
)

type RefOrValue interface {
	Resolved() bool
	ClearRef()
	GetRef() string
	IsRef() bool
}

func IsExternalRef(rr RefOrValue) bool {
	return strings.Index(rr.GetRef(), "#") >= 0
}

func clearResolvedExternalRef(rr RefOrValue) {
	if rr.IsRef() && IsExternalRef(rr) && rr.Resolved() {
		rr.ClearRef()
	}
}

// ClearResolvedExternalRefs Recursively iterate over the swagger structure, resetting <Type>Ref structs where
// the reference is remote and was resolved
func ClearResolvedExternalRefs(swagger *Swagger) {
	visited := map[reflect.Value]struct{}{}
	resetExternalRef(reflect.ValueOf(swagger), visited)
}

func resetExternalRef(c reflect.Value, visited map[reflect.Value]struct{}) {
	if _, ok := visited[c]; ok {
		return
	}
	visited[c] = struct{}{}
	switch c.Kind() {
	// If it is a struct, check if it's the desired type first before drilling into fields
	// Further if this is a <Type>Ref struct, reset the reference if it's remote and resolved
	case reflect.Struct:
		if c.CanAddr() {
			rov, ok := c.Addr().Interface().(RefOrValue)
			if ok {
				clearResolvedExternalRef(rov)
			}
		}
		for i := 0; i < c.NumField(); i++ {
			resetExternalRef(c.Field(i), visited)
		}

	// If it is a pointer or interface we need to unwrap and call once again
	case reflect.Interface, reflect.Ptr:
		c2 := c.Elem()
		if c2.IsValid() {
			resetExternalRef(c2, visited)
		}

	// If it is a slice we iterate over each each element
	case reflect.Slice:
		for i := 0; i < c.Len(); i++ {
			resetExternalRef(c.Index(i), visited)
		}

	// If it is a map we iterate over each of the key,value pairs
	case reflect.Map:
		mi := c.MapRange()
		for mi.Next() {
			resetExternalRef(mi.Value(), visited)
		}

	// And everything else will simply be ignored
	default:
	}
}
