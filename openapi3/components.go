package openapi3

import (
	"context"
	"fmt"
	"regexp"
	"sort"

	"github.com/getkin/kin-openapi/jsoninfo"
)

// Components is specified by OpenAPI/Swagger standard version 3.
// See https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.3.md#componentsObject
type Components struct {
	ExtensionProps

	Schemas         Schemas         `json:"schemas,omitempty" yaml:"schemas,omitempty"`
	Parameters      ParametersMap   `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Headers         Headers         `json:"headers,omitempty" yaml:"headers,omitempty"`
	RequestBodies   RequestBodies   `json:"requestBodies,omitempty" yaml:"requestBodies,omitempty"`
	Responses       Responses       `json:"responses,omitempty" yaml:"responses,omitempty"`
	SecuritySchemes SecuritySchemes `json:"securitySchemes,omitempty" yaml:"securitySchemes,omitempty"`
	Examples        Examples        `json:"examples,omitempty" yaml:"examples,omitempty"`
	Links           Links           `json:"links,omitempty" yaml:"links,omitempty"`
	Callbacks       Callbacks       `json:"callbacks,omitempty" yaml:"callbacks,omitempty"`
}

func NewComponents() Components {
	return Components{}
}

// MarshalJSON returns the JSON encoding of Components.
func (components *Components) MarshalJSON() ([]byte, error) {
	return jsoninfo.MarshalStrictStruct(components)
}

// UnmarshalJSON sets Components to a copy of data.
func (components *Components) UnmarshalJSON(data []byte) error {
	return jsoninfo.UnmarshalStrictStruct(data, components)
}

// Validate returns an error if Components does not comply with the OpenAPI spec.
func (components *Components) Validate(ctx context.Context) (err error) {
	schemas := make([]string, 0, len(components.Schemas))
	for name := range components.Schemas {
		schemas = append(schemas, name)
	}
	sort.Strings(schemas)
	for _, k := range schemas {
		v := components.Schemas[k]
		if err = ValidateIdentifier(k); err != nil {
			return fmt.Errorf("schema %q: %w", k, err)
		}
		if err = v.Validate(ctx); err != nil {
			return fmt.Errorf("schema %q: %w", k, err)
		}
	}

	parameters := make([]string, 0, len(components.Parameters))
	for name := range components.Parameters {
		parameters = append(parameters, name)
	}
	sort.Strings(parameters)
	for _, k := range parameters {
		v := components.Parameters[k]
		if err = ValidateIdentifier(k); err != nil {
			return fmt.Errorf("parameter %q: %w", k, err)
		}
		if err = v.Validate(ctx); err != nil {
			return fmt.Errorf("parameter %q: %w", k, err)
		}
	}

	requestBodies := make([]string, 0, len(components.RequestBodies))
	for name := range components.RequestBodies {
		requestBodies = append(requestBodies, name)
	}
	sort.Strings(requestBodies)
	for _, k := range requestBodies {
		v := components.RequestBodies[k]
		if err = ValidateIdentifier(k); err != nil {
			return fmt.Errorf("request body %q: %w", k, err)
		}
		if err = v.Validate(ctx); err != nil {
			return fmt.Errorf("request body %q: %w", k, err)
		}
	}

	responses := make([]string, 0, len(components.Responses))
	for name := range components.Responses {
		responses = append(responses, name)
	}
	sort.Strings(responses)
	for _, k := range responses {
		v := components.Responses[k]
		if err = ValidateIdentifier(k); err != nil {
			return fmt.Errorf("response %q: %w", k, err)
		}
		if err = v.Validate(ctx); err != nil {
			return fmt.Errorf("response %q: %w", k, err)
		}
	}

	headers := make([]string, 0, len(components.Headers))
	for name := range components.Headers {
		headers = append(headers, name)
	}
	sort.Strings(headers)
	for _, k := range headers {
		v := components.Headers[k]
		if err = ValidateIdentifier(k); err != nil {
			return fmt.Errorf("header %q: %w", k, err)
		}
		if err = v.Validate(ctx); err != nil {
			return fmt.Errorf("header %q: %w", k, err)
		}
	}

	securitySchemes := make([]string, 0, len(components.SecuritySchemes))
	for name := range components.SecuritySchemes {
		securitySchemes = append(securitySchemes, name)
	}
	sort.Strings(securitySchemes)
	for _, k := range securitySchemes {
		v := components.SecuritySchemes[k]
		if err = ValidateIdentifier(k); err != nil {
			return fmt.Errorf("security scheme %q: %w", k, err)
		}
		if err = v.Validate(ctx); err != nil {
			return fmt.Errorf("security scheme %q: %w", k, err)
		}
	}

	examples := make([]string, 0, len(components.Examples))
	for name := range components.Examples {
		examples = append(examples, name)
	}
	sort.Strings(examples)
	for _, k := range examples {
		v := components.Examples[k]
		if err = ValidateIdentifier(k); err != nil {
			return fmt.Errorf("example %q: %w", k, err)
		}
		if err = v.Validate(ctx); err != nil {
			return fmt.Errorf("example %q: %w", k, err)
		}
	}

	links := make([]string, 0, len(components.Links))
	for name := range components.Links {
		links = append(links, name)
	}
	sort.Strings(links)
	for _, k := range links {
		v := components.Links[k]
		if err = ValidateIdentifier(k); err != nil {
			return fmt.Errorf("link %q: %w", k, err)
		}
		if err = v.Validate(ctx); err != nil {
			return fmt.Errorf("link %q: %w", k, err)
		}
	}

	callbacks := make([]string, 0, len(components.Callbacks))
	for name := range components.Callbacks {
		callbacks = append(callbacks, name)
	}
	sort.Strings(callbacks)
	for _, k := range callbacks {
		v := components.Callbacks[k]
		if err = ValidateIdentifier(k); err != nil {
			return fmt.Errorf("callback %q: %w", k, err)
		}
		if err = v.Validate(ctx); err != nil {
			return fmt.Errorf("callback %q: %w", k, err)
		}
	}

	return
}

const identifierPattern = `^[a-zA-Z0-9._-]+$`

// IdentifierRegExp verifies whether Component object key matches 'identifierPattern' pattern, according to OapiAPI v3.x.0.
// Hovever, to be able supporting legacy OpenAPI v2.x, there is a need to customize above pattern in orde not to fail
// converted v2-v3 validation
var IdentifierRegExp = regexp.MustCompile(identifierPattern)

func ValidateIdentifier(value string) error {
	if IdentifierRegExp.MatchString(value) {
		return nil
	}
	return fmt.Errorf("identifier %q is not supported by OpenAPIv3 standard (regexp: %q)", value, identifierPattern)
}
