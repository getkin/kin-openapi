package openapi3gen

import (
	"reflect"
	"sort"
	"sync"
)

var (
	typeInfos      = map[reflect.Type]*theTypeInfo{}
	typeInfosMutex sync.RWMutex
)

// theTypeInfo contains information about JSON serialization of a type
type theTypeInfo struct {
	Type   reflect.Type
	Fields []theFieldInfo
}

// getTypeInfo returns theTypeInfo for the given type.
func getTypeInfo(t reflect.Type) *theTypeInfo {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	typeInfosMutex.RLock()
	typeInfo, exists := typeInfos[t]
	typeInfosMutex.RUnlock()
	if exists {
		return typeInfo
	}
	if t.Kind() != reflect.Struct {
		typeInfo = &theTypeInfo{
			Type: t,
		}
	} else {
		// Allocate
		typeInfo = &theTypeInfo{
			Type:   t,
			Fields: make([]theFieldInfo, 0, 16),
		}

		// Add fields
		typeInfo.Fields = appendFields(nil, nil, t)

		// Sort fields
		sort.Sort(sortableFieldInfos(typeInfo.Fields))
	}

	// Publish
	typeInfosMutex.Lock()
	typeInfos[t] = typeInfo
	typeInfosMutex.Unlock()
	return typeInfo
}
