[![CI](https://github.com/getkin/kin-openapi/workflows/go/badge.svg)](https://github.com/getkin/kin-openapi/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/getkin/kin-openapi.svg)](https://pkg.go.dev/github.com/getkin/kin-openapi)
[![Join Gitter Chat Channel -](https://badges.gitter.im/getkin/kin.svg)](https://gitter.im/getkin/kin?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
[![inspect.software](https://raw.githubusercontent.com/inspect-software/badges/main/v1/g/getkin/kin-openapi.svg)](https://inspect.software/software/getkin/kin-openapi)

# Introduction
A [Go](https://go.dev) project for handling [OpenAPI](https://www.openapis.org/) files. We target:
* [OpenAPI `v2.0`](https://github.com/OAI/OpenAPI-Specification/blob/main/versions/2.0.md) (formerly known as Swagger)
* [OpenAPI `v3.0`](https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.3.md)
* [OpenAPI `v3.1`](https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.1.0.md)
* [OpenAPI `v3.2`](https://spec.openapis.org/oas/v3.2.0.html) Partially: Media Type Object `itemSchema`, Path Item Object `query` (the HTTP `QUERY` method) and `additionalOperations` (custom HTTP methods).

Licensed under the [MIT License](./LICENSE).

## Contributors, users and sponsors
The project has received pull requests [from many people](https://github.com/getkin/kin-openapi/graphs/contributors). Thanks to everyone!

Please, [give back to this project](https://github.com/sponsors/fenollp) by becoming a sponsor.

Here's some projects that depend on _kin-openapi_:
  * [github.com/go-fuego/fuego](https://github.com/go-fuego/fuego) - "Framework generating OpenAPI 3 spec from source code"
  * [github.com/a-h/rest](https://github.com/a-h/rest) - "Generate OpenAPI 3.0 specifications from Go code without annotations or magic comments"
  * [github.com/Tufin/oasdiff](https://github.com/Tufin/oasdiff) - "A diff tool for OpenAPI Specification 3"
  * [github.com/danielgtaylor/apisprout](https://github.com/danielgtaylor/apisprout) - "Lightweight, blazing fast, cross-platform OpenAPI 3 mock server with validation"
  * [github.com/oapi-codegen/oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) - "Generate Go client and server boilerplate from OpenAPI 3 specifications"
  * [github.com/dunglas/vulcain](https://github.com/dunglas/vulcain) - "Use HTTP/2 Server Push to create fast and idiomatic client-driven REST APIs"
  * [github.com/danielgtaylor/restish](https://github.com/danielgtaylor/restish) - "...a CLI for interacting with REST-ish HTTP APIs with some nice features built-in"
  * [github.com/goadesign/goa](https://github.com/goadesign/goa) - "Design-based APIs and microservices in Go"
  * [github.com/hashicorp/nomad-openapi](https://github.com/hashicorp/nomad-openapi) - "Nomad is an easy-to-use, flexible, and performant workload orchestrator that can deploy a mix of microservice, batch, containerized, and non-containerized applications. Nomad is easy to operate and scale and has native Consul and Vault integrations."
  * [gitlab.com/jamietanna/httptest-openapi](https://gitlab.com/jamietanna/httptest-openapi) ([*blog post*](https://www.jvt.me/posts/2022/05/22/go-openapi-contract-test/)) - "Go OpenAPI Contract Verification for use with `net/http`"
  * [github.com/SIMITGROUP/openapigenerator](https://github.com/SIMITGROUP/openapigenerator) - "Openapi v3 microservices generator"
  * [https://github.com/projectsveltos/addon-controller](https://github.com/projectsveltos/addon-controller) - "Kubernetes add-on controller designed to manage tens of clusters."
  * (Feel free to add your project by [creating an issue](https://github.com/getkin/kin-openapi/issues/new) or a pull request)

## Alternatives
* [libopenapi](https://github.com/pb33f/libopenapi) a fully featured, high performance OpenAPI 3.1, 3.0 and Swagger parser, library, validator and toolkit
* [go-swagger](https://github.com/go-swagger/go-swagger) stated [*OpenAPIv3 won't be supported*](https://github.com/go-swagger/go-swagger/issues/1122#issuecomment-575968499)
* [swaggo](https://github.com/swaggo/swag) has an [open issue on OpenAPIv3](https://github.com/swaggo/swag/issues/386)

Be sure to check [OpenAPI Initiative](https://github.com/OAI)'s [great tooling list](https://github.com/OAI/OpenAPI-Specification/blob/master/IMPLEMENTATIONS.md) as well as [OpenAPI.Tools](https://openapi.tools/).

# Structure
  * _openapi2_ ([Go Reference](https://pkg.go.dev/github.com/getkin/kin-openapi/openapi2))
    * Support for OpenAPI 2 files, including serialization, deserialization, and validation.
  * _openapi2conv_ ([Go Reference](https://pkg.go.dev/github.com/getkin/kin-openapi/openapi2conv))
    * Converts OpenAPI 2 files into OpenAPI 3 files.
  * _openapi3_ ([Go Reference](https://pkg.go.dev/github.com/getkin/kin-openapi/openapi3))
    * Support for OpenAPI 3 files, including serialization, deserialization, and validation.
  * _openapi3filter_ ([Go Reference](https://pkg.go.dev/github.com/getkin/kin-openapi/openapi3filter))
    * Validates HTTP requests and responses
    * Provides a [gorilla/mux](https://github.com/gorilla/mux) router for OpenAPI operations
  * _openapi3gen_ ([Go Reference](https://pkg.go.dev/github.com/getkin/kin-openapi/openapi3gen))
    * Generates `*openapi3.Schema` values for Go types.

# Some recipes
## Validating an OpenAPI document
```shell
go run github.com/getkin/kin-openapi/cmd/validate@latest [--defaults] [--examples] [--ext] [--patterns] -- <local YAML or JSON file>
```

## Loading OpenAPI document
Use `openapi3.Loader`, which resolves all references:
```go
loader := openapi3.NewLoader()
doc, err := loader.LoadFromFile("my-openapi-spec.json")
```

## Tracking source locations (Origin)

When `IncludeOrigin` is enabled, the loader records the file, line, and column of each element in the OpenAPI document. This is useful for tools that need to report errors or changes with precise source locations (e.g. linters, diff tools, editors).

```go
loader := openapi3.NewLoader()
loader.IncludeOrigin = true
doc, err := loader.LoadFromFile("my-openapi-spec.json")

// Each element has an Origin field with source location info
fmt.Println(doc.Info.Origin.Key.File)   // "my-openapi-spec.json"
fmt.Println(doc.Info.Origin.Key.Line)   // 2
fmt.Println(doc.Info.Origin.Key.Column) // 1
```

The `Origin` struct contains three parts:
- **`Key`** — the location of the object itself (file, line, column).
- **`Fields`** — locations of scalar fields within the object (e.g. `origin.Fields["description"]` gives the line of the `description` field).
- **`Sequences`** — locations of items in sequence-valued fields. For example, `origin.Sequences["enum"]` gives the location of each item in an `enum` array. This is used for fields like `enum`, `required`, and `servers` where the individual items are scalars and don't have their own `Origin` field.

Origin data is populated by an internal post-processing step after YAML decoding — it is not part of the OpenAPI spec itself. For this reason, Origin fields are excluded from serialization. If you marshal a loaded document back to JSON/YAML, origin data will not appear in the output.

## Identifying validation errors by code

Each validation error carries a stable, kebab-case code (e.g. `operation-responses-required`), independent of the message text, so tools can suppress specific findings, assign per-rule severities, or emit machine-readable diagnostics. The full catalog is available from `openapi3.ValidationErrorCodes()`.

```go
err := doc.Validate(ctx, openapi3.EnableMultiError())
for _, e := range err.(openapi3.MultiError) {
	var coded openapi3.CodedError
	if errors.As(e, &coded) {
		fmt.Println(coded.Code(), e) // e.g. "operation-responses-required value of responses must be an object"
	}
}
```

## Getting OpenAPI operation that matches request
```go
loader := openapi3.NewLoader()
doc, _ := loader.LoadFromData([]byte(`...`))
_ = doc.Validate(loader.Context)
router, _ := gorillamux.NewRouter(doc)
route, pathParams, _ := router.FindRoute(httpRequest)
// Do something with route.Operation
```

## Validating HTTP requests/responses
```go
package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

func main() {
	ctx := context.Background()
	loader := &openapi3.Loader{Context: ctx, IsExternalRefsAllowed: true}
	doc, _ := loader.LoadFromFile(".../My-OpenAPIv3-API.yml")
	// Validate document
	_ = doc.Validate(ctx)
	router, _ := gorillamux.NewRouter(doc)
	httpReq, _ := http.NewRequest(http.MethodGet, "/items", nil)

	// Find route
	route, pathParams, _ := router.FindRoute(httpReq)

	// Validate request
	requestValidationInput := &openapi3filter.RequestValidationInput{
		Request:    httpReq,
		PathParams: pathParams,
		Route:      route,
	}
	_ = openapi3filter.ValidateRequest(ctx, requestValidationInput)

	// Handle that request
	// --> YOUR CODE GOES HERE <--
	responseHeaders := http.Header{"Content-Type": []string{"application/json"}}
	responseCode := 200
	responseBody := []byte(`{}`)

	// Validate response
	responseValidationInput := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: requestValidationInput,
		Status:                 responseCode,
		Header:                 responseHeaders,
	}
	responseValidationInput.SetBodyBytes(responseBody)
	_ = openapi3filter.ValidateResponse(ctx, responseValidationInput)
}
```

## Custom content type for body of HTTP request/response

By default, the library parses a body of the HTTP request and response of [a few content types](https://github.com/getkin/kin-openapi/blob/6da871e0e170b7637eb568c265c08bc2b5d6e7a3/openapi3filter/req_resp_decoder.go#L1264) e.g. `"text/plain"` or `"application/json"`.
To support other content types you must register decoders for them:

```go
func main() {
	// ...

	// Register a body's decoder for content type "application/xml".
	openapi3filter.RegisterBodyDecoder("application/xml", xmlBodyDecoder)

	// Now you can validate HTTP request that contains a body with content type "application/xml".
	requestValidationInput := &openapi3filter.RequestValidationInput{
		Request:    httpReq,
		PathParams: pathParams,
		Route:      route,
	}
	if err := openapi3filter.ValidateRequest(ctx, requestValidationInput); err != nil {
		panic(err)
	}

	// ...

	// And you can validate HTTP response that contains a body with content type "application/xml".
	if err := openapi3filter.ValidateResponse(ctx, responseValidationInput); err != nil {
		panic(err)
	}
}

func xmlBodyDecoder(body io.Reader, h http.Header, schema *openapi3.SchemaRef, encFn openapi3filter.EncodingFn) (decoded any, err error) {
	// Decode body to a primitive, []any, or map[string]any.
}
```

## Custom function to check uniqueness of array items

By default, the library checks unique items using the following predefined function:

```go
func isSliceOfUniqueItems(xs []any) bool {
	s := len(xs)
	m := make(map[string]struct{}, s)
	for _, x := range xs {
		key, _ := json.Marshal(&x)
		m[string(key)] = struct{}{}
	}
	return s == len(m)
}
```

In the predefined function `json.Marshal` is used to generate a string that can
be used as a map key which is to check the uniqueness of an array
when the array items are objects or arrays. You can register
you own function according to your input data to get better performance:

```go
func main() {
	// ...

	// Register a customized function used to check uniqueness of array.
	openapi3.RegisterArrayUniqueItemsChecker(arrayUniqueItemsChecker)

	// ... other validate codes
}

func arrayUniqueItemsChecker(items []any) bool {
	// Check the uniqueness of the input slice
}
```

## Custom function to change schema error messages

By default, the error message returned when validating a value includes the error reason, the schema, and the input value.

For example, given the following schema:

```json
{
  "type": "string",
  "allOf": [
    { "pattern": "[A-Z]" },
    { "pattern": "[a-z]" },
    { "pattern": "[0-9]" },
    { "pattern": "[!@#$%^&*()_+=-?~]" }
  ]
}
```

Passing the input value `"secret"` to this schema will produce the following error message:

```
string doesn't match the regular expression "[A-Z]"
Schema:
  {
    "pattern": "[A-Z]"
  }

Value:
  "secret"
```

Including the original value in the error message can be helpful for debugging, but it may not be appropriate for sensitive information such as secrets.

To disable the extra details in the schema error message, pass the `openapi3.DisableSchemaErrorDetails()` option:

```go
	// Disable schema error detailed error messages
	err := schema.VisitJSON(value, openapi3.DisableSchemaErrorDetails())

	// or, with openapi3filter
	options := &openapi3filter.Options{
		SchemaValidationOptions: []openapi3.SchemaValidationOption{openapi3.DisableSchemaErrorDetails()},
	}
```

This will shorten the error message to present only the reason:

```
string doesn't match the regular expression "[A-Z]"
```

This option cannot be used with `openapi3.EnableJSONSchema2020()`, which `openapi3filter` enables for OpenAPI 3.1+ documents: the JSON Schema 2020-12 validator includes values in its error messages, so validation then fails with `openapi3.ErrSchemaErrorDetailsWithJSONSchema2020`.

For more fine-grained control over the error message, you can pass a custom `openapi3filter.Options` object to `openapi3filter.RequestValidationInput` that includes a `openapi3filter.CustomSchemaErrorFunc`.

```go
func validationOptions() *openapi3filter.Options {
	options := &openapi3filter.Options{}
	options.WithCustomSchemaErrorFunc(safeErrorMessage)
	return options
}

func safeErrorMessage(err *openapi3.SchemaError) string {
	return err.Reason
}
```

This will change the schema validation errors to return only the `Reason` field, which is guaranteed to not include the original value.

## Reconciling component $ref types

`ReferencesComponentInRootDocument` is a useful helper function to check if a component reference
coincides with a reference in the root document's component objects fixed fields.

This can be used to determine if two schema definitions are of the same structure, helpful for
code generation tools when generating go type models.

```go
doc, err = loader.LoadFromFile("openapi.yml")

for _, path := range doc.Paths.InMatchingOrder() {
	pathItem := doc.Paths.Find(path)

	if pathItem.Get == nil || pathItem.Get.Responses.Status(200) {
		continue
	}

	for _, s := range pathItem.Get.Responses.Status(200).Value.Content {
		name, match := ReferencesComponentInRootDocument(doc, s.Schema)
		fmt.Println(path, match, name) // /record true #/components/schemas/BookRecord
	}
}
```

## CHANGELOG: Sub-v1 breaking API changes

### v0.152.0
* JSON keys of OpenAPI documents only fill the field of the same case (`json:",case:strict"`) when `encoding/json` is implemented with `encoding/json/v2` (by default from Go 1.27, and on Go 1.25 and 1.26 with `GOEXPERIMENT=jsonv2`): a key such as `"Components"` (not `"components"`) is kept as an extension instead of filling `openapi3.T.Components`. Other builds still match keys case-insensitively, as `encoding/json` v1 does.
* On Go 1.27+, unless built with `GOEXPERIMENT=nojsonv2`, `openapi2` and `openapi3` decode documents with `encoding/json/v2`: objects with duplicate member names and strings with invalid UTF-8 are rejected, and decoding errors are worded differently and locate the faulty value from the root of the document, keys of paths, responses and callbacks included (e.g. `.paths.~1pets.get.responses.200.description`). On these builds `openapi2.Responses`, `openapi3.Callback`, `openapi3.Paths` and `openapi3.Responses` also implement `encoding/json/v2.UnmarshalerFrom`.
* A schema with an empty `type: []` no longer marshals as `"type": null` which did not survive being read back.
* The error for a header with neither or both of `content` and `schema` reads `header must contain exactly one of content and schema`. It used to say `parameter ...` followed by a `%v` dump of the `Header` struct, which printed heap addresses: the message differed between two runs over the same document, and leaked addresses into anything that logged it.
* `openapi3filter` now parses `integer` values of parameters and `application/x-www-form-urlencoded`/`multipart/form-data` bodies as base 10 only, so `010` decodes to `10` rather than `8`, and Go integer literals such as `0x10`, `0b10`, `0o10` or `1_000` are now rejected as invalid integers.
* `openapi3filter.ValidateResponse` now decodes and validates response headers described with `content` rather than `schema`, with the body decoder registered for the header's media type (see `RegisterBodyDecoder`). Such a header is no longer reported missing when it is required and present, and a value that does not match its schema is now rejected. As with a response body, a media type with no registered decoder fails with `unsupported content type`.
* `Validate` no longer reports `extra sibling fields` for fields written next to a `$ref`: every version's Reference Object says they "SHALL be ignored". From OpenAPI 3.1 on, `summary` and `description` still override the referenced component's. `ExtraSiblingFieldsError` is still returned for unknown fields of inline objects, and for `x-` extensions next to a `$ref` under `ProhibitExtensionsWithRef`, which now applies to OpenAPI 3.1+ schemas too. It is no longer returned after `InternalizeRefs` for the fields that were next to a component's `$ref` it replaced, and which it leaves out.
* In OpenAPI 3.1+ documents, the keywords next to a schema's `$ref` now apply alongside the referenced schema, as JSON Schema 2020-12 has it, instead of replacing its keywords of the same name: `{"$ref": ..., "maxLength": 10}` next to a referenced `maxLength: 3` still rejects `"abcd"`, and sibling `required` or `properties` add to the referenced ones. The resolved `SchemaRef.Value` keeps the referenced schema's shape: annotations such as `description`, `readOnly` or `default` replace the referenced ones there, and the other sibling keywords are appended to its `AllOf`.
* Marshaling an `openapi3.*Ref` with a `$ref` now writes back the fields that were written next to it, instead of emitting a lone `{"$ref": ...}`. An OpenAPI 3.1 schema's sibling keywords, which constrain the referenced schema, were lost that way, and so was every other field next to a `$ref`.
* The loader now resolves a Path Item Object's `$ref` even when fields are written next to it, and merges the two: it used to skip the `$ref` then, silently dropping every operation of the referenced path item and never noticing a `$ref` to nothing. The specification leaves the result undefined only for a field set on both sides, which is then taken from next to the `$ref`. Marshaling writes back the `$ref` with the fields that were next to it, which it used to drop.
* The loader no longer panics with `interface conversion` when the same `$ref` is written where two kinds of objects are expected, one nested in the other, such as a path item and one of its responses.
* `InternalizeRefs` now makes a component of the schema an external `$ref` points to, rather than of that schema with the OpenAPI 3.1 keywords next to the `$ref` applied to it, which leaked them into every other use of the component. The external `$ref`s within those keywords are now rewritten in the marshaled document too.
* The keywords next to a schema's `$ref` in a file of bare schemas, which has no `openapi` field, now follow the version of the document that refers to the file: they were ignored, as in OpenAPI 3.0, even from an OpenAPI 3.1 document.
* In OpenAPI 3.1+ documents, loading now fails when the keywords next to a schema's `$ref` do not form a schema, such as `allOf: {type: integer}`, instead of silently ignoring them all.
* The errors raised by the keywords next to an OpenAPI 3.1 schema's `$ref` now carry their `Origin`.

### v0.151.0
* Removed the deprecated package-global `openapi3.IncludeOrigin` variable. Set the `IncludeOrigin` field of `openapi3.Loader` instead.
* `openapi3filter.UrlencodedBodyDecoder` (hence `application/x-www-form-urlencoded` request body validation) now returns a `*openapi3filter.ParseError` for a value that cannot be parsed into its property's schema type, such as `n=abc` for an `integer` property, instead of silently dropping it from the decoded body.
* `openapi3filter` now writes a query parameter's array `default` into the request URL following the parameter's `style` and `explode` (including their defaults), so an array without `explode` set is written as `?ids=1&ids=2` rather than `?ids=1,2`, and `pipeDelimited`/`spaceDelimited` arrays use their own delimiter.

### v0.150.0
* `openapi2.Operation.Responses` field type changed from `map[string]*Response` to `*Responses`, which keeps the Responses Object's `x-` extensions in `Responses.Extensions` instead of failing to decode them as responses.
* `(*openapi3.Schema).VisitJSON(..)` without the `openapi3.EnableJSONSchema2020()` option no longer validates the OpenAPI 3.1 `const` and `prefixItems` keywords. Like the other 3.1-only keywords, they are validated by the JSON Schema 2020-12 validator, which `openapi3filter` enables for 3.1 documents. `items` still applies only past the `prefixItems` positions.
* OpenAPI 3.0 validation (`(*openapi3.T).Validate(..)`, `(*openapi3.Schema).Validate(..)` without `openapi3.IsOpenAPI31OrLater()`) now rejects a `type` with more than one entry (`openapi3.TypeArrayFor31Plus`) and a numeric `exclusiveMinimum`/`exclusiveMaximum` (`openapi3.ExclusiveBoundNumberFor31Plus`), as those forms only exist since OpenAPI 3.1.
* OpenAPI 3.1+ validation now rejects a boolean `exclusiveMinimum`/`exclusiveMaximum` (`openapi3.ExclusiveBoundBooleanBefore31`), as JSON Schema 2020-12 requires a number. `openapi3.FieldVersionMismatchError` gained a `RemovedIn` field, set instead of `MinVersion` for such forms.
* Since OpenAPI 3.1 `nullable` is an unknown keyword: it is still accepted by `Validate` and still decoded into `openapi3.Schema.Nullable`, but `(*openapi3.Schema).VisitJSON(..)` with `openapi3.EnableJSONSchema2020()` (hence `openapi3filter` on 3.1+ documents) no longer lets it allow `null`. Use a `type` array including `"null"` instead. Without that option `nullable` keeps its OpenAPI 3.0 meaning.
* `(*openapi3.MediaType).Validate(..)` now validates its `encoding` entries, so `(*openapi3.T).Validate(..)` now rejects an Encoding Object with an invalid `style`/`explode` combination, or whose `contentType` is not a comma-separated list of media ranges (`openapi3.InvalidEncodingContentTypeError`). Errors are wrapped in `openapi3.MediaTypeEncodingValidationError`.
* `(openapi3.Content).Validate(..)` now rejects keys that match the same media type, compared case-insensitively (RFC 9110, section 8.3.1): e.g. `application/json` and `Application/JSON`, or `a/b` and `c/d, a/b` (`openapi3.ContentKeysOverlapError`). Keys that differ only by specificity (`application/json;charset=utf-8`, `application/json`, `application/*`, `*/*`) are still allowed.
* `(openapi3.Content).Get(..)` now matches media types case-insensitively and matches any entry of a key that is a comma-separated list.
* `openapi3filter` matches `Content-Type` headers against registered body decoders/encoders and `encoding.contentType` case-insensitively, so e.g. `Application/JSON` bodies are now decoded. The media type it returns or reports in errors is lowercased.
* `(*openapi3.Schema).VisitJSON(..)` with `openapi3.EnableJSONSchema2020()` no longer rewrites OpenAPI 3.0 forms (boolean `exclusiveMinimum`/`exclusiveMaximum`) or strips OpenAPI-only keywords before handing the schema to the JSON Schema 2020-12 validator. A schema must pass `Validate` before being used with `VisitJSON`.
* `openapi3filter.JSONBodyDecoder` (hence request and response body validation) now rejects objects with duplicate member names, such as `{"a":1,"a":2}`, on Go 1.27+ unless built with `GOEXPERIMENT=nojsonv2`. To keep accepting them, register your own decoder with `openapi3filter.RegisterBodyDecoder`.
* `(*openapi3.Parameter).SerializationMethod()` (for query parameters) and `(*openapi3.Encoding).SerializationMethod()` now default `explode` to `false` for `style: spaceDelimited` and `style: pipeDelimited`, as the spec says, so `openapi3filter` splits `?ids=1|2|3` and `application/x-www-form-urlencoded` body properties such as `ids=1|2|3` instead of expecting `ids=1&ids=2&ids=3`. Set `explode: true` to keep the previous behavior.
* Removed the package-global `openapi3.SchemaErrorDetailsDisabled` variable. Pass the new `openapi3.DisableSchemaErrorDetails()` option to `(*openapi3.Schema).VisitJSON(..)` instead, or add it to `openapi3filter.Options.SchemaValidationOptions`, which `openapi3filter` now also applies when validating parameters. Combined with `openapi3.EnableJSONSchema2020()` (hence with `openapi3filter` on OpenAPI 3.1+ documents), validation fails with `openapi3.ErrSchemaErrorDetailsWithJSONSchema2020`, as that validator's error messages include values.
* The message customizer set with `openapi3.SetSchemaErrorMessageCustomizer(..)` (hence `(*openapi3filter.Options).WithCustomSchemaErrorFunc(..)`) now also applies to discriminator errors, to errors of patterns that fail to compile and to errors of the JSON Schema 2020-12 validator (`openapi3.EnableJSONSchema2020()`).

### v0.147.0
* `(*openapi3.PathItem).SetOperation(string, *Operation)` no longer panics on unhandled HTTP methods: these are now stored in the new `openapi3.PathItem.AdditionalOperations` field (passing a nil operation deletes the entry).
* `(*openapi3.PathItem).GetOperation(string)` and `(*openapi3.PathItem).Operations()` now also report the new `openapi3.PathItem.Query` field (the OpenAPI 3.2 HTTP `QUERY` method) and the `openapi3.PathItem.AdditionalOperations` entries. Routers and validators consequently match these methods.
* `openapi2conv.FromV3(..)` and `openapi2conv.FromV3PathItem(..)` now return an error for operations whose HTTP method Swagger 2.0 cannot express (`QUERY`, `TRACE`, `CONNECT` and custom methods) instead of panic-ing.

### v0.145.0
* `(*openapi3.PathItem).GetOperation(string)` now returns `nil` on unhandled HTTP methods instead of panic-ing.
* `(*openapi2.PathItem).GetOperation(string)` now returns `nil` on unhandled HTTP methods instead of panic-ing.

### v0.144.0
* `openapi3filter.ValidationHandler.AuthenticationFunc` no longer defaults to `NoopAuthenticationFunc`. Users now have to explicitly set the field and implement their own `AuthenticationFunc` or use the noop impl should they wish to.

### v0.143.0
* Removed the `openapi3.StringMap[V]` type (an internal helper for origin-aware map unmarshalling, obsolete since origin tracking moved to a separate `OriginTree` pass). `openapi3.Discriminator.Mapping` field type changed from `StringMap[MappingRef]` to `map[string]MappingRef`, and `openapi3.OAuthFlow.Scopes` from `StringMap[string]` to `map[string]string`.

### v0.137.0
* Reinstated `openapi3.*Ptr(..)` funcs for Go 1.25

### v0.136.0
* `openapi3.Schema.ExclusiveMin` and `openapi3.Schema.ExclusiveMax` fields changed from `bool` to `ExclusiveBound` (a union type holding `*bool` for OpenAPI 3.0 or `*float64` for OpenAPI 3.1).
* `openapi3.Schema.PrefixItems` field changed from `[]*SchemaRef` to `SchemaRefs`.
* `openapi3.Schema.UnevaluatedItems` and `openapi3.Schema.UnevaluatedProperties` fields changed from `*SchemaRef` to `BoolSchema` (a union type accepting either a boolean or a schema object).
* Removed `openapi3.*Ptr(..)` funcs: they all can be replaced with `new(..)` since Go 1.26

### v0.135.0
* `openapi3.MediaType.Encoding` field type changed from `map[string]*Encoding` to `Encodings`
* `openapi3.Server.Variables` field type changed from `map[string]*ServerVariable` to `ServerVariables`

### v0.134.0
* `openapi3.Location` gained `File` and `Name` fields (`string` type, replacing previous `int`-only struct layout)
* `openapi3.Origin` gained `Sequences` field (`map[string][]Location`, extending previous `map[string]Location`-only struct)

### v0.131.0
* No longer `openapi3filter.RegisterBodyDecoder` the `openapi3filter.ZipFileBodyDecoder` by default.

### v0.129.0
* `openapi3.Discriminator.Mapping` and `openapi3.OAuthFlow.Scopes` fields went from a `map[string]string` to the new type `StringMap`

### v0.127.0
* Downgraded `github.com/gorilla/mux` dep from `1.8.1` to `1.8.0`.

### v0.126.0
* `openapi3.CircularReferenceError` and `openapi3.CircularReferenceCounter` are removed. `openapi3.Loader` now implements reference backtracking, so any kind of circular references should be properly resolved.
* `InternalizeRefs` now takes a refNameResolver that has access to `openapi3.T` and more properties of the reference needing resolving.
* The `DefaultRefNameResolver` has been updated, choosing names that will be less likely to collide with each other. Because of this internalized specs will likely change slightly.
* `openapi3.Format` and `openapi3.FormatCallback` are removed and the type of `openapi3.SchemaStringFormats` has changed.

### v0.125.0
* The `openapi3filter.ErrFunc` and `openapi3filter.LogFunc` func types now take the validated request's context as first argument.

### v0.124.0
* `openapi3.Schema.Type` & `openapi2.Parameter.Type` fields went from a `string` to the type `*Type` with methods: `Includes`, `Is`, `Permits` & `Slice`.

### v0.122.0
* `Paths` field of `openapi3.T` is now a pointer
* `Responses` field of `openapi3.Operation` is now a pointer
* `openapi3.Paths` went from `map[string]*PathItem` to a struct with an `Extensions` field and methods: `Set`, `Value`, `Len`, `Map`, and `New*`.
* `openapi3.Callback` went from `map[string]*PathItem` to a struct with an `Extensions` field and methods: `Set`, `Value`, `Len`, `Map`, and `New*`.
* `openapi3.Responses` went from `map[string]*ResponseRef` to a struct with an `Extensions` field and methods: `Set`, `Value`, `Len`, `Map`, and `New*`.
* `(openapi3.Responses).Get(int)` renamed to `(*openapi3.Responses).Status(int)`

### v0.121.0
* Introduce `openapi3.RequestBodies` (an alias on `map[string]*openapi3.ResponseRef`) and use it in place of `openapi3.Responses` for field `openapi3.Components.Responses`.

### v0.116.0
* Dropped `openapi3filter.DefaultOptions`. Use `&openapi3filter.Options{}` directly instead.

### v0.113.0
* The string format `email` has been removed by default. To use it please call `openapi3.DefineStringFormat("email", openapi3.FormatOfStringForEmail)`.
* Field `openapi3.T.Components` is now a pointer.
* Fields `openapi3.Schema.AdditionalProperties` and `openapi3.Schema.AdditionalPropertiesAllowed` are replaced by `openapi3.Schema.AdditionalProperties.Schema` and `openapi3.Schema.AdditionalProperties.Has` respectively.
* Type `openapi3.ExtensionProps` is now just `map[string]any` and extensions are accessible through the `Extensions` field.

### v0.112.0
* `(openapi3.ValidationOptions).ExamplesValidationDisabled` has been unexported.
* `(openapi3.ValidationOptions).SchemaFormatValidationEnabled` has been unexported.
* `(openapi3.ValidationOptions).SchemaPatternValidationDisabled` has been unexported.

### v0.111.0
* Changed `func (*_) Validate(ctx context.Context) error` to `func (*_) Validate(ctx context.Context, opts ...ValidationOption) error`.
* `openapi3.WithValidationOptions(ctx context.Context, opts *ValidationOptions) context.Context` prototype changed to `openapi3.WithValidationOptions(ctx context.Context, opts ...ValidationOption) context.Context`.

### v0.101.0
* `openapi3.SchemaFormatValidationDisabled` has been removed in favour of an option `openapi3.EnableSchemaFormatValidation()` passed to `openapi3.T.Validate`. The default behaviour is also now to not validate formats, as the OpenAPI spec mentions the `format` is an open value.

### v0.84.0
* The prototype of `openapi3gen.NewSchemaRefForValue` changed:
	* It no longer returns a map but that is still accessible under the field `(*Generator).SchemaRefs`.
	* It now takes in an additional argument (basically `doc.Components.Schemas`) which gets written to so `$ref` cycles can be properly handled.

### v0.61.0
* Renamed `openapi2.Swagger` to `openapi2.T`.
* Renamed `openapi2conv.FromV3Swagger` to `openapi2conv.FromV3`.
* Renamed `openapi2conv.ToV3Swagger` to `openapi2conv.ToV3`.
* Renamed `openapi3.LoadSwaggerFromData` to `openapi3.LoadFromData`.
* Renamed `openapi3.LoadSwaggerFromDataWithPath` to `openapi3.LoadFromDataWithPath`.
* Renamed `openapi3.LoadSwaggerFromFile` to `openapi3.LoadFromFile`.
* Renamed `openapi3.LoadSwaggerFromURI` to `openapi3.LoadFromURI`.
* Renamed `openapi3.NewSwaggerLoader` to `openapi3.NewLoader`.
* Renamed `openapi3.Swagger` to `openapi3.T`.
* Renamed `openapi3.SwaggerLoader` to `openapi3.Loader`.
* Renamed `openapi3filter.ValidationHandler.SwaggerFile` to `openapi3filter.ValidationHandler.File`.
* Renamed `routers.Route.Swagger` to `routers.Route.Spec`.

### v0.51.0
* Type `openapi3filter.Route` moved to `routers` (and `Route.Handler` was dropped. See https://github.com/getkin/kin-openapi/issues/329)
* Type `openapi3filter.RouteError` moved to `routers` (so did `ErrPathNotFound` and `ErrMethodNotAllowed` which are now `RouteError`s)
* Routers' `FindRoute(...)` method now takes only one argument: `*http.Request`
* `getkin/kin-openapi/openapi3filter.Router` moved to `getkin/kin-openapi/routers/legacy`
* `openapi3filter.NewRouter()` and its related `WithSwaggerFromFile(string)`, `WithSwagger(*openapi3.Swagger)`, `AddSwaggerFromFile(string)` and `AddSwagger(*openapi3.Swagger)` are all replaced with a single `<router package>.NewRouter(*openapi3.Swagger)`
	* NOTE: the `NewRouter(doc)` call now requires that the user ensures `doc` is valid (`doc.Validate() != nil`). This used to be asserted.

### v0.47.0
Field `(*openapi3.SwaggerLoader).LoadSwaggerFromURIFunc` of type `func(*openapi3.SwaggerLoader, *url.URL) (*openapi3.Swagger, error)` was removed after the addition of the field `(*openapi3.SwaggerLoader).ReadFromURIFunc` of type `func(*openapi3.SwaggerLoader, *url.URL) ([]byte, error)`.
