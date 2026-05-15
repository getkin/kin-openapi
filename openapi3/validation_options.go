package openapi3

import "context"

// ValidationOption allows the modification of how the OpenAPI document is validated.
type ValidationOption func(options *ValidationOptions)

// ValidationOptions provides configuration for validating OpenAPI documents.
type ValidationOptions struct {
	examplesValidationAsReq, examplesValidationAsRes bool
	examplesValidationDisabled                       bool
	schemaDefaultsValidationDisabled                 bool
	schemaFormatValidationEnabled                    bool
	schemaPatternValidationDisabled                  bool
	schemaExtensionsInRefProhibited                  bool
	jsonSchema2020ValidationEnabled                  bool
	isOpenAPI31OrLater                               bool
	multiErrorEnabled                                bool
	regexCompilerFunc                                RegexCompilerFunc
	extraSiblingFieldsAllowed                        map[string]struct{}
}

type validationOptionsKey struct{}

// AllowExtraSiblingFields called as AllowExtraSiblingFields("description") makes Validate not return an error when said field appears next to a $ref.
func AllowExtraSiblingFields(fields ...string) ValidationOption {
	return func(options *ValidationOptions) {
		if options.extraSiblingFieldsAllowed == nil && len(fields) != 0 {
			options.extraSiblingFieldsAllowed = make(map[string]struct{}, len(fields))
		}
		for _, field := range fields {
			options.extraSiblingFieldsAllowed[field] = struct{}{}
		}
	}
}

// IsOpenAPI31OrLater enables "JSON Schema Draft 2020-12"-compliant validation (for OpenAPI 3.1 documents).
func IsOpenAPI31OrLater() ValidationOption {
	return func(options *ValidationOptions) {
		options.isOpenAPI31OrLater = true              // To distinguish from v3.0
		options.jsonSchema2020ValidationEnabled = true // TODO: use even for v3.0
	}
}

// EnableSchemaFormatValidation makes Validate not return an error when validating documents that mention schema formats that are not defined by the OpenAPIv3 specification.
// By default, schema format validation is disabled.
func EnableSchemaFormatValidation() ValidationOption {
	return func(options *ValidationOptions) {
		options.schemaFormatValidationEnabled = true
	}
}

// DisableSchemaFormatValidation does the opposite of EnableSchemaFormatValidation.
// By default, schema format validation is disabled.
func DisableSchemaFormatValidation() ValidationOption {
	return func(options *ValidationOptions) {
		options.schemaFormatValidationEnabled = false
	}
}

// EnableSchemaPatternValidation does the opposite of DisableSchemaPatternValidation.
// By default, schema pattern validation is enabled.
func EnableSchemaPatternValidation() ValidationOption {
	return func(options *ValidationOptions) {
		options.schemaPatternValidationDisabled = false
	}
}

// DisableSchemaPatternValidation makes Validate not return an error when validating patterns that are not supported by the Go regexp engine.
func DisableSchemaPatternValidation() ValidationOption {
	return func(options *ValidationOptions) {
		options.schemaPatternValidationDisabled = true
	}
}

// EnableSchemaDefaultsValidation does the opposite of DisableSchemaDefaultsValidation.
// By default, schema default values are validated against their schema.
func EnableSchemaDefaultsValidation() ValidationOption {
	return func(options *ValidationOptions) {
		options.schemaDefaultsValidationDisabled = false
	}
}

// DisableSchemaDefaultsValidation disables schemas' default field validation.
// By default, schema default values are validated against their schema.
func DisableSchemaDefaultsValidation() ValidationOption {
	return func(options *ValidationOptions) {
		options.schemaDefaultsValidationDisabled = true
	}
}

// EnableExamplesValidation does the opposite of DisableExamplesValidation.
// By default, all schema examples are validated.
func EnableExamplesValidation() ValidationOption {
	return func(options *ValidationOptions) {
		options.examplesValidationDisabled = false
	}
}

// DisableExamplesValidation disables all example schema validation.
// By default, all schema examples are validated.
func DisableExamplesValidation() ValidationOption {
	return func(options *ValidationOptions) {
		options.examplesValidationDisabled = true
	}
}

// AllowExtensionsWithRef allows extensions (fields starting with 'x-')
// as siblings for $ref fields. This is the default.
// Non-extension fields are prohibited unless allowed explicitly with the
// AllowExtraSiblingFields option.
func AllowExtensionsWithRef() ValidationOption {
	return func(options *ValidationOptions) {
		options.schemaExtensionsInRefProhibited = false
	}
}

// ProhibitExtensionsWithRef causes the validation to return an
// error if extensions (fields starting with 'x-') are found as
// siblings for $ref fields. Non-extension fields are prohibited
// unless allowed explicitly with the AllowExtraSiblingFields option.
func ProhibitExtensionsWithRef() ValidationOption {
	return func(options *ValidationOptions) {
		options.schemaExtensionsInRefProhibited = true
	}
}

// EnableMultiError makes Validate aggregate independent validation errors and
// return them all as a MultiError, instead of returning the first one and stopping.
//
// Aggregation happens at container fan-out points (the document root, Paths,
// PathItem, Operation, Components, Responses, Webhooks). Validation of a single
// leaf element (for example, a single Schema) still stops at its first error.
//
// By default, Validate returns the first error encountered (fail-fast).
//
// Note: callers should use errors.As / errors.Is to inspect the returned error,
// since MultiError.As and MultiError.Is walk into the contained errors.
func EnableMultiError() ValidationOption {
	return func(options *ValidationOptions) {
		options.multiErrorEnabled = true
	}
}

// SetRegexCompiler allows to override the regex implementation used to validate
// field "pattern".
func SetRegexCompiler(c RegexCompilerFunc) ValidationOption {
	return func(options *ValidationOptions) {
		options.regexCompilerFunc = c
	}
}

// WithValidationOptions allows adding validation options to a context object that can be used when validating any OpenAPI type.
func WithValidationOptions(ctx context.Context, opts ...ValidationOption) context.Context {
	if len(opts) == 0 {
		return ctx
	}
	options := &ValidationOptions{}
	for _, opt := range opts {
		opt(options)
	}
	return context.WithValue(ctx, validationOptionsKey{}, options)
}

func getValidationOptions(ctx context.Context) *ValidationOptions {
	if options, ok := ctx.Value(validationOptionsKey{}).(*ValidationOptions); ok {
		return options
	}
	return &ValidationOptions{}
}

// errCollector aggregates validation errors at a container fan-out point.
//
// When multi-error mode is enabled (EnableMultiError), emit records the error
// and returns nil so the caller continues to the next sibling; if the error is
// itself a MultiError, its leaves are appended individually so the result is a
// flat MultiError of fully-wrapped problems (this matches what most consumers
// expect — one MultiError entry per independent problem).
//
// When multi-error mode is off, emit returns the error unchanged so the caller
// fails fast — preserving the historical behavior byte-for-byte.
//
// emitWrapped applies wrap to err, distributing wrap over each leaf when err
// is a MultiError. This is how container validators attach per-section /
// per-path / per-operation context to each aggregated leaf.
//
// result returns the accumulated MultiError, or nil if none were recorded.
type errCollector struct {
	multi bool
	errs  MultiError
}

func newErrCollector(ctx context.Context) *errCollector {
	return &errCollector{multi: getValidationOptions(ctx).multiErrorEnabled}
}

func (c *errCollector) emit(err error) error {
	if err == nil {
		return nil
	}
	if !c.multi {
		return err
	}
	if me, ok := err.(MultiError); ok {
		for _, sub := range me {
			if e := c.emit(sub); e != nil {
				return e
			}
		}
		return nil
	}
	c.errs = append(c.errs, err)
	return nil
}

func (c *errCollector) emitWrapped(wrap func(error) error, err error) error {
	if err == nil {
		return nil
	}
	if !c.multi {
		return wrap(err)
	}
	if me, ok := err.(MultiError); ok {
		for _, sub := range me {
			if e := c.emitWrapped(wrap, sub); e != nil {
				return e
			}
		}
		return nil
	}
	return c.emit(wrap(err))
}

func (c *errCollector) result() error {
	if len(c.errs) > 0 {
		return c.errs
	}
	return nil
}

// finalize emits err (typically the last sibling validation in a container,
// e.g. the extensions check) and returns the accumulated result. It collapses
// the trailing emit-then-result pattern into a single line at each call site.
func (c *errCollector) finalize(err error) error {
	if e := c.emit(err); e != nil {
		return e
	}
	return c.result()
}
