package openapi3filter

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func Test_validateResponseHeader(t *testing.T) {
	t.Parallel()

	type args struct {
		headerName string
		headerRef  *openapi3.HeaderRef
	}
	tests := []struct {
		name            string
		args            args
		isHeaderPresent bool
		headerVals      []string
		wantErr         bool
		wantErrMsg      string
	}{
		{
			name: "test required string header with single string value",
			args: args{
				headerName: "X-Blab",
				headerRef:  newHeaderRef(openapi3.NewStringSchema(), true),
			},
			isHeaderPresent: true,
			headerVals:      []string{"blab"},
			wantErr:         false,
		},
		{
			name: "test required string header with single, empty string value",
			args: args{
				headerName: "X-Blab",
				headerRef:  newHeaderRef(openapi3.NewStringSchema(), true),
			},
			isHeaderPresent: true,
			headerVals:      []string{""},
			wantErr:         false,
		},
		{
			name: "test optional string header with single string value",
			args: args{
				headerName: "X-Blab",
				headerRef:  newHeaderRef(openapi3.NewStringSchema(), false),
			},
			isHeaderPresent: false,
			headerVals:      []string{"blab"},
			wantErr:         false,
		},
		{
			name: "test required, but missing string header",
			args: args{
				headerName: "X-Blab",
				headerRef:  newHeaderRef(openapi3.NewStringSchema(), true),
			},
			isHeaderPresent: false,
			headerVals:      nil,
			wantErr:         true,
			wantErrMsg:      `response header "X-Blab" missing`,
		},
		{
			name: "test integer header with single integer value",
			args: args{
				headerName: "X-Blab",
				headerRef:  newHeaderRef(openapi3.NewIntegerSchema(), true),
			},
			isHeaderPresent: true,
			headerVals:      []string{"88"},
			wantErr:         false,
		},
		{
			name: "test integer header with single string value",
			args: args{
				headerName: "X-Blab",
				headerRef:  newHeaderRef(openapi3.NewIntegerSchema(), true),
			},
			isHeaderPresent: true,
			headerVals:      []string{"blab"},
			wantErr:         true,
			wantErrMsg:      `unable to decode header "X-Blab" value: value blab: an invalid integer: invalid syntax`,
		},
		{
			name: "test int64 header with single int64 value",
			args: args{
				headerName: "X-Blab",
				headerRef:  newHeaderRef(openapi3.NewInt64Schema(), true),
			},
			isHeaderPresent: true,
			headerVals:      []string{"88"},
			wantErr:         false,
		},
		{
			name: "test int32 header with single int32 value",
			args: args{
				headerName: "X-Blab",
				headerRef:  newHeaderRef(openapi3.NewInt32Schema(), true),
			},
			isHeaderPresent: true,
			headerVals:      []string{"88"},
			wantErr:         false,
		},
		{
			name: "test float64 header with single float64 value",
			args: args{
				headerName: "X-Blab",
				headerRef:  newHeaderRef(openapi3.NewFloat64Schema(), true),
			},
			isHeaderPresent: true,
			headerVals:      []string{"88.87"},
			wantErr:         false,
		},
		{
			name: "test integer header with multiple csv integer values",
			args: args{
				headerName: "X-blab",
				headerRef:  newHeaderRef(newArraySchema(openapi3.NewIntegerSchema()), true),
			},
			isHeaderPresent: true,
			headerVals:      []string{"87,88"},
			wantErr:         false,
		},
		{
			name: "test integer header with multiple integer values",
			args: args{
				headerName: "X-blab",
				headerRef:  newHeaderRef(newArraySchema(openapi3.NewIntegerSchema()), true),
			},
			isHeaderPresent: true,
			headerVals:      []string{"87", "88"},
			wantErr:         false,
		},
		{
			name: "test non-typed, nullable header with single string value",
			args: args{
				headerName: "X-blab",
				headerRef:  newHeaderRef(&openapi3.Schema{Nullable: true}, true),
			},
			isHeaderPresent: true,
			headerVals:      []string{"blab"},
			wantErr:         false,
		},
		{
			name: "test required non-typed, nullable header not present",
			args: args{
				headerName: "X-blab",
				headerRef:  newHeaderRef(&openapi3.Schema{Nullable: true}, true),
			},
			isHeaderPresent: false,
			headerVals:      []string{"blab"},
			wantErr:         true,
			wantErrMsg:      `response header "X-blab" missing`,
		},
		{
			name: "test non-typed, non-nullable header with single string value",
			args: args{
				headerName: "X-blab",
				headerRef:  newHeaderRef(&openapi3.Schema{Nullable: false}, true),
			},
			isHeaderPresent: true,
			headerVals:      []string{"blab"},
			wantErr:         true,
			wantErrMsg:      `response header "X-blab" doesn't match schema: Value is not nullable`,
		},
		{
			name: "test required content header with matching value",
			args: args{
				headerName: "x-blab",
				headerRef:  newContentHeaderRef(openapi3.NewObjectSchema().WithProperty("id", openapi3.NewIntegerSchema()), true),
			},
			isHeaderPresent: true,
			headerVals:      []string{`{"id":42}`},
			wantErr:         false,
		},
		{
			name: "test required content header with mismatching value",
			args: args{
				headerName: "X-Blab",
				headerRef:  newContentHeaderRef(openapi3.NewObjectSchema().WithProperty("id", openapi3.NewIntegerSchema()), true),
			},
			isHeaderPresent: true,
			headerVals:      []string{`{"id":"blab"}`},
			wantErr:         true,
			wantErrMsg:      `response header "X-Blab" doesn't match schema: Error at "/id": value must be an integer`,
		},
		{
			name: "test optional content header with mismatching value",
			args: args{
				headerName: "X-Blab",
				headerRef:  newContentHeaderRef(openapi3.NewObjectSchema().WithProperty("id", openapi3.NewIntegerSchema()), false),
			},
			isHeaderPresent: true,
			headerVals:      []string{`{"id":"blab"}`},
			wantErr:         true,
			wantErrMsg:      `response header "X-Blab" doesn't match schema: Error at "/id": value must be an integer`,
		},
		{
			name: "test optional content header not present",
			args: args{
				headerName: "X-Blab",
				headerRef:  newContentHeaderRef(openapi3.NewObjectSchema(), false),
			},
			isHeaderPresent: false,
			wantErr:         false,
		},
		{
			name: "test required, but missing content header",
			args: args{
				headerName: "X-Blab",
				headerRef:  newContentHeaderRef(openapi3.NewObjectSchema(), true),
			},
			isHeaderPresent: false,
			wantErr:         true,
			wantErrMsg:      `response header "X-Blab" missing`,
		},
		{
			name: "test content header with multiple values",
			args: args{
				headerName: "X-Blab",
				headerRef:  newContentHeaderRef(openapi3.NewObjectSchema(), true),
			},
			isHeaderPresent: true,
			headerVals:      []string{`{}`, `{}`},
			wantErr:         true,
			wantErrMsg:      `unable to decode header "X-Blab" value: header "X-Blab" cannot have multiple values`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := newInputDefault()
			opts := []openapi3.SchemaValidationOption(nil)
			if tt.isHeaderPresent {
				input.Header = map[string][]string{http.CanonicalHeaderKey(tt.args.headerName): tt.headerVals}
			}

			err := validateResponseHeader(tt.args.headerName, tt.args.headerRef, input, opts)
			if tt.wantErr {
				require.NotEmpty(t, tt.wantErrMsg, "wanted error message is not populated")
				require.Error(t, err)
				require.ErrorContains(t, err, tt.wantErrMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateResponseContentHeader(t *testing.T) {
	t.Parallel()

	const spec = `
openapi: 3.0.0
info:
  title: 'Validator'
  version: 0.0.1
paths:
  /quota:
    get:
      responses:
        '200':
          description: Ok
          headers:
            X-Quota:
              required: true
              content:
                application/json:
                  schema:
                    type: object
                    required: [remaining]
                    properties:
                      remaining:
                        type: integer
`
	req, err := http.NewRequest(http.MethodGet, "/quota", nil)
	require.NoError(t, err)
	router := setupTestRouter(t, spec)
	route, pathParams, err := router.FindRoute(req)
	require.NoError(t, err)

	validate := func(header http.Header) error {
		return ValidateResponse(t.Context(), &ResponseValidationInput{
			RequestValidationInput: &RequestValidationInput{
				Request:    req,
				PathParams: pathParams,
				Route:      route,
			},
			Status: http.StatusOK,
			Header: header,
			Body:   io.NopCloser(strings.NewReader("")),
		})
	}

	err = validate(http.Header{"X-Quota": {`{"remaining":3}`}})
	require.NoError(t, err)

	err = validate(http.Header{"X-Quota": {`{"remaining":"three"}`}})
	require.ErrorContains(t, err, `response header "X-Quota" doesn't match schema`)

	err = validate(http.Header{})
	require.ErrorContains(t, err, `response header "X-Quota" missing`)
}

func TestValidateResponseContentHeaderMediaType(t *testing.T) {
	t.Parallel()

	newHeaderRef := func(mediaType string) *openapi3.HeaderRef {
		return &openapi3.HeaderRef{Value: &openapi3.Header{Parameter: openapi3.Parameter{
			Content:  openapi3.NewContentWithSchema(openapi3.NewStringSchema().WithMaxLength(3), []string{mediaType}),
			Required: true,
		}}}
	}
	validate := func(mediaType, value string) error {
		input := newInputDefault()
		if value != "" {
			input.Header = http.Header{"X-Blab": {value}}
		}
		return validateResponseHeader("X-Blab", newHeaderRef(mediaType), input, nil)
	}

	err := validate("text/plain", "bla")
	require.NoError(t, err)

	err = validate("text/plain", "blab")
	require.ErrorContains(t, err, `response header "X-Blab" doesn't match schema: maximum string length is 3`)

	err = validate("text/plain", "")
	require.ErrorContains(t, err, `response header "X-Blab" missing`)

	// A +json media type without a decoder of its own is decoded as JSON.
	err = validate("application/vnd.blab+json", `"bla"`)
	require.NoError(t, err)

	err = validate("application/vnd.blab+json", "bla")
	require.ErrorContains(t, err, `unable to decode header "X-Blab" value`)

	err = validate("application/xml", "<blab/>")
	require.ErrorContains(t, err, `unable to decode header "X-Blab" value: unsupported content type "application/xml"`)
}

func newInputDefault() *ResponseValidationInput {
	return &ResponseValidationInput{
		RequestValidationInput: &RequestValidationInput{
			Request:    nil,
			PathParams: nil,
			Route:      nil,
		},
		Status: 200,
		Header: nil,
		Body:   io.NopCloser(strings.NewReader(`{}`)),
	}
}

func newHeaderRef(schema *openapi3.Schema, required bool) *openapi3.HeaderRef {
	return &openapi3.HeaderRef{Value: &openapi3.Header{Parameter: openapi3.Parameter{Schema: &openapi3.SchemaRef{Value: schema}, Required: required}}}
}

func newContentHeaderRef(schema *openapi3.Schema, required bool) *openapi3.HeaderRef {
	return &openapi3.HeaderRef{Value: &openapi3.Header{Parameter: openapi3.Parameter{Content: openapi3.NewContentWithJSONSchema(schema), Required: required}}}
}

func newArraySchema(schema *openapi3.Schema) *openapi3.Schema {
	arraySchema := openapi3.NewArraySchema()
	arraySchema.Items = openapi3.NewSchemaRef("", schema)

	return arraySchema
}
