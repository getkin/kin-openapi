package openapi3

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/santhosh-tekuri/jsonschema/v6/kind"
)

// jsonSchemaValidator wraps the santhosh-tekuri/jsonschema validator
type jsonSchemaValidator struct {
	compiler        *jsonschema.Compiler
	schema          *jsonschema.Schema
	hasInternalRefs bool
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
	hasInternalRefs := containsInternalSchemaRef(schemaDocument)
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
		compiler:        compiler,
		schema:          compiledSchema,
		hasInternalRefs: hasInternalRefs,
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

// addInternalSchemaRefs preserves document-relative references present on a
// resolved Schema. VisitJSON receives a subtree rather than the containing
// OpenAPI document, so copy each resolved target into local $defs and rewrite
// its reference. This also handles references to a nested property or item.
func addInternalSchemaRefs(schemaMap map[string]any, schema *Schema) error {
	if schemaMap == nil || schema == nil {
		return nil
	}
	refs := make(map[string]any)
	// Recursive schemas reference an already-visited target, which WalkSubtree would not report.
	err := NewSchemaRef("", schema).walkSubtreeRefs(func(_ string, ref *SchemaRef) error {
		if ref.Ref == "" || !strings.HasPrefix(ref.Ref, "#/components/schemas/") || ref.Value == nil {
			return nil
		}
		if _, ok := refs[ref.Ref]; ok {
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

// validate validates a value against the compiled JSON Schema
func (v *jsonSchemaValidator) validate(value any) error {
	if err := v.schema.Validate(value); err != nil {
		// Convert jsonschema error to SchemaError
		return convertJSONSchemaError(err, value)
	}
	return nil
}

// convertJSONSchemaError converts a jsonschema validation error to OpenAPI SchemaError format.
// root is the validated value, from which the failing value is looked up.
func convertJSONSchemaError(err error, root any) error {
	// TODO: Go 1.26
	// if err, ok := errors.AsType[*jsonschema.ValidationError](err); ok {
	// 	return formatValidationError(err, "")
	var validationErr *jsonschema.ValidationError
	if errors.As(err, &validationErr) {
		return formatValidationError(validationErr, "", root)
	}
	return err
}

// formatValidationError recursively formats validation errors into SchemaErrors
// that carry the same structure as the ones of the built-in validator: the path
// of the failing value (JSONPointer), the failing keyword (SchemaField) and the
// failing value (Value). The jsonschema error itself stays reachable with
// errors.As.
func formatValidationError(verr *jsonschema.ValidationError, parentPath string, root any) error {
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

	// A wrapper with a single chain of causes (the schema itself, a $ref) is
	// about the value its innermost cause is about.
	failing := verr
	for len(failing.Causes) == 1 && isWrapperErrorKind(failing.ErrorKind) {
		failing = failing.Causes[0]
	}

	schemaErr := &SchemaError{
		Reason:      msg.String(),
		reversePath: reversedPath(failing.InstanceLocation),
		cause:       verr,
	}
	if keywords := failing.ErrorKind.KeywordPath(); len(keywords) > 0 {
		schemaErr.SchemaField = keywords[len(keywords)-1]
	}
	if value, ok := valueAt(root, failing.InstanceLocation); ok {
		schemaErr.Value = value
	}

	// If there are sub-errors, format them too
	if len(verr.Causes) > 0 {
		var subErrors MultiError
		for _, cause := range verr.Causes {
			if subErr := formatValidationError(cause, path, root); subErr != nil {
				subErrors = append(subErrors, subErr)
			}
		}
		if len(subErrors) > 0 {
			schemaErr.Origin = fmt.Errorf("validation failed due to: %w", subErrors)
		}
	}

	return schemaErr
}

// isWrapperErrorKind reports whether k only groups the errors of its causes.
func isWrapperErrorKind(k jsonschema.ErrorKind) bool {
	switch k.(type) {
	case *kind.Schema, *kind.Group, *kind.Reference:
		return true
	}
	return false
}

// reversedPath is location in the reversed order SchemaError keeps its path in.
func reversedPath(location []string) []string {
	if len(location) == 0 {
		return nil
	}
	reversed := make([]string /*,*/, len(location))
	for i, token := range location {
		reversed[len(location)-1-i] = token
	}
	return reversed
}

// valueAt is the value at location (unescaped JSON pointer tokens) inside root.
func valueAt(root any, location []string) (any, bool) {
	value := root
	for _, token := range location {
		switch v := value.(type) {
		case map[string]any:
			next, ok := v[token]
			if !ok {
				return nil, false
			}
			value = next
		case []any:
			i, err := strconv.Atoi(token)
			if err != nil || i < 0 || i >= len(v) {
				return nil, false
			}
			value = v[i]
		default:
			return nil, false
		}
	}
	return value, true
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
	if validator.hasInternalRefs && !usesJSONSchema2020Features {
		return schema.visitJSON(settings, value)
	}

	return validator.validate(value)
}

// schemaUsesJSONSchema2020Features checks the keywords that guard the legacy
// fallback. Annotations such as examples must not turn an ordinary reference
// into a compilation error instead of a fallback.
func schemaUsesJSONSchema2020Features(schema *Schema) bool {
	if schema == nil {
		return false
	}
	// WalkSubtree follows resolved refs and guards against cycles. Stop at the
	// first validation keyword rather than serializing the schema just to inspect it.
	err := NewSchemaRef("", schema).WalkSubtree(func(_ string, ref *SchemaRef) error {
		if schemaHasJSONSchema2020FallbackGuardKeyword(ref.Value) {
			return errJSONSchema2020FeatureFound
		}
		return nil
	})
	return errors.Is(err, errJSONSchema2020FeatureFound)
}

var errJSONSchema2020FeatureFound = errors.New("json schema 2020 validation keyword found")

func schemaHasJSONSchema2020FallbackGuardKeyword(schema *Schema) bool {
	return schema.Const != nil ||
		len(schema.PrefixItems) != 0 || schema.Contains != nil ||
		len(schema.PatternProperties) != 0 ||
		len(schema.DependentSchemas) != 0 || schema.PropertyNames != nil ||
		schema.UnevaluatedItems.Has != nil || schema.UnevaluatedItems.Schema != nil ||
		schema.UnevaluatedProperties.Has != nil || schema.UnevaluatedProperties.Schema != nil ||
		schema.If != nil || schema.Then != nil || schema.Else != nil ||
		len(schema.DependentRequired) != 0 || len(schema.Defs) != 0 ||
		schema.ContentSchema != nil
}
