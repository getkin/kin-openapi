#!/bin/bash -eux
set -o pipefail

types=()
types+=("Callback")
types+=("Example")
types+=("Header")
types+=("Link")
types+=("Parameter")
types+=("RequestBody")
types+=("Response")
types+=("Schema")
types+=("SecurityScheme")

cat <<EOF
package openapi3

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/go-openapi/jsonpointer"
	"github.com/perimeterx/marshmallow"
)
EOF

for type in "${types[@]}"; do
	cat <<EOF

// ${type}Ref represents either a ${type} or a \$ref to a ${type}.
// When serializing and both fields are set, Ref is preferred over Value.
type ${type}Ref struct {
	Ref   string
	Value *${type}
	extra []string
}

var _ jsonpointer.JSONPointable = (*${type}Ref)(nil)

func (x *${type}Ref) isEmpty() bool { return x == nil || x.Ref == "" && x.Value == nil }

// MarshalYAML returns the YAML encoding of ${type}Ref.
func (x ${type}Ref) MarshalYAML() (interface{}, error) {
	if ref := x.Ref; ref != "" {
		return &Ref{Ref: ref}, nil
	}
	return x.Value, nil
}

// MarshalJSON returns the JSON encoding of ${type}Ref.
func (x ${type}Ref) MarshalJSON() ([]byte, error) {
	if ref := x.Ref; ref != "" {
		return json.Marshal(Ref{Ref: ref})
	}
EOF

	case $type in
		# Callback) echo '	return x.Value.MarshalJSON()' ;; TODO: when https://github.com/getkin/kin-openapi/issues/687
		Example) echo '	return x.Value.MarshalJSON()' ;;
		Header) echo '	return x.Value.MarshalJSON()' ;;
		Link) echo '	return x.Value.MarshalJSON()' ;;
		Parameter) echo '	return x.Value.MarshalJSON()' ;;
		RequestBody) echo '	return x.Value.MarshalJSON()' ;;
		Response) echo '	return x.Value.MarshalJSON()' ;;
		Schema) echo '	return x.Value.MarshalJSON()' ;;
		SecurityScheme) echo '	return x.Value.MarshalJSON()' ;;
		*) echo '	return json.Marshal(x.Value)'
	esac

	cat <<EOF
}

// UnmarshalJSON sets ${type}Ref to a copy of data.
func (x *${type}Ref) UnmarshalJSON(data []byte) error {
	var refOnly Ref
	if extra, err := marshmallow.Unmarshal(data, &refOnly, marshmallow.WithExcludeKnownFieldsFromMap(true)); err == nil && refOnly.Ref != "" {
		x.Ref = refOnly.Ref
		if len(extra) != 0 {
			x.extra = make([]string, 0, len(extra))
			for key := range extra {
				x.extra = append(x.extra, key)
			}
			sort.Strings(x.extra)
		}
		return nil
	}
	return json.Unmarshal(data, &x.Value)
}

// Validate returns an error if ${type}Ref does not comply with the OpenAPI spec.
func (x *${type}Ref) Validate(ctx context.Context, opts ...ValidationOption) error {
	ctx = WithValidationOptions(ctx, opts...)
	if extra := x.extra; len(extra) != 0 {
		extras := make([]string, 0, len(extra))
		allowed := getValidationOptions(ctx).extraSiblingFieldsAllowed
		for _, ex := range extra {
			if allowed != nil {
				if _, ok := allowed[ex]; ok {
					continue
				}
			}
			extras = append(extras, ex)
		}
		if len(extras) != 0 {
			return fmt.Errorf("extra sibling fields: %+v", extras)
		}
	}
	if v := x.Value; v != nil {
		return v.Validate(ctx)
	}
	return foundUnresolvedRef(x.Ref)
}

// JSONLookup implements https://pkg.go.dev/github.com/go-openapi/jsonpointer#JSONPointable
func (x *${type}Ref) JSONLookup(token string) (interface{}, error) {
	if token == "\$ref" {
		return x.Ref, nil
	}
	ptr, _, err := jsonpointer.GetForToken(x.Value, token)
	return ptr, err
}
EOF

done
