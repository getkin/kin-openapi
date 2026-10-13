package openapi3

import (
	"encoding/json"
	"fmt"
	"maps"
	"math"
	"math/big"
	"reflect"
	"slices"
	"strings"
	"unicode/utf16"
)

func (schema *Schema) IsMatching(value any) bool {
	settings := newSchemaValidationSettings(FailFast())
	return schema.visitJSON(settings, value) == nil
}

func (schema *Schema) IsMatchingJSONBoolean(value bool) bool {
	settings := newSchemaValidationSettings(FailFast())
	return schema.visitJSON(settings, value) == nil
}

func (schema *Schema) IsMatchingJSONNumber(value float64) bool {
	settings := newSchemaValidationSettings(FailFast())
	return schema.visitJSON(settings, value) == nil
}

func (schema *Schema) IsMatchingJSONString(value string) bool {
	settings := newSchemaValidationSettings(FailFast())
	return schema.visitJSON(settings, value) == nil
}

func (schema *Schema) IsMatchingJSONArray(value []any) bool {
	settings := newSchemaValidationSettings(FailFast())
	return schema.visitJSON(settings, value) == nil
}

func (schema *Schema) IsMatchingJSONObject(value map[string]any) bool {
	settings := newSchemaValidationSettings(FailFast())
	return schema.visitJSON(settings, value) == nil
}

// visitChildJSON validates a value nested within the one currently visited,
// such as an array item or an object property.
// Cycle detection only applies to schemas visiting the same value,
// so recursive schemas still get applied to nested values.
func (schema *Schema) visitChildJSON(settings *schemaValidationSettings, value any) error {
	var start int
	start, settings.visitedBase = settings.visitedBase, len(settings.visitedSchemas)
	defer func() { settings.visitedBase = start }()
	return schema.visitJSON(settings, value)
}

func (schema *Schema) visitJSON(settings *schemaValidationSettings, value any) (err error) {
	if slices.Contains(settings.visitedSchemas[settings.visitedBase:], schema) {
		return nil
	}
	depth := len(settings.visitedSchemas)
	settings.visitedSchemas = append(settings.visitedSchemas, schema)
	defer func() { settings.visitedSchemas = settings.visitedSchemas[:depth] }()

	switch value := value.(type) {
	case nil:
		// Don't use VisitJSONNull, as we still want to reach 'visitXOFOperations', since
		// those could allow for a nullable value even though this one doesn't
		if schema.permitsNull(settings) {
			return
		}
	case float64:
		if math.IsNaN(value) {
			return ErrSchemaInputNaN
		}
		if math.IsInf(value, 0) {
			return ErrSchemaInputInf
		}
	}

	if x := schema.Always; x != nil {
		if *x {
			return
		}
		if settings.failfast {
			return errSchema
		}
		return settings.newSchemaError(schema, value).
			withField("boolean").
			withReason("the false schema accepts no value")
	}

	if schema.IsEmpty() {
		switch value.(type) {
		case nil:
			return schema.visitJSONNull(settings)
		default:
			return
		}
	}

	if err = schema.visitNotOperation(settings, value); err != nil {
		return
	}
	var run bool
	if err, run = schema.visitXOFOperations(settings, value); err != nil || !run {
		return
	}
	if err = schema.visitEnumOperation(settings, value); err != nil {
		return
	}

	switch value := value.(type) {
	case nil:
		return schema.visitJSONNull(settings)
	case bool:
		return schema.visitJSONBoolean(settings, value)
	case json.Number:
		valueFloat64, err := value.Float64()
		if err != nil {
			return settings.newSchemaError(schema, value).
				withField("type").
				withReason("cannot convert json.Number to float64").
				withOrigin(err)
		}
		return schema.visitJSONNumber(settings, valueFloat64)
	case int:
		return schema.visitJSONNumber(settings, float64(value))
	case int32:
		return schema.visitJSONNumber(settings, float64(value))
	case int64:
		return schema.visitJSONNumber(settings, float64(value))
	case float64:
		return schema.visitJSONNumber(settings, value)
	case string:
		return schema.visitJSONString(settings, value)
	case []any:
		return schema.visitJSONArray(settings, value)
	case map[string]any:
		return schema.visitJSONObject(settings, value)
	case map[any]any: // for YAML cf. issue https://github.com/getkin/kin-openapi/issues/444
		values := make(map[string]any, len(value))
		for key, v := range value {
			if k, ok := key.(string); ok {
				values[k] = v
			}
		}
		if len(value) == len(values) {
			return schema.visitJSONObject(settings, values)
		}
	}

	// Catch slice of non-empty interface type
	if reflect.TypeOf(value).Kind() == reflect.Slice {
		valueR := reflect.ValueOf(value)
		newValue := make([]any, 0, valueR.Len())
		for i := 0; i < valueR.Len(); i++ {
			newValue = append(newValue, valueR.Index(i).Interface())
		}
		return schema.visitJSONArray(settings, newValue)
	}

	return settings.newSchemaError(schema, value).
		withField("type").
		withReason(fmt.Sprintf("unhandled value of type %T", value))
}

func (schema *Schema) visitEnumOperation(settings *schemaValidationSettings, value any) (err error) {
	if enum := schema.Enum; len(enum) != 0 {
		for _, v := range enum {
			if equalJSONValues(v, value) {
				return
			}
		}
		if settings.failfast {
			return errSchema
		}
		allowedValues, _ := json.Marshal(enum)
		return settings.newSchemaError(schema, value).
			withField("enum").
			withReason("value is not one of the allowed values " + string(allowedValues))
	}
	return
}

// Helps with `1 == json.Number("1")`
func equalJSONValues(a, b any) bool {
	if x, ok := jsonNumberAsFloat64(a); ok {
		y, ok := jsonNumberAsFloat64(b)
		return ok && x == y
	}
	switch a := a.(type) {
	case []any:
		b, ok := b.([]any)
		return ok && slices.EqualFunc(a, b, equalJSONValues)
	case map[string]any:
		b, ok := b.(map[string]any)
		return ok && maps.EqualFunc(a, b, equalJSONValues)
	}
	return reflect.DeepEqual(a, b)
}

func jsonNumberAsFloat64(v any) (float64, bool) {
	switch v := v.(type) {
	case json.Number:
		f, err := v.Float64()
		return f, err == nil
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	}
	return 0, false
}

func (schema *Schema) visitNotOperation(settings *schemaValidationSettings, value any) (err error) {
	if ref := schema.Not; ref != nil {
		v := ref.Value
		if v == nil {
			return newUnresolvedRef(ref.Ref, ref.Origin)
		}
		if err := v.visitJSON(settings, value); err == nil {
			if settings.failfast {
				return errSchema
			}
			return settings.newSchemaError(schema, value).withField("not")
		}
	}
	return
}

// If the XOF operations pass successfully, abort further run of validation, as they will already be satisfied (unless the schema
// itself is badly specified
// resolveDiscriminatorRef resolves the discriminator reference for oneOf/anyOf validation.
// Returns the discriminator ref string and any error encountered during resolution.
func (schema *Schema) resolveDiscriminatorRef(settings *schemaValidationSettings, value any) (string, error) {
	if schema.Discriminator == nil {
		return "", nil
	}
	pn := schema.Discriminator.PropertyName
	valuemap, okcheck := value.(map[string]any)
	if !okcheck {
		return "", nil
	}
	discriminatorVal, okcheck := valuemap[pn]
	if !okcheck {
		return "", settings.newSchemaError(schema, nil).
			withField("discriminator").
			withReason(fmt.Sprintf("input does not contain the discriminator property %q", pn))
	}

	discriminatorValString, okcheck := discriminatorVal.(string)
	if !okcheck {
		return "", settings.newSchemaError(schema, discriminatorVal).
			withField("discriminator").
			withReason(fmt.Sprintf("value of discriminator property %q is not a string", pn))
	}

	if discriminatorRef, okcheck := schema.Discriminator.Mapping[discriminatorValString]; len(schema.Discriminator.Mapping) > 0 && !okcheck {
		return "", settings.newSchemaError(schema, discriminatorVal).
			withField("discriminator").
			withReason(fmt.Sprintf("discriminator property %q has invalid value", pn))
	} else if ref := discriminatorRef.Ref; ref != "" && !strings.Contains(ref, "/") {
		// Mapping values can also be plain schema names like "Dog"
		return "#/components/schemas/" + ref, nil
	} else {
		return ref, nil
	}
}

func (schema *Schema) visitXOFOperations(settings *schemaValidationSettings, value any) (err error, run bool) {
	var visitedOneOf, visitedAnyOf, visitedAllOf bool
	if v := schema.OneOf; len(v) > 0 {
		discriminatorRef, err := schema.resolveDiscriminatorRef(settings, value)
		if err != nil {
			return err, false
		}

		var (
			ok                  = 0
			validationErrors    = multiErrorForOneOf{}
			matchedOneOfIndices = make([]int, 0)
			tempValue           = value
		)
		for idx, item := range v {
			v := item.Value
			if v == nil {
				return newUnresolvedRef(item.Ref, item.Origin), false
			}

			if discriminatorRef != "" && discriminatorRef != item.Ref {
				continue
			}

			// make a deep copy to protect origin value from being injected default value that defined in mismatched oneOf schema
			if settings.asreq || settings.asrep {
				tempValue = deepCopyJSONValue(value)
			}

			if err := v.visitJSON(settings, tempValue); err != nil {
				validationErrors = append(validationErrors, err)
				continue
			}

			matchedOneOfIndices = append(matchedOneOfIndices, idx)
			ok++
		}

		if ok != 1 {
			if settings.failfast {
				return errSchema, false
			}
			e := settings.newSchemaError(schema, value).withField("oneOf")
			if ok > 1 {
				e.withReason(fmt.Sprintf(`value matches more than one schema from "oneOf" (matches schemas at indices %v)`, matchedOneOfIndices)).
					withOrigin(ErrOneOfConflict)
			} else {
				e.withReason(`value doesn't match any schema from "oneOf"`).
					withOrigin(fmt.Errorf("doesn't match schema due to: %w", validationErrors))
			}

			return e, false
		}

		// run again to inject default value that defined in matched oneOf schema
		if settings.asreq || settings.asrep {
			_ = v[matchedOneOfIndices[0]].Value.visitJSON(settings, value)
		}
		visitedOneOf = true
	}

	if v := schema.AnyOf; len(v) > 0 {
		discriminatorRef, err := schema.resolveDiscriminatorRef(settings, value)
		if err != nil {
			return err, false
		}

		var (
			ok              = false
			matchedAnyOfIdx = 0
			tempValue       = value
		)
		for idx, item := range v {
			v := item.Value
			if v == nil {
				return newUnresolvedRef(item.Ref, item.Origin), false
			}

			if discriminatorRef != "" && discriminatorRef != item.Ref {
				continue
			}

			// make a deep copy to protect origin value from being injected default value that defined in mismatched anyOf schema
			if settings.asreq || settings.asrep {
				tempValue = deepCopyJSONValue(value)
			}
			if err := v.visitJSON(settings, tempValue); err == nil {
				ok = true
				matchedAnyOfIdx = idx
				break
			}
		}
		if !ok {
			if settings.failfast {
				return errSchema, false
			}
			return settings.newSchemaError(schema, value).
				withField("anyOf").
				withReason(`doesn't match any schema from "anyOf"`), false
		}

		_ = v[matchedAnyOfIdx].Value.visitJSON(settings, value)
		visitedAnyOf = true
	}

	validationErrors := multiErrorForAllOf{}
	for _, item := range schema.AllOf {
		v := item.Value
		if v == nil {
			return newUnresolvedRef(item.Ref, item.Origin), false
		}
		if err := v.visitJSON(settings, value); err != nil {
			if settings.failfast {
				return errSchema, false
			}
			validationErrors = append(validationErrors, err)
		}
		visitedAllOf = true
	}
	if len(validationErrors) > 0 {
		return settings.newSchemaError(schema, value).
			withField("allOf").
			withReason(`doesn't match all schemas from "allOf"`).
			withOrigin(fmt.Errorf("doesn't match schema due to: %w", validationErrors)), false
	}

	run = !((visitedOneOf || visitedAnyOf || visitedAllOf) && value == nil)
	return
}

// The value is not considered in visitJSONNull because according to the spec
// "null is not supported as a type" unless `nullable` is also set to true
// https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.3.md#data-types
// https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.3.md#schema-object
func (schema *Schema) visitJSONNull(settings *schemaValidationSettings) (err error) {
	if schema.permitsNull(settings) {
		return
	}
	if settings.failfast {
		return errSchema
	}
	return settings.newSchemaError(schema, nil).
		withField("nullable").
		withReason("Value is not nullable")
}

func (schema *Schema) VisitJSONBoolean(value bool) error {
	settings := newSchemaValidationSettings()
	return schema.visitJSONBoolean(settings, value)
}

func (schema *Schema) visitJSONBoolean(settings *schemaValidationSettings, value bool) (err error) {
	if !schema.Type.Permits(TypeBoolean) {
		return schema.expectedType(settings, value)
	}
	return
}

// multipleOfSlack bounds, as a negative power of two, how far a value may sit
// from the nearest multiple and still count as one, measured relative to the
// value itself. A float64 carries about 2^-53 of relative error, so this is a
// couple of units in the last place: enough to absorb the error a caller's own
// arithmetic introduced, small enough that a genuine remainder never fits.
//
// The window must be measured against the value and not against the quotient.
// A quotient-relative window grows without bound, and once value/multipleOf
// reaches 2^49 it exceeds one half, at which point every number validates.
const multipleOfSlack = 51

// multipleOfFailure reports whether value satisfies multipleOf, and the reason
// when it does not.
//
// The comparison runs on exact rationals rather than on decimals formatted at a
// fixed precision. A fixed precision is only correct inside a narrow magnitude
// band: above it the formatting emits binary representation error and rejects
// valid values, and below it every operand formats to zero, which both accepts
// any value and divides by zero.
func multipleOfFailure(value, multipleOf float64) (string, bool) {
	// A non-finite multipleOf cannot be reported: SchemaError renders the whole
	// schema as JSON and NaN has no JSON encoding, so building that error would
	// panic instead of describing the problem. Treat it as no constraint, which
	// at least replaces the division-by-zero panic the fixed-precision
	// formatting produced here.
	if math.IsNaN(multipleOf) || math.IsInf(multipleOf, 0) {
		return "", true
	}
	// JSON Schema requires multipleOf to be strictly greater than zero. Without
	// this guard a zero or negative value either divided by zero or accepted
	// every number.
	if multipleOf <= 0 {
		return fmt.Sprintf("multipleOf must be greater than zero, got %g", multipleOf), false
	}
	// VisitJSON rejects a non-finite number before any keyword runs, but the
	// exported VisitJSONNumber does not, so one can still reach this point.
	// Reporting it here is out of scope: SchemaError marshals Value, and NaN has
	// no JSON encoding, so the error would panic when rendered. Preserve the
	// existing outcome and leave that entry point to a separate change.
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return "", true
	}
	notAMultiple := fmt.Sprintf("number must be a multiple of %g", multipleOf)

	exactValue := new(big.Rat).SetFloat64(value)
	exactMultipleOf := new(big.Rat).SetFloat64(multipleOf)
	quotient := new(big.Rat).Quo(exactValue, exactMultipleOf)
	if quotient.IsInt() {
		return "", true
	}

	// The quotient is not an integer, so floor and floor+1 bracket it. Both are
	// computed through big.Int, which no quotient can overflow. Compare the two
	// candidate multiples against the value itself rather than comparing the
	// quotient against an integer, so the window stays tied to the value's own
	// precision at every magnitude.
	floor := new(big.Int).Div(quotient.Num(), quotient.Denom())
	var distance *big.Rat
	for _, k := range []*big.Int{floor, new(big.Int).Add(floor, big.NewInt(1))} {
		gap := new(big.Rat).Sub(exactValue, new(big.Rat).Mul(new(big.Rat).SetInt(k), exactMultipleOf))
		gap.Abs(gap)
		if distance == nil || gap.Cmp(distance) < 0 {
			distance = gap
		}
	}

	slack := new(big.Rat).SetFrac(big.NewInt(1), new(big.Int).Lsh(big.NewInt(1), multipleOfSlack))
	tolerance := new(big.Rat).Mul(new(big.Rat).Abs(exactValue), slack)
	if distance.Cmp(tolerance) <= 0 {
		return "", true
	}
	return notAMultiple, false
}

func (schema *Schema) VisitJSONNumber(value float64) error {
	settings := newSchemaValidationSettings()
	return schema.visitJSONNumber(settings, value)
}

func (schema *Schema) visitJSONNumber(settings *schemaValidationSettings, value float64) error {
	var me MultiError
	schemaType := schema.Type
	requireInteger := false
	if schemaType.Permits(TypeInteger) && !schemaType.Permits(TypeNumber) {
		requireInteger = true
		if bigFloat := big.NewFloat(value); !bigFloat.IsInt() {
			if settings.failfast {
				return errSchema
			}
			err := settings.newSchemaError(schema, value).
				withField("type").
				withReason("value must be an integer")
			if !settings.multiError {
				return err
			}
			me = append(me, err)
		}
	} else if !(schemaType.Permits(TypeInteger) || schemaType.Permits(TypeNumber)) {
		return schema.expectedType(settings, value)
	}

	// formats
	var formatStrErr string
	var formatErr error
	format := schema.Format
	if format != "" {
		if requireInteger {
			// Check per-validation validators first, then fall back to global
			f, ok := settings.integerFormats[format]
			if !ok {
				f, ok = SchemaIntegerFormats[format]
			}
			if ok {
				if err := validateIntegerFormat(f, value); err != nil {
					formatStrErr, formatErr = settings.formatError("integer", format, err)
				}
			}
		} else {
			// Check per-validation validators first, then fall back to global
			f, ok := settings.numberFormats[format]
			if !ok {
				f, ok = SchemaNumberFormats[format]
			}
			if ok {
				if err := f.Validate(value); err != nil {
					formatStrErr, formatErr = settings.formatError("number", format, err)
				}
			}
		}
	}

	if formatStrErr != "" || formatErr != nil {
		err := settings.newSchemaError(schema, value).
			withField("format").
			withReason(formatStrErr).
			withOrigin(formatErr)
		if !settings.multiError {
			return err
		}
		me = append(me, err)
	}

	// "exclusiveMinimum"
	// OpenAPI 3.0: boolean modifier for minimum
	// OpenAPI 3.1: number representing the actual exclusive bound
	if eb := schema.ExclusiveMin; eb.IsSet() {
		var exclusiveMinBound float64
		var valid bool
		if eb.Value != nil {
			// OpenAPI 3.1 style: exclusiveMinimum is the bound itself
			exclusiveMinBound = *eb.Value
			valid = value > exclusiveMinBound
		} else if eb.Bool != nil && *eb.Bool && schema.Min != nil {
			// OpenAPI 3.0 style: exclusiveMinimum modifies minimum
			exclusiveMinBound = *schema.Min
			valid = value > exclusiveMinBound
		} else {
			valid = true
		}
		if !valid {
			if settings.failfast {
				return errSchema
			}
			err := settings.newSchemaError(schema, value).
				withField("exclusiveMinimum").
				withReason(fmt.Sprintf("number must be more than %g", exclusiveMinBound))
			if !settings.multiError {
				return err
			}
			me = append(me, err)
		}
	}

	// "exclusiveMaximum"
	// OpenAPI 3.0: boolean modifier for maximum
	// OpenAPI 3.1: number representing the actual exclusive bound
	if eb := schema.ExclusiveMax; eb.IsSet() {
		var exclusiveMaxBound float64
		var valid bool
		if eb.Value != nil {
			// OpenAPI 3.1 style: exclusiveMaximum is the bound itself
			exclusiveMaxBound = *eb.Value
			valid = value < exclusiveMaxBound
		} else if eb.Bool != nil && *eb.Bool && schema.Max != nil {
			// OpenAPI 3.0 style: exclusiveMaximum modifies maximum
			exclusiveMaxBound = *schema.Max
			valid = value < exclusiveMaxBound
		} else {
			valid = true
		}
		if !valid {
			if settings.failfast {
				return errSchema
			}
			err := settings.newSchemaError(schema, value).
				withField("exclusiveMaximum").
				withReason(fmt.Sprintf("number must be less than %g", exclusiveMaxBound))
			if !settings.multiError {
				return err
			}
			me = append(me, err)
		}
	}

	// "minimum"
	if v := schema.Min; v != nil && !(*v <= value) {
		if settings.failfast {
			return errSchema
		}
		err := settings.newSchemaError(schema, value).
			withField("minimum").
			withReason(fmt.Sprintf("number must be at least %g", *v))
		if !settings.multiError {
			return err
		}
		me = append(me, err)
	}

	// "maximum"
	if v := schema.Max; v != nil && !(*v >= value) {
		if settings.failfast {
			return errSchema
		}
		err := settings.newSchemaError(schema, value).
			withField("maximum").
			withReason(fmt.Sprintf("number must be at most %g", *v))
		if !settings.multiError {
			return err
		}
		me = append(me, err)
	}

	// "multipleOf"
	if v := schema.MultipleOf; v != nil {
		// "A numeric instance is valid only if division by this keyword's
		//    value results in an integer."
		reason, ok := multipleOfFailure(value, *v)
		if !ok {
			if settings.failfast {
				return errSchema
			}
			err := settings.newSchemaError(schema, value).
				withField("multipleOf").
				withReason(reason)
			if !settings.multiError {
				return err
			}
			me = append(me, err)
		}
	}

	if len(me) > 0 {
		return me
	}

	return nil
}

func (schema *Schema) VisitJSONString(value string) error {
	settings := newSchemaValidationSettings()
	return schema.visitJSONString(settings, value)
}

func (schema *Schema) visitJSONString(settings *schemaValidationSettings, value string) error {
	if !schema.Type.Permits(TypeString) {
		return schema.expectedType(settings, value)
	}

	var me MultiError

	// "minLength" and "maxLength"
	minLength := schema.MinLength
	maxLength := schema.MaxLength
	if minLength != 0 || maxLength != nil {
		// JSON schema string lengths are UTF-16, not UTF-8!
		length := int64(0)
		for _, r := range value {
			if utf16.IsSurrogate(r) {
				length += 2
			} else {
				length++
			}
		}
		if minLength != 0 && length < int64(minLength) {
			if settings.failfast {
				return errSchema
			}
			err := settings.newSchemaError(schema, value).
				withField("minLength").
				withReason(fmt.Sprintf("minimum string length is %d", minLength))
			if !settings.multiError {
				return err
			}
			me = append(me, err)
		}
		if maxLength != nil && length > int64(*maxLength) {
			if settings.failfast {
				return errSchema
			}
			err := settings.newSchemaError(schema, value).
				withField("maxLength").
				withReason(fmt.Sprintf("maximum string length is %d", *maxLength))
			if !settings.multiError {
				return err
			}
			me = append(me, err)
		}
	}

	// "pattern"
	if !settings.patternValidationDisabled && schema.Pattern != "" {
		cpiface, _ := compiledPatterns.Load(schema.Pattern)
		cp, _ := cpiface.(RegexMatcher)
		if cp == nil {
			var err error
			if cp, err = schema.compilePattern(settings, settings.regexCompiler); err != nil {
				if !settings.multiError {
					return err
				}
				me = append(me, err)
			}
		}
		if cp != nil && !cp.MatchString(value) {
			err := settings.newSchemaError(schema, value).
				withField("pattern").
				withReason(fmt.Sprintf(`string doesn't match the regular expression "%s"`, schema.Pattern))
			if !settings.multiError {
				return err
			}
			me = append(me, err)
		}
	}

	// "format"
	var formatStrErr string
	var formatErr error
	if format := schema.Format; format != "" {
		// Check per-validation validators first, then fall back to global
		f, ok := settings.stringFormats[format]
		if !ok {
			f, ok = SchemaStringFormats[format]
		}
		if ok {
			if err := f.Validate(value); err != nil {
				formatStrErr, formatErr = settings.formatError("string", format, err)
			}
		}
	}
	if formatStrErr != "" || formatErr != nil {
		err := settings.newSchemaError(schema, value).
			withField("format").
			withReason(formatStrErr).
			withOrigin(formatErr)
		if !settings.multiError {
			return err
		}
		me = append(me, err)

	}

	if len(me) > 0 {
		return me
	}

	return nil
}

func (schema *Schema) VisitJSONArray(value []any) error {
	settings := newSchemaValidationSettings()
	return schema.visitJSONArray(settings, value)
}

func (schema *Schema) visitJSONArray(settings *schemaValidationSettings, value []any) error {
	if !schema.Type.Permits(TypeArray) {
		return schema.expectedType(settings, value)
	}

	var me MultiError

	lenValue := int64(len(value))

	// "minItems"
	if v := schema.MinItems; v != 0 && lenValue < int64(v) {
		if settings.failfast {
			return errSchema
		}
		err := settings.newSchemaError(schema, value).
			withField("minItems").
			withReason(fmt.Sprintf("minimum number of items is %d", v))
		if !settings.multiError {
			return err
		}
		me = append(me, err)
	}

	// "maxItems"
	if v := schema.MaxItems; v != nil && lenValue > int64(*v) {
		if settings.failfast {
			return errSchema
		}
		err := settings.newSchemaError(schema, value).
			withField("maxItems").
			withReason(fmt.Sprintf("maximum number of items is %d", *v))
		if !settings.multiError {
			return err
		}
		me = append(me, err)
	}

	// "uniqueItems"
	if v := schema.UniqueItems; v && !sliceUniqueItemsChecker(value) {
		if settings.failfast {
			return errSchema
		}
		err := settings.newSchemaError(schema, value).
			withField("uniqueItems").
			withReason("duplicate items found")
		if !settings.multiError {
			return err
		}
		me = append(me, err)
	}

	// "items" applies to the positions prefixItems does not cover. Without
	// prefixItems that is every position, which is the OAS 3.0 behaviour.
	// prefixItems itself is OpenAPI 3.1 only and left to the JSON Schema 2020-12
	// validator (EnableJSONSchema2020), like the other 3.1 sub-schema keywords.
	if itemSchemaRef := schema.Items; itemSchemaRef != nil {
		itemSchema := itemSchemaRef.Value
		if itemSchema == nil {
			return newUnresolvedRef(itemSchemaRef.Ref, itemSchemaRef.Origin)
		}
		for i := len(schema.PrefixItems); i < len(value); i++ {
			if err := itemSchema.visitChildJSON(settings, value[i]); err != nil {
				err = markSchemaErrorIndex(err, i)
				if !settings.multiError {
					return err
				}
				if itemMe, ok := err.(MultiError); ok {
					me = append(me, itemMe...)
				} else {
					me = append(me, err)
				}
			}
		}
	}

	if len(me) > 0 {
		return me
	}

	return nil
}

func (schema *Schema) VisitJSONObject(value map[string]any) error {
	settings := newSchemaValidationSettings()
	return schema.visitJSONObject(settings, value)
}

func (schema *Schema) visitJSONObject(settings *schemaValidationSettings, value map[string]any) error {
	if !schema.Type.Permits(TypeObject) {
		return schema.expectedType(settings, value)
	}

	var me MultiError

	if settings.asreq || settings.asrep {
		for _, propName := range componentNames(schema.Properties) {
			propSchema := schema.Properties[propName]
			reqRO := settings.asreq && propSchema.Value.ReadOnly && !settings.readOnlyValidationDisabled
			repWO := settings.asrep && propSchema.Value.WriteOnly && !settings.writeOnlyValidationDisabled

			if f := settings.defaultsSet; f != nil && value[propName] == nil {
				if dflt := propSchema.Value.Default; dflt != nil && !reqRO && !repWO {
					value[propName] = deepCopyJSONValue(dflt)
					settings.onceSettingDefaults.Do(f)
				}
			}

			if v := value[propName]; v != nil && (reqRO || repWO) {
				if settings.failfast {
					return errSchema
				}
				field, direction := "readOnly", "request"
				if repWO {
					field, direction = "writeOnly", "response"
				}
				err := settings.newSchemaError(propSchema.Value, v).
					withPathKey(propName).
					withField(field).
					withReason(fmt.Sprintf("%s property %q in %s", field, propName, direction))
				if !settings.multiError {
					return err
				}
				me = append(me, err)
			}
		}
	}

	// "properties"
	properties := schema.Properties
	lenValue := int64(len(value))

	// "minProperties"
	if v := schema.MinProps; v != 0 && lenValue < int64(v) {
		if settings.failfast {
			return errSchema
		}
		err := settings.newSchemaError(schema, value).
			withField("minProperties").
			withReason(fmt.Sprintf("there must be at least %d properties", v))
		if !settings.multiError {
			return err
		}
		me = append(me, err)
	}

	// "maxProperties"
	if v := schema.MaxProps; v != nil && lenValue > int64(*v) {
		if settings.failfast {
			return errSchema
		}
		err := settings.newSchemaError(schema, value).
			withField("maxProperties").
			withReason(fmt.Sprintf("there must be at most %d properties", *v))
		if !settings.multiError {
			return err
		}
		me = append(me, err)
	}

	// "additionalProperties"
	var additionalProperties *Schema
	if ref := schema.AdditionalProperties.Schema; ref != nil {
		additionalProperties = ref.Value
	}
	for _, k := range componentNames(value) {
		v := value[k]
		if properties != nil {
			propertyRef := properties[k]
			if propertyRef != nil {
				p := propertyRef.Value
				if p == nil {
					return newUnresolvedRef(propertyRef.Ref, propertyRef.Origin)
				}
				if err := p.visitChildJSON(settings, v); err != nil {
					if settings.failfast {
						return errSchema
					}
					err = markSchemaErrorKey(err, k)
					if !settings.multiError {
						return err
					}
					if v, ok := err.(MultiError); ok {
						me = append(me, v...)
						continue
					}
					me = append(me, err)
				}
				continue
			}
		}
		if allowed := schema.AdditionalProperties.Has; allowed == nil || *allowed {
			if additionalProperties != nil {
				if err := additionalProperties.visitChildJSON(settings, v); err != nil {
					if settings.failfast {
						return errSchema
					}
					err = markSchemaErrorKey(err, k)
					if !settings.multiError {
						return err
					}
					if v, ok := err.(MultiError); ok {
						me = append(me, v...)
						continue
					}
					me = append(me, err)
				}
			}
			continue
		}
		if settings.failfast {
			return errSchema
		}
		err := settings.newSchemaError(schema, value).
			withField("properties").
			withReason(fmt.Sprintf("property %q is unsupported", k))
		if !settings.multiError {
			return err
		}
		me = append(me, err)
	}

	// "required"
	for _, k := range schema.Required {
		if _, ok := value[k]; !ok {
			if s := schema.Properties[k]; s != nil && s.Value.ReadOnly && settings.asreq {
				continue
			}
			if s := schema.Properties[k]; s != nil && s.Value.WriteOnly && settings.asrep {
				continue
			}
			if settings.failfast {
				return errSchema
			}
			err := settings.newSchemaError(schema, value).
				withPathKey(k).
				withField("required").
				withReason(fmt.Sprintf("property %q is missing", k))
			if !settings.multiError {
				return err
			}
			me = append(me, err)
		}
	}

	if len(me) > 0 {
		return me
	}

	return nil
}

func (schema *Schema) expectedType(settings *schemaValidationSettings, value any) error {
	if settings.failfast {
		return errSchema
	}

	a := "a"
	var x string
	schemaTypes := (*schema.Type)
	if len(schemaTypes) == 1 {
		x = schemaTypes[0]
		switch x {
		case TypeArray, TypeObject, TypeInteger:
			a = "an"
		}
	} else {
		a = "one of"
		x = strings.Join(schemaTypes, ", ")
	}
	return settings.newSchemaError(schema, value).
		withField("type").
		withReason(fmt.Sprintf("value must be %s %s", a, x))
}

func isSliceOfUniqueItems(xs []any) bool {
	s := len(xs)
	m := make(map[string]struct{}, s)
	for _, x := range xs {
		// The input slice is converted from a JSON string, there shall
		// have no error when convert it back.
		key, _ := json.Marshal(&x)
		m[string(key)] = struct{}{}
	}
	return s == len(m)
}

// SliceUniqueItemsChecker is an function used to check if an given slice
// have unique items.
type SliceUniqueItemsChecker func(items []any) bool

// By default using predefined func isSliceOfUniqueItems which make use of
// json.Marshal to generate a key for map used to check if a given slice
// have unique items.
var sliceUniqueItemsChecker SliceUniqueItemsChecker = isSliceOfUniqueItems

// RegisterArrayUniqueItemsChecker is used to register a customized function
// used to check if JSON array have unique items.
// Passing nil restores the default checker.
func RegisterArrayUniqueItemsChecker(fn SliceUniqueItemsChecker) {
	if fn == nil {
		fn = isSliceOfUniqueItems
	}
	sliceUniqueItemsChecker = fn
}

func unsupportedFormat(format string) error {
	return fmt.Errorf("unsupported 'format' value %q", format)
}

func deepCopyJSONValue(v any) any {
	switch v := v.(type) {
	case map[string]any:
		cp := make(map[string]any, len(v))
		for k, val := range v {
			cp[k] = deepCopyJSONValue(val)
		}
		return cp
	case []any:
		cp := make([]any, 0, len(v))
		for _, val := range v {
			cp = append(cp, deepCopyJSONValue(val))
		}
		return cp
	default:
		return v // string, float64, bool, nil — all immutable
	}
}
