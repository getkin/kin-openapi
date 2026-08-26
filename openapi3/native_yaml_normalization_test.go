package openapi3

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// Three functions decode a mapping, and each has to apply the same four
// normalizations to what it produces:
//
//	struct  - decodeStructWithExtensions, for objects with declared fields
//	maplike - unmarshalMaplikeYAML, for keyed collections (Responses, Callback)
//	ref     - unmarshalRefYAML, for the $ref wrappers
//
// The normalizations:
//
//	siblings - an undeclared non-x- key is recorded, so Validate can report it
//	numbers  - an integer behind an any is a float64, whatever the notation
//	keys     - a non-string mapping key becomes a string, so the document marshals
//	merge    - a << key is applied, neither ignored nor treated as content
//
// This table covers each function against each normalization. A function that
// leaves one out fails here.
var nativeDecodeTests = []struct {
	path, property string
	spec           string
	assert         func(t *testing.T, doc *T)
}{
	{
		path: "struct", property: "siblings",
		spec: `
openapi: 3.0.0
info: {title: t, version: "1", bogus: true}
paths: {}
`,
		assert: func(t *testing.T, doc *T) { requireValidateFails(t, doc, "bogus") },
	},
	{
		path: "struct", property: "numbers",
		spec: `
openapi: 3.0.0
info: {title: t, version: "1", x-num: 7}
paths: {}
`,
		assert: func(t *testing.T, doc *T) { requireFloat(t, doc.Info.Extensions["x-num"]) },
	},
	{
		path: "struct", property: "keys",
		spec: `
openapi: 3.0.0
info:
  title: t
  version: "1"
  x-thing: {nested: {1: one}}
paths: {}
`,
		assert: requireMarshalable,
	},
	{
		path: "struct", property: "merge",
		spec: `
openapi: 3.0.0
x-defaults: &d
  title: t
  x-num: 7
info:
  <<: *d
  version: "1"
paths: {}
`,
		assert: func(t *testing.T, doc *T) {
			if _, ok := doc.Info.Extensions[mergeKey]; ok {
				t.Errorf("merge key kept as an extension: %v", doc.Info.Extensions)
			}
			if doc.Info.Title != "t" {
				t.Errorf("merged field lost: title = %q", doc.Info.Title)
			}
			if doc.Info.Extensions["x-num"] == nil {
				t.Error("merged extension lost")
			}
			requireValidates(t, doc)
		},
	},
	{
		path: "maplike", property: "numbers",
		spec: `
openapi: 3.0.0
info: {title: t, version: "1"}
paths:
  /a:
    get:
      responses:
        x-num: 7
        "200": {description: ok}
`,
		assert: func(t *testing.T, doc *T) { requireFloat(t, responsesOf(doc).Extensions["x-num"]) },
	},
	{
		path: "maplike", property: "keys",
		spec: `
openapi: 3.0.0
info: {title: t, version: "1"}
paths:
  /a:
    get:
      responses:
        x-thing: {nested: {1: one}}
        "200": {description: ok}
`,
		assert: requireMarshalable,
	},
	{
		path: "maplike", property: "merge",
		spec: `
openapi: 3.0.0
info: {title: t, version: "1"}
x-tmpl: &d
  "200": {description: ok}
paths:
  /a:
    get:
      responses:
        <<: *d
`,
		assert: func(t *testing.T, doc *T) {
			if responsesOf(doc).Value("200") == nil {
				t.Error("merged response lost")
			}
			requireValidates(t, doc)
		},
	},
	{
		path: "ref", property: "siblings",
		spec: `
openapi: 3.0.0
info: {title: t, version: "1"}
paths:
  /a:
    get:
      responses:
        "200": {$ref: '#/components/responses/R', bogus: true}
components:
  responses:
    R: {description: ok}
`,
		assert: func(t *testing.T, doc *T) { requireValidateFails(t, doc, "bogus") },
	},
	{
		path: "ref", property: "numbers",
		spec: `
openapi: 3.0.0
info: {title: t, version: "1"}
paths:
  /a:
    get:
      responses:
        "200": {$ref: '#/components/responses/R', x-num: 7}
components:
  responses:
    R: {description: ok}
`,
		assert: func(t *testing.T, doc *T) {
			requireFloat(t, responsesOf(doc).Map()["200"].Extensions["x-num"])
		},
	},
	{
		path: "ref", property: "keys",
		spec: `
openapi: 3.0.0
info: {title: t, version: "1"}
paths:
  /a:
    get:
      responses:
        "200":
          $ref: '#/components/responses/R'
          x-meta: {1: one}
components:
  responses:
    R: {description: ok}
`,
		assert: requireMarshalable,
	},
	{
		path: "ref", property: "merge",
		spec: `
openapi: 3.0.0
info: {title: t, version: "1"}
x-tmpl: &d
  $ref: '#/components/responses/R'
paths:
  /a:
    get:
      responses:
        "200":
          <<: *d
components:
  responses:
    R: {description: ok}
`,
		assert: func(t *testing.T, doc *T) {
			if got := responsesOf(doc).Map()["200"].Ref; got != "#/components/responses/R" {
				t.Errorf("merged $ref lost: %q", got)
			}
			requireValidates(t, doc)
		},
	},
}

func TestNativeDecodePathsApplyTheSameNormalizations(t *testing.T) {
	for _, c := range nativeDecodeTests {
		t.Run(c.path+"/"+c.property, func(t *testing.T) {
			loader := NewLoader()
			doc, err := loader.LoadFromData([]byte(c.spec))
			if err != nil {
				t.Fatalf("load: %v", err)
			}
			c.assert(t, doc)
		})
	}
}

// notApplicable records the combinations that cannot exist, with the reason.
// A combination that is neither tested nor listed here fails, so adding a decode
// function or a normalization has to come with its tests.
var notApplicable = map[string]string{
	"maplike/siblings": "every non-x- key of a keyed collection is a member of it, so there is no undeclared sibling to record",
}

func TestNativeDecodeCoversEveryNormalization(t *testing.T) {
	tested := map[string]bool{}
	for _, c := range nativeDecodeTests {
		tested[c.path+"/"+c.property] = true
	}
	for _, path := range []string{"struct", "maplike", "ref"} {
		for _, property := range []string{"siblings", "numbers", "keys", "merge"} {
			name := path + "/" + property
			_, exempt := notApplicable[name]
			if !tested[name] && !exempt {
				t.Errorf("%s is untested\n  add it to nativeDecodeTests, or list it in notApplicable with a reason", name)
			}
			if tested[name] && exempt {
				t.Errorf("%s is tested but still listed in notApplicable\n  remove it from notApplicable", name)
			}
		}
	}
}

func responsesOf(doc *T) *Responses {
	return doc.Paths.Value("/a").Get.Responses
}

func requireValidates(t *testing.T, doc *T) {
	t.Helper()
	if err := doc.Validate(context.Background()); err != nil {
		t.Errorf("validate: %v", err)
	}
}

func requireValidateFails(t *testing.T, doc *T, name string) {
	t.Helper()
	err := doc.Validate(context.Background())
	if err == nil {
		t.Errorf("undeclared sibling %q was accepted", name)
		return
	}
	if want := fmt.Sprintf("[%s]", name); !strings.Contains(err.Error(), want) {
		t.Errorf("validate reported %v, want it to name %s", err, want)
	}
}

func requireFloat(t *testing.T, v any) {
	t.Helper()
	if _, ok := v.(float64); !ok {
		t.Errorf("integer behind an any is %T(%v), want float64", v, v)
	}
}

func requireMarshalable(t *testing.T, doc *T) {
	t.Helper()
	if _, err := json.Marshal(doc); err != nil {
		t.Errorf("document loaded but does not marshal: %v", err)
	}
}
