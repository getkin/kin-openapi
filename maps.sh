#!/bin/bash -eux
set -o pipefail

maplike=./openapi3/maplike.go
maplike_test=./openapi3/maplike_test.go

types=()
types+=('*Responses')
types+=('*Callback')
types+=('*Paths')

value_types=()
value_types+=('*ResponseRef')
value_types+=('*PathItem')
value_types+=('*PathItem')

deref_vs=()
deref_vs+=('*Response = v.Value')
deref_vs+=('*PathItem = v')
deref_vs+=('*PathItem = v')

names=()
names+=('responses')
names+=('callback')
names+=('paths')

[[ "${#types[@]}" = "${#value_types[@]}" ]]
[[ "${#types[@]}" = "${#deref_vs[@]}" ]]
[[ "${#types[@]}" = "${#names[@]}" ]]
[[ "${#types[@]}" = "$(git grep -InF ' m map[string]*' -- openapi3/loader.go | wc -l)" ]]


maplike_header() {
	cat <<EOF >"$maplike"
package openapi3

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/go-openapi/jsonpointer"
)

EOF
}


test_header() {
	cat <<EOF >"$maplike_test"
package openapi3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMaplikeMethods(t *testing.T) {
	t.Parallel()

EOF
}


test_footer() {
	echo "}" >>"$maplike_test"
}


maplike_NewWithCapa() {
	cat <<EOF >>"$maplike"
// New${type#'*'}WithCapacity builds a ${name} object of the given capacity.
func New${type#'*'}WithCapacity(cap int) ${type} {
	if cap == 0 {
		return &${type#'*'}{m: make(map[string]${value_type})}
	}
	return &${type#'*'}{m: make(map[string]${value_type}, cap)}
}

EOF
}


maplike_ValueSetLen() {
	cat <<EOF >>"$maplike"
// Value returns the ${name} for key or nil
func (${name} ${type}) Value(key string) ${value_type} {
	if ${name}.Len() == 0 {
		return nil
	}
	return ${name}.m[key]
}

// Set adds or replaces key 'key' of '${name}' with 'value'.
// Note: '${name}' MUST be non-nil
func (${name} ${type}) Set(key string, value ${value_type}) {
	if ${name}.m == nil {
		${name}.m = make(map[string]${value_type})
	}
	${name}.m[key] = value
}

// Len returns the amount of keys in ${name} excluding ${name}.Extensions.
func (${name} ${type}) Len() int {
	if ${name} == nil || ${name}.m == nil {
		return 0
	}
	return len(${name}.m)
}

// Map returns ${name} as a 'map'.
// Note: iteration on Go maps is not ordered.
func (${name} ${type}) Map() (m map[string]${value_type}) {
	if ${name} == nil || len(${name}.m) == 0 {
		return make(map[string]${value_type})
	}
	m = make(map[string]${value_type}, len(${name}.m))
	for k, v := range ${name}.m {
		m[k] = v
	}
	return
}

EOF
}


maplike_Pointable() {
	cat <<EOF >>"$maplike"
var _ jsonpointer.JSONPointable = (${type})(nil)

// JSONLookup implements https://github.com/go-openapi/jsonpointer#JSONPointable
func (${name} ${type#'*'}) JSONLookup(token string) (interface{}, error) {
	if v := ${name}.Value(token); v == nil {
		vv, _, err := jsonpointer.GetForToken(${name}.Extensions, token)
		return vv, err
	} else if ref := v.Ref; ref != "" {
		return &Ref{Ref: ref}, nil
	} else {
		var vv ${deref_v}
		return vv, nil
	}
}

EOF
}


maplike_UnMarsh() {
	cat <<EOF >>"$maplike"
// MarshalJSON returns the JSON encoding of ${type#'*'}.
func (${name} ${type}) MarshalJSON() ([]byte, error) {
	m := make(map[string]interface{}, ${name}.Len()+len(${name}.Extensions))
	for k, v := range ${name}.Extensions {
		m[k] = v
	}
	for k, v := range ${name}.Map() {
		m[k] = v
	}
	return json.Marshal(m)
}

// UnmarshalJSON sets ${type#'*'} to a copy of data.
func (${name} ${type}) UnmarshalJSON(data []byte) (err error) {
	var m map[string]interface{}
	if err = json.Unmarshal(data, &m); err != nil {
		return
	}

	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)

	x := ${type#'*'}{
		Extensions: make(map[string]interface{}),
		m:          make(map[string]${value_type}, len(m)),
	}

	for _, k := range ks {
		v := m[k]
		if strings.HasPrefix(k, "x-") {
			x.Extensions[k] = v
			continue
		}

		var data []byte
		if data, err = json.Marshal(v); err != nil {
			return
		}
		var vv ${value_type#'*'}
		if err = vv.UnmarshalJSON(data); err != nil {
			return
		}
		x.m[k] = &vv
	}
	*${name} = x
	return
}
EOF
}


test_body() {
	cat <<EOF >>"$maplike_test"
	t.Run("${type}", func(t *testing.T) {
		t.Parallel()
		t.Run("nil", func(t *testing.T) {
			x := (${type})(nil)
			require.Equal(t, 0, x.Len())
			require.Equal(t, map[string]${value_type}{}, x.Map())
			require.Equal(t, (${value_type})(nil), x.Value("key"))
			require.Panics(t, func() { x.Set("key", &${value_type#'*'}{}) })
		})
		t.Run("nonnil", func(t *testing.T) {
			x := &${type#'*'}{}
			require.Equal(t, 0, x.Len())
			require.Equal(t, map[string]${value_type}{}, x.Map())
			require.Equal(t, (${value_type})(nil), x.Value("key"))
			x.Set("key", &${value_type#'*'}{})
			require.Equal(t, 1, x.Len())
			require.Equal(t, map[string]${value_type}{"key": {}}, x.Map())
			require.Equal(t, &${value_type#'*'}{}, x.Value("key"))
		})
	})

EOF
}



maplike_header
test_header

for i in "${!types[@]}"; do
	type=${types[$i]}
	value_type=${value_types[$i]}
	deref_v=${deref_vs[$i]}
	name=${names[$i]}

	type="$type" name="$name" value_type="$value_type" maplike_NewWithCapa
	type="$type" name="$name" value_type="$value_type" maplike_ValueSetLen
	type="$type" name="$name"    deref_v="$deref_v"    maplike_Pointable
	type="$type" name="$name" value_type="$value_type" maplike_UnMarsh
	[[ $((i+1)) != "${#types[@]}" ]] && echo >>"$maplike"

	type="$type" value_type="$value_type" test_body


done

test_footer
