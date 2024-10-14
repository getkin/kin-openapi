[![CI](https://github.com/getkin/kin-openapi/workflows/go/badge.svg)](https://github.com/getkin/kin-openapi/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/getkin/kin-openapi)](https://goreportcard.com/report/github.com/getkin/kin-openapi)
[![GoDoc](https://godoc.org/github.com/getkin/kin-openapi?status.svg)](https://godoc.org/github.com/getkin/kin-openapi)
[![Join Gitter Chat Channel -](https://badges.gitter.im/getkin/kin.svg)](https://gitter.im/getkin/kin?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

# Introduction
A [Go](https://golang.org) project for handling [OpenAPI](https://www.openapis.org/) files. We target:
* [OpenAPI `v2.0`](https://github.com/OAI/OpenAPI-Specification/blob/main/versions/2.0.md) (formerly known as Swagger)
* [OpenAPI `v3.0`](https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.3.md)
* [OpenAPI `v3.1`](https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.1.0.md) Soon! [Tracking issue here.](https://github.com/getkin/kin-openapi/issues/230)

Licensed under the [MIT License](./LICENSE).

## Contributors, users and sponsors
The project has received pull requests [from many people](https://github.com/getkin/kin-openapi/graphs/contributors). Thanks to everyone!

Please, [give back to this project](https://github.com/sponsors/fenollp) by becoming a sponsor.

Here's some projects that depend on _kin-openapi_:
  * [github.com/go-fuego/fuego](https://github.com/go-fuego/fuego) - "Framework generating OpenAPI 3 spec from source code"
  * [github.com/a-h/rest](https://github.com/a-h/rest) - "Generate OpenAPI 3.0 specifications from Go code without annotations or magic comments"
  * [github.com/Tufin/oasdiff](https://github.com/Tufin/oasdiff) - "A diff tool for OpenAPI Specification 3"
  * [github.com/danielgtaylor/apisprout](https://github.com/danielgtaylor/apisprout) - "Lightweight, blazing fast, cross-platform OpenAPI 3 mock server with validation"
  * [github.com/deepmap/oapi-codegen](https://github.com/deepmap/oapi-codegen) - "Generate Go client and server boilerplate from OpenAPI 3 specifications"
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
* [go-openapi](https://github.com/go-openapi)'s [spec3](https://github.com/go-openapi/spec3)
	* an iteration on [spec](https://github.com/go-openapi/spec) (for OpenAPIv2)
	* see [README](https://github.com/go-openapi/spec3/tree/3fab9faa9094e06ebd19ded7ea96d156c2283dca#oai-object-model---) for the missing parts

Be sure to check [OpenAPI Initiative](https://github.com/OAI)'s [great tooling list](https://github.com/OAI/OpenAPI-Specification/blob/master/IMPLEMENTATIONS.md) as well as [OpenAPI.Tools](https://openapi.tools/).

# Structure
  * _openapi2_ ([godoc](https://godoc.org/github.com/getkin/kin-openapi/openapi2))
    * Support for OpenAPI 2 files, including serialization, deserialization, and validation.
  * _openapi2conv_ ([godoc](https://godoc.org/github.com/getkin/kin-openapi/openapi2conv))
    * Converts OpenAPI 2 files into OpenAPI 3 files.
  * _openapi3_ ([godoc](https://godoc.org/github.com/getkin/kin-openapi/openapi3))
    * Support for OpenAPI 3 files, including serialization, deserialization, and validation.
  * _openapi3filter_ ([godoc](https://godoc.org/github.com/getkin/kin-openapi/openapi3filter))
    * Validates HTTP requests and responses
    * Provides a [gorilla/mux](https://github.com/gorilla/mux) router for OpenAPI operations
  * _openapi3gen_ ([godoc](https://godoc.org/github.com/getkin/kin-openapi/openapi3gen))
    * Generates `*openapi3.Schema` values for Go types.

# Some recipes
## Validating an OpenAPI document
```shell
go run github.com/getkin/kin-openapi/cmd/validate@latest [--circular] [--defaults] [--examples] [--ext] [--patterns] -- <local YAML or JSON file>
```

## Loading OpenAPI document
Use `openapi3.Loader`, which resolves all references:
```go
loader := openapi3.NewLoader()
doc, err := loader.LoadFromFile("my-openapi-spec.json")
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

By default, the library parses a body of the HTTP request and response
if it has one of the following content types: `"text/plain"` or `"application/json"`.
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

To disable the extra details in the schema error message, you can set the `openapi3.SchemaErrorDetailsDisabled` option to `true`:

```go
func main() {
	// ...

	// Disable schema error detailed error messages
	openapi3.SchemaErrorDetailsDisabled = true

	// ... other validate codes
}
```

This will shorten the error message to present only the reason:

```
string doesn't match the regular expression "[A-Z]"
```

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
