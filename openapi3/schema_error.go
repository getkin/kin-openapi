package openapi3

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// ErrSchemaErrorDetailsWithJSONSchema2020 is returned when validating with both
// DisableSchemaErrorDetails and EnableJSONSchema2020, the latter being used for
// OpenAPI 3.1+ documents: the JSON Schema 2020-12 validator includes values in
// the reason of its errors, so disabling details would not keep them out of messages.
var ErrSchemaErrorDetailsWithJSONSchema2020 = errors.New("the setting DisableSchemaErrorDetails() is ineffective together with EnableJSONSchema2020() (used for OpenAPI 3.1+ documents): the latter's error messages are opaque")

// SchemaError is an error that occurs during schema validation.
type SchemaError struct {
	// Value is the value that failed validation.
	Value any
	// reversePath is the path to the value that failed validation.
	reversePath []string
	// Schema is the schema that failed validation.
	Schema *Schema
	// SchemaField is the field of the schema that failed validation.
	SchemaField string
	// Reason is a human-readable message describing the error.
	// The message should never include the original value to prevent leakage of potentially sensitive inputs in error messages.
	Reason string
	// Origin is the original error that caused this error.
	Origin error
	// cause is the error of the JSON Schema 2020-12 validator this error was converted from.
	cause error
	// customizeMessageError is a function that can be used to customize the error message.
	customizeMessageError func(err *SchemaError) string
	// detailsDisabled leaves the schema and the value out of the error message.
	detailsDisabled bool
}

func markSchemaErrorKey(err error, key string) error {
	switch v := err.(type) {
	case *SchemaError:
		v.reversePath = append(v.reversePath, key)
		if v.Origin != nil {
			if unwrapped := errors.Unwrap(v.Origin); unwrapped != nil {
				if me, ok := unwrapped.(multiErrorForOneOf); ok {
					_ = markSchemaErrorKey(MultiError(me), key)
				}
				if me, ok := unwrapped.(multiErrorForAllOf); ok {
					_ = markSchemaErrorKey(MultiError(me), key)
				}
			}
		}
		return v
	case MultiError:
		for _, e := range v {
			_ = markSchemaErrorKey(e, key)
		}
		return v
	}
	return err
}

func markSchemaErrorIndex(err error, index int) error {
	return markSchemaErrorKey(err, strconv.FormatInt(int64(index), 10))
}

func (err *SchemaError) JSONPointer() []string {
	reversePath := err.reversePath
	path := append([]string(nil), reversePath...)
	for left, right := 0, len(path)-1; left < right; left, right = left+1, right-1 {
		path[left], path[right] = path[right], path[left]
	}
	return path
}

func (err *SchemaError) Error() string {
	if err.customizeMessageError != nil {
		if msg := err.customizeMessageError(err); msg != "" {
			return msg
		}
	}

	buf := bytes.NewBuffer(make([]byte, 0, 256))

	// The Reason of an error converted from the JSON Schema 2020-12 validator
	// already names the path.
	if len(err.reversePath) > 0 && err.cause == nil {
		buf.WriteString(`Error at "`)
		reversePath := err.reversePath
		for _, r := range slices.Backward(reversePath) {
			buf.WriteByte('/')
			buf.WriteString(r)
		}
		buf.WriteString(`": `)
	}

	if err.Origin != nil {
		buf.WriteString(err.Origin.Error())

		return buf.String()
	}

	reason := err.Reason
	if reason == "" {
		buf.WriteString(`Doesn't match schema "`)
		buf.WriteString(err.SchemaField)
		buf.WriteString(`"`)
	} else {
		buf.WriteString(reason)
	}

	if !err.detailsDisabled {
		buf.WriteString("\nSchema:\n  ")
		encoder := json.NewEncoder(buf)
		encoder.SetIndent("  ", "  ")
		if err := encoder.Encode(err.Schema); err != nil {
			panic(err)
		}
		buf.WriteString("\nValue:\n  ")
		if err := encoder.Encode(err.Value); err != nil {
			panic(err)
		}
	}

	return buf.String()
}

func (err SchemaError) Unwrap() error {
	if err.Origin == nil && err.cause != nil {
		return err.cause
	}
	return err.Origin
}

// newSchemaError starts building a SchemaError about value failing schema.
func newSchemaError(schema *Schema, value any) *SchemaError {
	return &SchemaError{
		Value:  value,
		Schema: schema,
	}
}

// newSchemaError starts building a SchemaError about value failing schema,
// carrying over the error message settings of settings, which may be nil.
func (settings *schemaValidationSettings) newSchemaError(schema *Schema, value any) *SchemaError {
	err := newSchemaError(schema, value)
	if settings != nil {
		err.customizeMessageError = settings.customizeMessageError
		err.detailsDisabled = settings.schemaErrorDetailsDisabled
	}
	return err
}

// formatError describes err, as returned by a validator of format for a value of kind.
func (settings *schemaValidationSettings) formatError(kind, format string, err error) (reason string, origin error) {
	prefix := fmt.Sprintf("%s doesn't match the format %q", kind, format)
	reasons := schemaErrorReasons(nil, err)
	if len(reasons) == 0 {
		return fmt.Sprintf("%s (%v)", prefix, err), fmt.Errorf("%s: %w", prefix, err)
	}
	joined := strings.Join(reasons, " | ")
	reason = fmt.Sprintf("%s (%s)", prefix, joined)
	if settings.schemaErrorDetailsDisabled {
		// The SchemaErrors of format validators are built without settings,
		// and the messages of the errors wrapping them were rendered with their details.
		return reason, &messageError{msg: prefix + ": " + joined, err: err}
	}
	return reason, fmt.Errorf("%s: %w", prefix, err)
}

// schemaErrorReasons appends to reasons those of the outermost SchemaErrors within err.
func schemaErrorReasons(reasons []string, err error) []string {
	switch e := err.(type) {
	case nil:
	case *SchemaError:
		reasons = append(reasons, e.Reason)
	case MultiError:
		for _, err := range e {
			reasons = schemaErrorReasons(reasons, err)
		}
	case interface{ Unwrap() error }:
		reasons = schemaErrorReasons(reasons, e.Unwrap())
	case interface{ Unwrap() []error }:
		for _, err := range e.Unwrap() {
			reasons = schemaErrorReasons(reasons, err)
		}
	}
	return reasons
}

// messageError replaces the message of the error it wraps.
type messageError struct {
	msg string
	err error
}

func (e *messageError) Error() string { return e.msg }

func (e *messageError) Unwrap() error { return e.err }

// withPathKey appends key to the path of the value that failed validation.
func (err *SchemaError) withPathKey(key string) *SchemaError {
	err.reversePath = append(err.reversePath, key)
	return err
}

func (err *SchemaError) withField(field string) *SchemaError {
	err.SchemaField = field
	return err
}

func (err *SchemaError) withReason(reason string) *SchemaError {
	err.Reason = reason
	return err
}

func (err *SchemaError) withOrigin(origin error) *SchemaError {
	err.Origin = origin
	return err
}

// withCause records the error of the JSON Schema 2020-12 validator this error is converted from.
func (err *SchemaError) withCause(cause error) *SchemaError {
	err.cause = cause
	return err
}
