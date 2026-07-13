package openapi3

import (
	"slices"
	"strings"
	"unicode"
)

// CodedError is implemented by every typed validation error. Code returns a
// stable, kebab-case identifier for the validation rule that failed (e.g.
// "operation-responses-required"), so consumers can suppress specific
// findings, assign per-rule severities, emit machine-readable diagnostics
// (e.g. the LSP Diagnostic.code field), or link to per-rule documentation
// without depending on message text.
//
// Codes are a contract: message wording may change freely, codes may not.
// Renaming or removing a code is a breaking change and is recorded in the
// README changelog. The full set is available from ValidationErrorCodes.
//
// Reach the code of a wrapped error with errors.As:
//
//	var coded openapi3.CodedError
//	if errors.As(err, &coded) {
//		fmt.Println(coded.Code())
//	}
type CodedError interface {
	error
	Code() string
}

func (e *RequiredFieldError) Code() string {
	return codeToken(e.Field) + "-required"
}

func (e *FieldVersionMismatchError) Code() string {
	return codeToken(e.Field) + "-field-for-" + strings.ReplaceAll(e.MinVersion, ".", "-") + "-plus"
}

func (e *SchemaValueError) Code() string {
	return codeToken(e.ValueKind) + "-violates-schema"
}

func (e *MutuallyExclusiveFieldsError) Code() string {
	return codeToken(e.Field1) + "-" + codeToken(e.Field2) + "-mutually-exclusive"
}

func (e *ForbiddenFieldError) Code() string {
	return codeToken(e.Field) + "-forbidden"
}

func (e *EitherFieldRequiredError) Code() string {
	return joinCodeTokens(e.Fields) + "-required"
}

func (e *ExactlyOneFieldError) Code() string {
	return joinCodeTokens(e.Fields) + "-exactly-one"
}

func (e *SchemaBothFormsExclusive) Code() string {
	return codeToken(e.Field) + "-both-forms-exclusive"
}

func (e *SingleEntryContentError) Code() string {
	return codeToken(e.Subject) + "-content-single-entry"
}

func (e *DuplicateRequiredFieldError) Code() string { return "duplicate-required-field" }
func (e *DuplicateTagError) Code() string           { return "duplicate-tag" }
func (e *PathParametersError) Code() string         { return "path-parameters-mismatch" }
func (e *ServerURLTemplateError) Code() string      { return "server-url-template-invalid" }
func (e *WebhookNilError) Code() string             { return "webhook-nil" }
func (e *PathParameterRequiredError) Code() string  { return "path-parameter-required" }
func (e *DuplicateOperationIDError) Code() string   { return "duplicate-operation-id" }
func (e *ExtraSiblingFieldsError) Code() string     { return "extra-sibling-fields" }
func (e *SchemaTypeError) Code() string             { return "schema-type-unsupported" }
func (e *InvalidParameterInError) Code() string     { return "parameter-in-invalid" }
func (e *SchemaPatternRegexError) Code() string     { return "schema-pattern-regex-invalid" }
func (e *InvalidSecuritySchemeTypeError) Code() string {
	return "security-scheme-type-invalid"
}
func (e *InvalidHTTPSchemeError) Code() string      { return "security-scheme-http-scheme-invalid" }
func (e *UnresolvedRefError) Code() string          { return "unresolved-ref" }
func (e *APIKeyInInvalidError) Code() string        { return "security-scheme-apikey-in-invalid" }
func (e *PathMustStartWithSlashError) Code() string { return "path-must-start-with-slash" }
func (e *ConflictingPathsError) Code() string       { return "conflicting-paths" }
func (e *DuplicateParameterError) Code() string     { return "duplicate-parameter" }
func (e *InvalidSerializationMethodError) Code() string {
	return "serialization-method-invalid"
}

// codeToken renders a field path as a code segment: "$" is stripped, "." joins
// with "-", known acronyms stay single words ("oAuthFlow" -> "oauth-flow"),
// and camelCase splits on "-" ("externalDocs.url" -> "external-docs-url").
func codeToken(field string) string {
	field = strings.TrimPrefix(field, "$")
	field = strings.ReplaceAll(field, ".", "-")
	field = strings.ReplaceAll(field, "oAuth", "oauth")
	field = strings.ReplaceAll(field, "openId", "openid")
	var b strings.Builder
	for i, r := range field {
		if i > 0 && unicode.IsUpper(r) {
			b.WriteByte('-')
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}

func joinCodeTokens(fields []string) string {
	tokens := make([]string, len(fields))
	for i, f := range fields {
		tokens[i] = codeToken(f)
	}
	return strings.Join(tokens, "-or-")
}

// ValidationErrorCodes returns every code a validation error can carry,
// sorted. TestValidationErrorCodes keeps it in sync with the errors the
// validators emit.
func ValidationErrorCodes() []string {
	return slices.Clone(validationErrorCodes)
}

var validationErrorCodes = []string{
	"additional-properties-both-forms-exclusive",
	"anchor-field-for-3-1-plus",
	"authorization-url-forbidden",
	"bearer-format-forbidden",
	"comment-field-for-3-1-plus",
	"conflicting-paths",
	"const-field-for-3-1-plus",
	"contains-field-for-3-1-plus",
	"content-encoding-field-for-3-1-plus",
	"content-media-type-field-for-3-1-plus",
	"content-or-schema-exactly-one",
	"content-schema-field-for-3-1-plus",
	"default-required",
	"default-violates-schema",
	"defs-field-for-3-1-plus",
	"dependent-required-field-for-3-1-plus",
	"dependent-schemas-field-for-3-1-plus",
	"duplicate-operation-id",
	"duplicate-parameter",
	"duplicate-required-field",
	"duplicate-tag",
	"dynamic-anchor-field-for-3-1-plus",
	"dynamic-ref-field-for-3-1-plus",
	"else-field-for-3-1-plus",
	"example-examples-mutually-exclusive",
	"example-violates-schema",
	"examples-field-for-3-1-plus",
	"external-docs-url-required",
	"extra-sibling-fields",
	"flows-forbidden",
	"flows-required",
	"header-content-single-entry",
	"id-field-for-3-1-plus",
	"identifier-field-for-3-1-plus",
	"if-field-for-3-1-plus",
	"in-forbidden",
	"info-required",
	"info-title-required",
	"info-version-required",
	"item-schema-field-for-3-2-plus",
	"json-schema-dialect-required",
	"jsonschemadialect-field-for-3-1-plus",
	"license-name-required",
	"max-contains-field-for-3-1-plus",
	"min-contains-field-for-3-1-plus",
	"name-forbidden",
	"oauth-flow-authorization-url-required",
	"oauth-flow-scopes-required",
	"oauth-flow-token-url-required",
	"openapi-required",
	"openid-connect-url-required",
	"operation-id-operation-ref-mutually-exclusive",
	"operation-id-or-operation-ref-required",
	"operation-responses-required",
	"parameter-content-single-entry",
	"parameter-in-invalid",
	"parameter-name-required",
	"path-must-start-with-slash",
	"path-parameter-required",
	"path-parameters-mismatch",
	"paths-required",
	"pattern-properties-field-for-3-1-plus",
	"prefix-items-field-for-3-1-plus",
	"property-names-field-for-3-1-plus",
	"read-only-write-only-mutually-exclusive",
	"request-body-content-required",
	"response-description-required",
	"responses-required",
	"schema-field-for-3-1-plus",
	"schema-items-required",
	"schema-pattern-regex-invalid",
	"schema-type-unsupported",
	"security-scheme-apikey-in-invalid",
	"security-scheme-http-scheme-invalid",
	"security-scheme-name-required",
	"security-scheme-type-invalid",
	"serialization-method-invalid",
	"server-url-required",
	"server-url-template-invalid",
	"summary-field-for-3-1-plus",
	"then-field-for-3-1-plus",
	"token-url-forbidden",
	"unevaluated-items-both-forms-exclusive",
	"unevaluated-items-field-for-3-1-plus",
	"unevaluated-properties-both-forms-exclusive",
	"unevaluated-properties-field-for-3-1-plus",
	"unresolved-ref",
	"url-identifier-mutually-exclusive",
	"value-external-value-mutually-exclusive",
	"value-or-external-value-required",
	"webhook-nil",
	"webhooks-field-for-3-1-plus",
}
