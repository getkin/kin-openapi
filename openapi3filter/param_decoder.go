package openapi3filter

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

const (
	errMsgInvalidValue          = "an invalid value"
	errMsgInvalidSerializationF = "%s parameter %q has an invalid serialization method: style=%q, explode=%v"
)

// ParseError describes errors which happens while parse operation's parameters.
type ParseError struct {
	Value  interface{}
	Path   []interface{}
	Reason string
	Cause  error
}

func (e *ParseError) Error() string {
	var msg []string
	if e.Path != nil {
		msg = append(msg, fmt.Sprintf("failed to parse value %v at path %v", e.Value, e.Path))
	} else {
		msg = append(msg, fmt.Sprintf("failed to parse value %v", e.Value))
	}
	if e.Reason != "" {
		msg = append(msg, e.Reason)
	}
	if e.Cause != nil {
		msg = append(msg, e.Cause.Error())
	}
	return strings.Join(msg, ": ")
}

// decodeParameter returns a value of an operation's parameter from HTTP request.
func decodeParameter(param *openapi3.Parameter, input *RequestValidationInput) (interface{}, error) {
	var decoder interface {
		DecodePrimitive(param *openapi3.Parameter) (interface{}, error)
		DecodeArray(param *openapi3.Parameter) ([]interface{}, error)
		DecodeObject(param *openapi3.Parameter) (map[string]interface{}, error)
	}

	switch param.In {
	case openapi3.ParameterInPath:
		decoder = &pathParamDecoder{input: input}
	case openapi3.ParameterInQuery:
		decoder = &queryParamDecoder{input: input}
	case openapi3.ParameterInHeader:
		decoder = &headerParamDecoder{input: input}
	case openapi3.ParameterInCookie:
		decoder = &cookieParamDecoder{input: input}
	default:
		return nil, fmt.Errorf("unsupported parameter's 'in': %s", param.In)
	}

	switch param.Schema.Value.Type {
	case "array":
		return decoder.DecodeArray(param)
	case "object":
		return decoder.DecodeObject(param)
	default:
		return decoder.DecodePrimitive(param)
	}
}

// pathParamDecoder decodes values of path parameters.
type pathParamDecoder struct {
	input *RequestValidationInput
}

func (d *pathParamDecoder) DecodePrimitive(param *openapi3.Parameter) (interface{}, error) {
	sm, err := param.SerializationMethod()
	if err != nil {
		return nil, err
	}
	var prefix string
	switch sm.Style {
	case "simple":
		// A prefix is empty for style "simple".
	case "label":
		prefix = "."
	case "matrix":
		prefix = ";" + param.Name + "="
	default:
		return nil, fmt.Errorf(errMsgInvalidSerializationF, param.In, param.Name, sm.Style, sm.Explode)
	}

	if d.input.PathParams == nil {
		// HTTP request does not contains a value of the target path parameter.
		return nil, nil
	}
	raw, ok := d.input.PathParams[d.paramKey(param, sm)]
	if !ok || raw == "" {
		// HTTP request does not contains a value of the target path parameter.
		return nil, nil
	}
	src, err := cutPrefix(raw, prefix)
	if err != nil {
		return nil, &RequestError{Input: d.input, Parameter: param, Reason: errMsgInvalidValue, Err: err}
	}
	val, err := parsePrimitive(src, param.Schema)
	if err != nil {
		return nil, &RequestError{Input: d.input, Parameter: param, Reason: errMsgInvalidValue, Err: err}
	}
	return val, nil
}

func (d *pathParamDecoder) DecodeArray(param *openapi3.Parameter) ([]interface{}, error) {
	sm, err := param.SerializationMethod()
	if err != nil {
		return nil, err
	}
	var prefix, delim string
	switch {
	case sm.Style == "simple":
		delim = ","
	case sm.Style == "label" && sm.Explode == false:
		prefix = "."
		delim = ","
	case sm.Style == "label" && sm.Explode == true:
		prefix = "."
		delim = "."
	case sm.Style == "matrix" && sm.Explode == false:
		prefix = ";" + param.Name + "="
		delim = ","
	case sm.Style == "matrix" && sm.Explode == true:
		prefix = ";" + param.Name + "="
		delim = ";" + param.Name + "="
	default:
		return nil, fmt.Errorf(errMsgInvalidSerializationF, param.In, param.Name, sm.Style, sm.Explode)
	}

	if d.input.PathParams == nil {
		// HTTP request does not contains a value of the target path parameter.
		return nil, nil
	}
	raw, ok := d.input.PathParams[d.paramKey(param, sm)]
	if !ok || raw == "" {
		// HTTP request does not contains a value of the target path parameter.
		return nil, nil
	}
	src, err := cutPrefix(raw, prefix)
	if err != nil {
		return nil, &RequestError{Input: d.input, Parameter: param, Reason: errMsgInvalidValue, Err: err}
	}
	val, err := parseArray(strings.Split(src, delim), param.Schema)
	if err != nil {
		return nil, &RequestError{Input: d.input, Parameter: param, Reason: errMsgInvalidValue, Err: err}
	}
	return val, nil
}

func (d *pathParamDecoder) DecodeObject(param *openapi3.Parameter) (map[string]interface{}, error) {
	sm, err := param.SerializationMethod()
	if err != nil {
		return nil, err
	}
	var prefix, propsDelim, valueDelim string
	switch {
	case sm.Style == "simple" && sm.Explode == false:
		propsDelim = ","
		valueDelim = ","
	case sm.Style == "simple" && sm.Explode == true:
		propsDelim = ","
		valueDelim = "="
	case sm.Style == "label" && sm.Explode == false:
		prefix = "."
		propsDelim = ","
		valueDelim = ","
	case sm.Style == "label" && sm.Explode == true:
		prefix = "."
		propsDelim = "."
		valueDelim = "="
	case sm.Style == "matrix" && sm.Explode == false:
		prefix = ";" + param.Name + "="
		propsDelim = ","
		valueDelim = ","
	case sm.Style == "matrix" && sm.Explode == true:
		prefix = ";"
		propsDelim = ";"
		valueDelim = "="
	default:
		return nil, fmt.Errorf(errMsgInvalidSerializationF, param.In, param.Name, sm.Style, sm.Explode)
	}

	if d.input.PathParams == nil {
		// HTTP request does not contains a value of the target path parameter.
		return nil, nil
	}
	raw, ok := d.input.PathParams[d.paramKey(param, sm)]
	if !ok || raw == "" {
		// HTTP request does not contains a value of the target path parameter.
		return nil, nil
	}
	src, err := cutPrefix(raw, prefix)
	if err != nil {
		return nil, &RequestError{Input: d.input, Parameter: param, Reason: errMsgInvalidValue, Err: err}
	}
	props, err := propsFromString(src, propsDelim, valueDelim)
	if err != nil {
		return nil, &RequestError{Input: d.input, Parameter: param, Reason: errMsgInvalidValue, Err: err}
	}
	val, err := makeObject(props, param.Schema)
	if err != nil {
		return nil, &RequestError{Input: d.input, Parameter: param, Reason: errMsgInvalidValue, Err: err}
	}
	return val, nil
}

// paramKey returns a key to get a raw value of a path parameter.
func (d *pathParamDecoder) paramKey(param *openapi3.Parameter, sm *openapi3.SerializationMethod) string {
	switch sm.Style {
	case "label":
		return "." + param.Name
	case "matrix":
		return ";" + param.Name
	default:
		return param.Name
	}
}

// cutPrefix validates that a raw value of a path parameter has the specified prefix,
// and returns a raw value without the prefix.
func cutPrefix(raw, prefix string) (string, error) {
	if prefix == "" {
		return raw, nil
	}
	if len(raw) < len(prefix) || raw[:len(prefix)] != prefix {
		return "", &ParseError{Value: raw, Reason: fmt.Sprintf("a value must be prefixed with %q", prefix)}
	}
	return raw[len(prefix):], nil
}

// queryParamDecoder decodes values of query parameters.
type queryParamDecoder struct {
	input *RequestValidationInput
}

func (d *queryParamDecoder) DecodePrimitive(param *openapi3.Parameter) (interface{}, error) {
	sm, err := param.SerializationMethod()
	if err != nil {
		return nil, err
	}
	if sm.Style != "form" {
		return nil, fmt.Errorf(errMsgInvalidSerializationF, param.In, param.Name, sm.Style, sm.Explode)
	}

	values := d.input.GetQueryParams()[param.Name]
	if len(values) == 0 {
		// HTTP request does not contain a value of the target query parameter.
		return nil, nil
	}
	val, err := parsePrimitive(values[0], param.Schema)
	if err != nil {
		return nil, &RequestError{Input: d.input, Parameter: param, Reason: errMsgInvalidValue, Err: err}
	}
	return val, nil
}

func (d *queryParamDecoder) DecodeArray(param *openapi3.Parameter) ([]interface{}, error) {
	sm, err := param.SerializationMethod()
	if err != nil {
		return nil, err
	}
	if sm.Style == "deepObject" {
		return nil, fmt.Errorf(errMsgInvalidSerializationF, param.In, param.Name, sm.Style, sm.Explode)
	}

	values := d.input.GetQueryParams()[param.Name]
	if len(values) == 0 {
		// HTTP request does not contain a value of the target query parameter.
		return nil, nil
	}
	if !sm.Explode {
		var delim string
		switch sm.Style {
		case "form":
			delim = ","
		case "spaceDelimited":
			delim = " "
		case "pipeDelimited":
			delim = "|"
		}
		values = strings.Split(values[0], delim)
	}
	val, err := parseArray(values, param.Schema)
	if err != nil {
		return nil, &RequestError{Input: d.input, Parameter: param, Reason: errMsgInvalidValue, Err: err}
	}
	return val, nil
}

func (d *queryParamDecoder) DecodeObject(param *openapi3.Parameter) (map[string]interface{}, error) {
	var propsFn func(map[string][]string) (map[string]string, error)
	sm, err := param.SerializationMethod()
	if err != nil {
		return nil, err
	}
	switch sm.Style {
	case "form":
		propsFn = func(params map[string][]string) (map[string]string, error) {
			if len(params) == 0 {
				// HTTP request does not contain query parameters.
				return nil, nil
			}
			if sm.Explode {
				props := make(map[string]string)
				for key, values := range params {
					props[key] = values[0]
				}
				return props, nil
			}
			values := params[param.Name]
			if len(values) == 0 {
				// HTTP request does not contain a value of the target query parameter.
				return nil, nil
			}
			return propsFromString(values[0], ",", ",")
		}
	case "deepObject":
		propsFn = func(params map[string][]string) (map[string]string, error) {
			props := make(map[string]string)
			for key, values := range params {
				groups := regexp.MustCompile(fmt.Sprintf("%s\\[(.+?)\\]", param.Name)).FindAllStringSubmatch(key, -1)
				if len(groups) == 0 {
					// A query parameter's name does not match the required format, so skip it.
					continue
				}
				props[groups[0][1]] = values[0]
			}
			if len(props) == 0 {
				// HTTP request does not contain query parameters encoded by rules of style "deepObject".
				return nil, nil
			}
			return props, nil
		}
	default:
		return nil, fmt.Errorf(errMsgInvalidSerializationF, param.In, param.Name, sm.Style, sm.Explode)
	}

	props, err := propsFn(d.input.GetQueryParams())
	if err != nil {
		return nil, &RequestError{Input: d.input, Parameter: param, Reason: errMsgInvalidValue, Err: err}
	}
	if props == nil {
		return nil, nil
	}
	val, err := makeObject(props, param.Schema)
	if err != nil {
		return nil, &RequestError{Input: d.input, Parameter: param, Reason: errMsgInvalidValue, Err: err}
	}
	return val, nil
}

// headerParamDecoder decodes values of header parameters.
type headerParamDecoder struct {
	input *RequestValidationInput
}

func (d *headerParamDecoder) DecodePrimitive(param *openapi3.Parameter) (interface{}, error) {
	sm, err := param.SerializationMethod()
	if err != nil {
		return nil, err
	}
	if sm.Style != "simple" {
		return nil, fmt.Errorf(errMsgInvalidSerializationF, param.In, param.Name, sm.Style, sm.Explode)
	}

	raw := d.input.Request.Header.Get(http.CanonicalHeaderKey(param.Name))
	val, err := parsePrimitive(raw, param.Schema)
	if err != nil {
		return nil, &RequestError{Input: d.input, Parameter: param, Reason: errMsgInvalidValue, Err: err}
	}
	return val, nil
}

func (d *headerParamDecoder) DecodeArray(param *openapi3.Parameter) ([]interface{}, error) {
	sm, err := param.SerializationMethod()
	if err != nil {
		return nil, err
	}
	if sm.Style != "simple" {
		return nil, fmt.Errorf(errMsgInvalidSerializationF, param.In, param.Name, sm.Style, sm.Explode)
	}

	raw := d.input.Request.Header.Get(http.CanonicalHeaderKey(param.Name))

	val, err := parseArray(strings.Split(raw, ","), param.Schema)
	if raw == "" {
		// HTTP request does not contains a corresponding header
		return nil, nil
	}
	if err != nil {
		return nil, &RequestError{Input: d.input, Parameter: param, Reason: errMsgInvalidValue, Err: err}
	}
	return val, nil
}

func (d *headerParamDecoder) DecodeObject(param *openapi3.Parameter) (map[string]interface{}, error) {
	sm, err := param.SerializationMethod()
	if err != nil {
		return nil, err
	}
	if sm.Style != "simple" {
		return nil, fmt.Errorf(errMsgInvalidSerializationF, param.In, param.Name, sm.Style, sm.Explode)
	}
	valueDelim := ","
	if sm.Explode {
		valueDelim = "="
	}

	raw := d.input.Request.Header.Get(http.CanonicalHeaderKey(param.Name))
	if raw == "" {
		// HTTP request does not contain a corresponding header.
		return nil, nil
	}
	props, err := propsFromString(raw, ",", valueDelim)
	if err != nil {
		return nil, &RequestError{Input: d.input, Parameter: param, Reason: errMsgInvalidValue, Err: err}
	}
	val, err := makeObject(props, param.Schema)
	if err != nil {
		return nil, &RequestError{Input: d.input, Parameter: param, Reason: errMsgInvalidValue, Err: err}
	}
	return val, nil
}

// cookieParamDecoder decodes values of cookie parameters.
type cookieParamDecoder struct {
	input *RequestValidationInput
}

func (d *cookieParamDecoder) DecodePrimitive(param *openapi3.Parameter) (interface{}, error) {
	sm, err := param.SerializationMethod()
	if err != nil {
		return nil, err
	}
	if sm.Style != "form" {
		return nil, fmt.Errorf(errMsgInvalidSerializationF, param.In, param.Name, sm.Style, sm.Explode)
	}

	cookie, err := d.input.Request.Cookie(param.Name)
	if err == http.ErrNoCookie {
		// HTTP request does not contain a corresponding cookie.
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("decode param %q: %s", param.Name, err)
	}
	val, err := parsePrimitive(cookie.Value, param.Schema)
	if err != nil {
		return nil, &RequestError{Input: d.input, Parameter: param, Reason: errMsgInvalidValue, Err: err}
	}
	return val, nil
}

func (d *cookieParamDecoder) DecodeArray(param *openapi3.Parameter) ([]interface{}, error) {
	sm, err := param.SerializationMethod()
	if err != nil {
		return nil, err
	}
	if sm.Style != "form" || sm.Explode {
		return nil, fmt.Errorf(errMsgInvalidSerializationF, param.In, param.Name, sm.Style, sm.Explode)
	}

	cookie, err := d.input.Request.Cookie(param.Name)
	if err == http.ErrNoCookie {
		// HTTP request does not contain a corresponding cookie.
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("decode param %q: %s", param.Name, err)
	}
	val, err := parseArray(strings.Split(cookie.Value, ","), param.Schema)
	if err != nil {
		return nil, &RequestError{Input: d.input, Parameter: param, Reason: errMsgInvalidValue, Err: err}
	}
	return val, nil
}

func (d *cookieParamDecoder) DecodeObject(param *openapi3.Parameter) (map[string]interface{}, error) {
	sm, err := param.SerializationMethod()
	if err != nil {
		return nil, err
	}
	if sm.Style != "form" || sm.Explode {
		return nil, fmt.Errorf(errMsgInvalidSerializationF, param.In, param.Name, sm.Style, sm.Explode)
	}

	cookie, err := d.input.Request.Cookie(param.Name)
	if err == http.ErrNoCookie {
		// HTTP request does not contain a corresponding cookie.
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("decode param %q: %s", param.Name, err)
	}
	props, err := propsFromString(cookie.Value, ",", ",")
	if err != nil {
		return nil, &RequestError{Input: d.input, Parameter: param, Reason: errMsgInvalidValue, Err: err}
	}
	val, err := makeObject(props, param.Schema)
	if err != nil {
		return nil, &RequestError{Input: d.input, Parameter: param, Reason: errMsgInvalidValue, Err: err}
	}
	return val, nil
}

// propsFromString returns a properties map that is created by splitting a source string by propDelim and valueDelim.
// The source string must have a valid format: pairs <propName><valueDelim><propValue> separated by <propDelim>.
// The function returns an error when the source string has an invalid format.
func propsFromString(src, propDelim, valueDelim string) (map[string]string, error) {
	props := make(map[string]string)
	pairs := strings.Split(src, propDelim)

	// When propDelim and valueDelim is equal the source string follow the next rule:
	// every even item of pairs is a properies's name, and the subsequent odd item is a property's value.
	if propDelim == valueDelim {
		// Taking into account the rule above, a valid source string must be splitted by propDelim
		// to an array with an even number of items.
		if len(pairs)%2 != 0 {
			return nil, &ParseError{
				Value:  src,
				Reason: fmt.Sprintf("a value must be a list of object's properties in format \"name%svalue\" separated by %s", valueDelim, propDelim),
			}
		}
		for i := 0; i < len(pairs)/2; i++ {
			props[pairs[i*2]] = pairs[i*2+1]
		}
		return props, nil
	}

	// When propDelim and valueDelim is not equal the source string follow the next rule:
	// every item of pairs is a string that follows format <propName><valueDelim><propValue>.
	for _, pair := range pairs {
		prop := strings.Split(pair, valueDelim)
		if len(prop) != 2 {
			return nil, &ParseError{
				Value:  src,
				Reason: fmt.Sprintf("a value must be a list of object's properties in format \"name%svalue\" separated by %s", valueDelim, propDelim),
			}
		}
		props[prop[0]] = prop[1]
	}
	return props, nil
}

// makeObject returns an object that contains properties from props.
// A value of every property is parsed as a primitive value.
// The function returns an error when an error happened while parse object's properties.
func makeObject(props map[string]string, schema *openapi3.SchemaRef) (map[string]interface{}, error) {
	obj := make(map[string]interface{})
	for propName, propSchema := range schema.Value.Properties {
		value, err := parsePrimitive(props[propName], propSchema)
		if err != nil {
			if v, ok := err.(*ParseError); ok {
				return nil, &ParseError{
					Value:  v.Value,
					Reason: v.Reason,
					Cause:  v.Cause,
					Path:   []interface{}{propName},
				}
			}
			return nil, err
		}
		obj[propName] = value
	}
	return obj, nil
}

// parseArray returns an array that contains items from a raw array.
// Every item is parsed as a primitive value.
// The function returns an error when an error happened while parse array's items.
func parseArray(raw []string, schemaRef *openapi3.SchemaRef) ([]interface{}, error) {
	var value []interface{}
	for i, v := range raw {
		item, err := parsePrimitive(v, schemaRef.Value.Items)
		if err != nil {
			if v, ok := err.(*ParseError); ok {
				return nil, &ParseError{
					Value:  v.Value,
					Reason: v.Reason,
					Cause:  v.Cause,
					Path:   []interface{}{i},
				}
			}
			return nil, err
		}
		value = append(value, item)
	}
	return value, nil
}

// parsePrimitive returns a value that is created by parsing a source string to a primitive type
// that is specified by a JSON schema. The function returns nil when the source string is empty.
// The function panics when a JSON schema has a non primitive type.
func parsePrimitive(raw string, schema *openapi3.SchemaRef) (interface{}, error) {
	if raw == "" {
		return nil, nil
	}
	switch schema.Value.Type {
	case "integer", "number":
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return nil, &ParseError{Value: raw, Cause: err}
		}
		return v, nil
	case "boolean":
		v, err := strconv.ParseBool(raw)
		if err != nil {
			return nil, &ParseError{Value: raw, Cause: err}
		}
		return v, nil
	case "string":
		return raw, nil
	default:
		panic(fmt.Sprintf("schema has non primitive type %q", schema.Value.Type))
	}
}
