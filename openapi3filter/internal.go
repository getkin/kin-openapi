package openapi3filter

import (
	"reflect"
	"strings"
)

func parseMediaType(contentType string) string {
	before, _, ok := strings.Cut(contentType, ";")
	if !ok {
		return contentType
	}
	return before
}

// lookupByContentType returns the entry registered for contentType.
//
// An exact key always wins. Otherwise the lookup falls back to a
// case-insensitive comparison, as RFC 9110 section 8.3.1 defines media type and
// subtype tokens as case-insensitive. Values missing a subtype are not valid
// media types and are matched exactly so they cannot resolve through the
// fallback.
func lookupByContentType[V any](registered map[string]V, contentType string) (V, bool) {
	if v, ok := registered[contentType]; ok {
		return v, true
	}
	var zero V
	if !strings.ContainsRune(contentType, '/') {
		return zero, false
	}
	// Compare against each entry. The lexicographically smallest match is taken
	// so that several registered case variants of one content type still
	// resolve deterministically.
	match := ""
	for candidate := range registered {
		if match != "" && candidate >= match {
			continue
		}
		if strings.EqualFold(contentType, candidate) {
			match = candidate
		}
	}
	if match == "" {
		return zero, false
	}
	return registered[match], true
}

func isNilValue(value any) bool {
	if value == nil {
		return true
	}
	switch reflect.TypeOf(value).Kind() {
	case reflect.Pointer, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(value).IsNil()
	}
	return false
}
