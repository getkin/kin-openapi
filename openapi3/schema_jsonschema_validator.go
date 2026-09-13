package openapi3

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// jsonSchemaValidator wraps the santhosh-tekuri/jsonschema validator
type jsonSchemaValidator struct {
	compiler                   *jsonschema.Compiler
	schema                     *jsonschema.Schema
	hasInternalRefs            bool
	usesJSONSchema2020Features bool
}

// newJSONSchemaValidator creates a new validator using JSON Schema 2020-12
func newJSONSchemaValidator(schema *Schema, settings *schemaValidationSettings) (*jsonSchemaValidator, error) {
	// Convert OpenAPI Schema to JSON Schema format
	schemaBytes, err := json.Marshal(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema: %w", err)
	}

	var schemaDocument any
	if err := json.Unmarshal(schemaBytes, &schemaDocument); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema: %w", err)
	}

	// Boolean schemas are valid JSON Schema resources as well.
	schemaMap, _ := schemaDocument.(map[string]any)
	transformOpenAPIToJSONSchema(schemaMap)
	hasInternalRefs := containsInternalSchemaRef(schemaDocument)
	usesJSONSchema2020Features := hasJSONSchema2020Features(schemaDocument)
	if err := addInternalSchemaRefs(schemaMap, schema); err != nil {
		return nil, fmt.Errorf("failed to prepare schema references: %w", err)
	}

	// Create compiler
	compiler := jsonschema.NewCompiler()
	compiler.DefaultDraft(jsonschema.Draft2020)

	// Keep enforcing the formats the built-in validator enforces
	registerFormatValidators(compiler, schemaMap, settings)

	// Add the schema
	schemaURL := "https://example.com/schema.json"
	if err := compiler.AddResource(schemaURL, schemaDocument); err != nil {
		return nil, fmt.Errorf("failed to add schema resource: %w", err)
	}

	// Compile the schema
	compiledSchema, err := compiler.Compile(schemaURL)
	if err != nil {
		return nil, fmt.Errorf("failed to compile schema: %w", err)
	}

	return &jsonSchemaValidator{
		compiler:                   compiler,
		schema:                     compiledSchema,
		hasInternalRefs:            hasInternalRefs,
		usesJSONSchema2020Features: usesJSONSchema2020Features,
	}, nil
}

func containsInternalSchemaRef(node any) bool {
	switch node := node.(type) {
	case map[string]any:
		if ref, ok := node["$ref"].(string); ok && strings.HasPrefix(ref, "#/components/schemas/") {
			return true
		}
		for _, value := range node {
			if containsInternalSchemaRef(value) {
				return true
			}
		}
	case []any:
		for _, value := range node {
			if containsInternalSchemaRef(value) {
				return true
			}
		}
	}
	return false
}

func hasJSONSchema2020Features(node any) bool {
	features := map[string]struct{}{
		"const": {}, "contains": {}, "contentSchema": {}, "dependentRequired": {},
		"dependentSchemas": {}, "$defs": {}, "if": {}, "then": {}, "else": {},
		"prefixItems": {}, "propertyNames": {}, "unevaluatedItems": {},
		"unevaluatedProperties": {},
	}
	switch node := node.(type) {
	case map[string]any:
		for key, value := range node {
			if _, ok := features[key]; ok {
				return true
			}
			if hasJSONSchema2020Features(value) {
				return true
			}
		}
	case []any:
		for _, value := range node {
			if hasJSONSchema2020Features(value) {
				return true
			}
		}
	}
	return false
}

// addInternalSchemaRefs preserves document-relative references present on a
// resolved Schema. VisitJSON receives a subtree rather than the containing
// OpenAPI document, so copy each resolved target into local $defs and rewrite
// its reference. This also handles references to a nested property or item.
func addInternalSchemaRefs(schemaMap map[string]any, schema *Schema) error {
	if schemaMap == nil || schema == nil {
		return nil
	}
	refs := make(map[string]any)
	err := NewSchemaRef("", schema).WalkSubtree(func(_ string, ref *SchemaRef) error {
		if ref.Ref == "" || !strings.HasPrefix(ref.Ref, "#/components/schemas/") || ref.Value == nil {
			return nil
		}
		var target any
		bytes, err := json.Marshal(ref.Value)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(bytes, &target); err != nil {
			return err
		}
		if targetMap, ok := target.(map[string]any); ok {
			transformOpenAPIToJSONSchema(targetMap)
		}
		refs[ref.Ref] = target
		return nil
	})
	if err != nil {
		return err
	}
	if len(refs) == 0 {
		return nil
	}
	defs := make(map[string]any, len(refs))
	ids := make(map[string]string, len(refs))
	for ref := range refs {
		ids[ref] = fmt.Sprintf("ref%d", len(ids))
	}
	rewriteInternalSchemaRefs(schemaMap, refs, ids, defs)
	if len(defs) > 0 {
		schemaMap["$defs"] = defs
		rewriteInternalSchemaRefs(defs, refs, ids, defs)
	}
	return nil
}

func rewriteInternalSchemaRefs(node any, refs map[string]any, ids map[string]string, defs map[string]any) {
	switch node := node.(type) {
	case map[string]any:
		if ref, ok := node["$ref"].(string); ok {
			if id, exists := ids[ref]; exists {
				node["$ref"] = "#/$defs/" + id
				if _, added := defs[id]; !added {
					defs[id] = refs[ref]
				}
			}
		}
		for _, value := range node {
			rewriteInternalSchemaRefs(value, refs, ids, defs)
		}
	case []any:
		for _, value := range node {
			rewriteInternalSchemaRefs(value, refs, ids, defs)
		}
	}
}

func registerFormatValidators(compiler *jsonschema.Compiler, schemaMap map[string]any, settings *schemaValidationSettings) {
	formats := make(map[string]struct{})
	collectFormats(schemaMap, formats)
	if len(formats) == 0 {
		return
	}

	for format := range formats {
		compiler.RegisterFormat(&jsonschema.Format{
			Name:     format,
			Validate: formatValidator(format, settings),
		})
	}
	compiler.AssertFormat() // has to be explicitly asserted
}

func collectFormats(node any, formats map[string]struct{}) {
	switch node := node.(type) {
	case map[string]any:
		if format, ok := node["format"].(string); ok && format != "" {
			formats[format] = struct{}{}
		}
		for _, value := range node {
			collectFormats(value, formats)
		}
	case []any:
		for _, value := range node {
			collectFormats(value, formats)
		}
	}
}

func formatValidator(format string, settings *schemaValidationSettings) func(any) error {
	return func(value any) error {
		switch value := value.(type) {
		case string:
			f, ok := settings.stringFormats[format]
			if !ok {
				if f, ok = SchemaStringFormats[format]; !ok {
					return nil
				}
			}
			return f.Validate(value)
		case json.Number:
			if number, err := value.Float64(); err == nil {
				return validateNumberFormat(format, settings, number)
			}
		case float64:
			return validateNumberFormat(format, settings, value)
		case float32:
			return validateNumberFormat(format, settings, float64(value))
		case int:
			return validateNumberFormat(format, settings, float64(value))
		case int32:
			return validateNumberFormat(format, settings, float64(value))
		case int64:
			return validateNumberFormat(format, settings, float64(value))
		}
		return nil
	}
}

func validateNumberFormat(format string, settings *schemaValidationSettings, value float64) error {
	if value == math.Trunc(value) && !math.IsInf(value, 0) {
		f, ok := settings.integerFormats[format]
		if !ok {
			f, ok = SchemaIntegerFormats[format]
		}
		if ok {
			return f.Validate(int64(value))
		}
	}

	f, ok := settings.numberFormats[format]
	if !ok {
		if f, ok = SchemaNumberFormats[format]; !ok {
			return nil
		}
	}
	return f.Validate(value)
}

// transformOpenAPIToJSONSchema converts OpenAPI 3.0/3.1 specific keywords to JSON Schema format
func transformOpenAPIToJSONSchema(schema map[string]any) {
	// Handle nullable - in OpenAPI 3.0, nullable is a boolean flag
	// In OpenAPI 3.1 / JSON Schema 2020-12, we use type arrays
	if nullable, ok := schema["nullable"].(bool); ok && nullable {
		if typeVal, ok := schema["type"].(string); ok {
			// Convert to type array with null
			schema["type"] = []any{typeVal, "null"}
		} else if _, hasType := schema["type"]; !hasType {
			// nullable: true without type - add "null" to allow null values
			schema["type"] = []any{"null"}
		}
		delete(schema, "nullable")
	}

	// Handle exclusiveMinimum/exclusiveMaximum
	// In OpenAPI 3.0, these are booleans alongside minimum/maximum
	// In JSON Schema 2020-12, they are numeric values
	if exclusiveMin, ok := schema["exclusiveMinimum"].(bool); ok {
		if exclusiveMin {
			if schemaMin, ok := schema["minimum"].(float64); ok {
				schema["exclusiveMinimum"] = schemaMin
				delete(schema, "minimum")
			} else {
				delete(schema, "exclusiveMinimum")
			}
		} else {
			// exclusiveMinimum: false means inclusive, which is the JSON Schema default
			delete(schema, "exclusiveMinimum")
		}
	}
	if exclusiveMax, ok := schema["exclusiveMaximum"].(bool); ok {
		if exclusiveMax {
			if schemaMax, ok := schema["maximum"].(float64); ok {
				schema["exclusiveMaximum"] = schemaMax
				delete(schema, "maximum")
			} else {
				delete(schema, "exclusiveMaximum")
			}
		} else {
			// exclusiveMaximum: false means inclusive, which is the JSON Schema default
			delete(schema, "exclusiveMaximum")
		}
	}

	// Remove OpenAPI-specific keywords that aren't in JSON Schema
	delete(schema, "discriminator")
	delete(schema, "xml")
	delete(schema, "externalDocs")
	delete(schema, "example") // Use "examples" in 2020-12

	// Recursively transform nested schemas (single schema fields)
	for _, key := range []string{
		"additionalProperties", "items", "not",
		// OpenAPI 3.1 / JSON Schema 2020-12 fields
		"contains", "propertyNames", "unevaluatedItems", "unevaluatedProperties",
		"if", "then", "else", "contentSchema",
	} {
		if val, ok := schema[key]; ok {
			if nestedSchema, ok := val.(map[string]any); ok {
				transformOpenAPIToJSONSchema(nestedSchema)
			}
		}
	}

	// Transform schema arrays (oneOf, anyOf, allOf, prefixItems)
	for _, key := range []string{"oneOf", "anyOf", "allOf", "prefixItems"} {
		if val, ok := schema[key].([]any); ok {
			for _, item := range val {
				if nestedSchema, ok := item.(map[string]any); ok {
					transformOpenAPIToJSONSchema(nestedSchema)
				}
			}
		}
	}

	// Transform schema maps (properties, patternProperties, dependentSchemas, $defs)
	for _, key := range []string{"properties", "patternProperties", "dependentSchemas", "$defs"} {
		if props, ok := schema[key].(map[string]any); ok {
			for _, propVal := range props {
				if propSchema, ok := propVal.(map[string]any); ok {
					transformOpenAPIToJSONSchema(propSchema)
				}
			}
		}
	}
}

// validate validates a value against the compiled JSON Schema
func (v *jsonSchemaValidator) validate(value any) error {
	if err := v.schema.Validate(value); err != nil {
		// Convert jsonschema error to SchemaError
		return convertJSONSchemaError(err)
	}
	return nil
}

// convertJSONSchemaError converts a jsonschema validation error to OpenAPI SchemaError format
func convertJSONSchemaError(err error) error {
	// TODO: Go 1.26
	// if err, ok := errors.AsType[*jsonschema.ValidationError](err); ok {
	// 	return formatValidationError(err, "")
	var validationErr *jsonschema.ValidationError
	if errors.As(err, &validationErr) {
		return formatValidationError(validationErr, "")
	}
	return err
}

// formatValidationError recursively formats validation errors
func formatValidationError(verr *jsonschema.ValidationError, parentPath string) error {
	// Build the path from InstanceLocation slice
	path := "/" + strings.Join(verr.InstanceLocation, "/")
	if parentPath != "" && path != "/" {
		path = parentPath + path
	} else if path == "/" {
		path = parentPath
	}

	// Build error message using the Error() method
	var msg strings.Builder
	if path != "" {
		fmt.Fprintf(&msg, `error at "%s": `, path)
	}
	msg.WriteString(verr.Error())

	// If there are sub-errors, format them too
	if len(verr.Causes) > 0 {
		var subErrors MultiError
		for _, cause := range verr.Causes {
			if subErr := formatValidationError(cause, path); subErr != nil {
				subErrors = append(subErrors, subErr)
			}
		}
		if len(subErrors) > 0 {
			return &SchemaError{
				Reason: msg.String(),
				Origin: fmt.Errorf("validation failed due to: %w", subErrors),
			}
		}
	}

	return &SchemaError{
		Reason: msg.String(),
	}
}

// useJSONSchema2020 validates using the JSON Schema 2020-12 validator
func (schema *Schema) useJSONSchema2020(settings *schemaValidationSettings, value any) error {
	usesJSONSchema2020Features := schemaUsesJSONSchema2020Features(schema)
	validator, err := newJSONSchemaValidator(schema, settings)
	if err != nil {
		if !usesJSONSchema2020Features {
			return schema.visitJSON(settings, value)
		}
		// A resolved OpenAPI component reference can be registered above. If a
		// reference still cannot be resolved, do not silently drop its constraints.
		// Keep the existing fallback for other compiler limitations (for example,
		// regex syntax accepted by OpenAPI but rejected by Go's regexp package).
		if strings.Contains(err.Error(), "json-pointer") && strings.Contains(err.Error(), "not found") {
			return err
		}
		return schema.visitJSON(settings, value)
	}
	if validator.hasInternalRefs && !validator.usesJSONSchema2020Features {
		return schema.visitJSON(settings, value)
	}

	return validator.validate(value)
}

func schemaUsesJSONSchema2020Features(schema *Schema) bool {
	bytes, err := json.Marshal(schema)
	if err != nil {
		return false
	}
	var document any
	if err := json.Unmarshal(bytes, &document); err != nil {
		return false
	}
	return hasJSONSchema2020Features(document)
}
