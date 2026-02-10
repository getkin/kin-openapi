# Open Issues — OpenAPI 3.1 Support

## Missing Features

### 1. `$schema` keyword not added to Schema type
The JSON Schema 2020-12 `$schema` keyword allows individual schemas to declare their dialect. It is not present on the `Schema` struct. The document-level `jsonSchemaDialect` field exists on the root `T` struct, but per-schema `$schema` is not supported.

### 2. `$dynamicRef` / `$dynamicAnchor` not resolved
`$id`, `$anchor`, `$dynamicRef`, and `$dynamicAnchor` are parsed and serialized but the loader does not actually resolve them. Schemas that rely on `$dynamicRef` for recursive schema references will not work correctly.

### 3. Built-in validator ignores most 3.1 keywords
The built-in validator (used when `EnableJSONSchema2020()` is NOT set) only validates `const` from the new keywords. All other 3.1 keywords are silently ignored:
- `if` / `then` / `else`
- `dependentRequired`
- `prefixItems`
- `contains` / `minContains` / `maxContains`
- `patternProperties`
- `dependentSchemas`
- `propertyNames`
- `unevaluatedItems` / `unevaluatedProperties`

Users must call `EnableJSONSchema2020()` to get validation for these keywords (which delegates to the external `santhosh-tekuri/jsonschema/v6` library).

### 4. `contentMediaType` / `contentEncoding` not validated at runtime
These keywords are parsed and serialized but the built-in validator does not decode or validate content based on them. JSON Schema 2020-12 treats these as annotation-only by default, so this is spec-compliant, but some users may expect validation.

### 5. Missing `$defs` keyword
The `$defs` keyword from JSON Schema 2020-12 core vocabulary is not present on the `Schema` struct. In OpenAPI 3.1, schemas can use `$defs` to define local reusable schemas (an alternative to `#/components/schemas`). Without this field, `$defs` blocks are silently stored in `Extensions`, the loader does not resolve `$ref`s inside them, and round-trip fidelity is affected.

### 6. Missing `$comment` keyword
The `$comment` keyword from JSON Schema 2020-12 core vocabulary is not present on the `Schema` struct. While purely informational with no validation impact, it should be a first-class field for round-trip fidelity. Its absence means `$comment` values end up in `Extensions`.

### 7. Missing `pathItems` in Components struct
Per the OpenAPI 3.1 spec, the Components Object added a new `pathItems` field: `pathItems: Map[string, Path Item Object | Reference Object]`. This field is completely absent from the `Components` struct — not in the struct definition, not in MarshalYAML/UnmarshalJSON, not in Validate, and not in the loader's components resolution. Any 3.1 spec using `components/pathItems` will have those items silently dropped.

### 8. ~~`Schema.JSONLookup()` missing all new 3.1 fields~~ **FIXED**
All new 3.1 fields have been added to `JSONLookup()`: `const`, `examples`, `prefixItems`, `contains`, `minContains`, `maxContains`, `patternProperties`, `dependentSchemas`, `propertyNames`, `unevaluatedItems`, `unevaluatedProperties`, `if`, `then`, `else`, `dependentRequired`, `$id`, `$anchor`, `$dynamicRef`, `$dynamicAnchor`, `contentMediaType`, `contentEncoding`, `contentSchema`.

## Bugs / Correctness

### 9. Silent fallback in `visitJSONWithJSONSchema`
When the external JSON Schema 2020-12 validator fails to compile a schema, `visitJSONWithJSONSchema()` silently falls back to the built-in validator (`schema.go:187`). This can hide errors — a schema that is valid JSON Schema 2020-12 but fails compilation will be validated with the less-capable built-in validator without any warning.

### 10. `Const: nil` cannot express "value must be null"
In `visitConstOperation()`, `schema.Const == nil` is treated as "not set" and skips validation. There is no way to express "the value must be JSON null" using `const` in the built-in validator, since Go `nil` is the zero value for `any`.

### 11. ~~`IsEmpty()` missing checks for most new 3.1 fields~~ **FIXED**
Added checks for `PrefixItems`, `Contains`, `MinContains`, `MaxContains`, `PatternProperties`, `DependentSchemas`, `PropertyNames`, `UnevaluatedItems`, `UnevaluatedProperties`, `Examples`.

### 12. ~~`validate()` requires `items` for type `array` even in 3.1 mode~~ **FIXED**
Relaxed the check: `items` is no longer required when `jsonSchema2020ValidationEnabled` is true or when `prefixItems` is present.

### 13. Built-in validator applies `items` to ALL array elements, ignoring `prefixItems`
In JSON Schema 2020-12, when both `prefixItems` and `items` are present, `items` applies only to elements beyond the `prefixItems` tuple. The built-in `visitJSONArray` unconditionally applies `items` to all elements. Without `prefixItems` awareness, the `items` semantics are wrong for schemas that use both keywords together.

### 14. `patternProperties` omission changes `additionalProperties` semantics
In JSON Schema, a property is "additional" only if it doesn't match any `properties` key OR any `patternProperties` pattern. The built-in `visitJSONObject` has no `patternProperties` support, so `additionalProperties: false` incorrectly rejects properties that should be allowed by pattern matches.

### 15. ~~`transformOpenAPIToJSONSchema` does not clean up `exclusiveMinimum: false`~~ **FIXED**
Both `exclusiveMinimum: false` and `exclusiveMaximum: false` are now deleted during transformation. Also handles the case where `exclusiveMinimum: true` but `minimum` is absent.

### 16. ~~`transformOpenAPIToJSONSchema` drops `nullable: true` without `type`~~ **FIXED**
When `nullable: true` is present without a `type` field, the transformation now sets `type: ["null"]` to preserve the null intent.

### 17. ~~`MarshalYAML` unconditionally emits `"paths": null` for 3.1 docs without paths~~ **FIXED**
`MarshalYAML` now only sets `m["paths"]` when `doc.Paths` is non-nil.

### 18. ~~`visitConstOperation` uses `==` for `json.Number` comparison~~ **FIXED**
Changed to use `reflect.DeepEqual` for consistency with the `int64` and `default` cases.

### 19. `exclusiveBoundToBool` loses constraint data during OAS 3.1 to OAS 2.0 conversion
In `openapi2conv`, when converting a 3.1 schema with `exclusiveMinimum: 5` (no `minimum` field) to OAS 2.0, the conversion produces `exclusiveMinimum: true` with `minimum: nil`. The actual bound value `5` is lost entirely. The correct conversion should set `minimum: 5, exclusiveMinimum: true`.

### 20. `jsonSchemaDialect` URI validation is effectively a no-op
In `openapi3.go:250-254`, `url.Parse()` is used to validate `jsonSchemaDialect` as a URI, but Go's `url.Parse` accepts almost any string. For example, `url.Parse("not a url")` succeeds. The validation provides no meaningful checking.

### 21. ~~Webhooks validation iterates map in non-deterministic order~~ **FIXED**
Changed to use `componentNames()` for deterministic iteration order, consistent with the loader.

### 22. `doc.Validate()` does not auto-enable 3.1 validation mode
When a document has `openapi: "3.1.0"`, users must explicitly pass `EnableJSONSchema2020Validation()` to avoid spurious errors for `type: "null"` and `$ref` siblings. `cmd/validate` auto-detects this, but library users get a poor out-of-box experience for 3.1 documents.

## Breaking API Changes (need README documentation)

### 23. `ExclusiveMin` / `ExclusiveMax` type changed
Changed from `bool` to `ExclusiveBound` (a union type holding `*bool` or `*float64`). Any code that reads `schema.ExclusiveMin` as a `bool` will fail to compile. This needs to be added to the "Sub-v1 breaking API changes" section in README.md.

### 24. `Paths` JSON tag changed
Changed from `json:"paths"` to `json:"paths,omitempty"`. This makes paths optional in serialized output for both 3.0 and 3.1 documents. A 3.0 document with nil `Paths` will now serialize without the `paths` key, which is technically invalid per the 3.0 spec. The `Validate()` function still enforces paths as required for 3.0, but the serialization does not. This needs to be added to the "Sub-v1 breaking API changes" section in README.md.

## Code Quality

### 25. Discriminator logic duplicated for `anyOf`
The discriminator handling in `visitXOFOperations()` was copy-pasted from the `oneOf` path to support `anyOf`. This could be refactored into a shared helper to reduce duplication.

### 26. `PrefixItems` type inconsistency
`PrefixItems` uses `[]*SchemaRef` while the equivalent fields `OneOf`, `AnyOf`, `AllOf` use `SchemaRefs` (which has a `JSONLookup` method for JSON Pointer support). JSON Pointer paths like `#/components/schemas/Foo/prefixItems/0` will not work correctly.

## Test Coverage Gaps

### 27. No ref resolution test for `ContentSchema`
The test file `loader_31_schema_refs_test.go` and testdata `schema31refs.yml` do not include a test case for `$ref` inside `contentSchema`.

### 28. No ref resolution test for `If`/`Then`/`Else`
The dedicated ref-resolution test suite for 3.1 fields (`schema31refs.yml`) does not cover `if`/`then`/`else`. These may be tested elsewhere, but the centralized test is incomplete.

### 29. No `transformOpenAPIToJSONSchema` test for `ContentSchema`
The test `TestJSONSchema2020Validator_TransformRecursesInto31Fields` does not include a sub-test for `contentSchema` with an OAS 3.0-ism like `nullable: true`.

### 30. No `Schema.validate()` test for `ContentSchema`
The test `TestSchemaValidate31SubSchemas` does not include a test for `contentSchema` containing an invalid sub-schema.
