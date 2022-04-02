package openapi3

import (
	"fmt"
	"sort"
	"strings"
)

func validateExtensions(extensions map[string]interface{}) error { // FIXME: newtype + Validate(...)
	var unknowns []string
	for k := range extensions {
		if !strings.HasPrefix(k, "x-") {
			unknowns = append(unknowns, k)
		}
	}
	if len(unknowns) != 0 {
		sort.Strings(unknowns)
		return fmt.Errorf("extra sibling fields: %+v", unknowns)
	}
	return nil
}
