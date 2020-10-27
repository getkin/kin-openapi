package openapi3filter

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// ValidationErrorEncoder wraps a base ErrorEncoder to handle ValidationErrors
type ValidationErrorEncoder struct {
	Encoder ErrorEncoder
}

// Encode implements the ErrorEncoder interface for encoding ValidationErrors
func (enc *ValidationErrorEncoder) Encode(ctx context.Context, err error, w http.ResponseWriter) {
	var cErr *ValidationError

	if e, ok := err.(*RouteError); ok {
		cErr = convertRouteError(e)
		enc.Encoder(ctx, cErr, w)
		return
	}

	e, ok := err.(*RequestError)
	if !ok {
		enc.Encoder(ctx, err, w)
		return
	}

	if e.Err == nil {
		cErr = convertBasicRequestError(e)
	} else if e.Err == ErrInvalidRequired {
		cErr = convertErrInvalidRequired(e)
	} else if innerErr, ok := e.Err.(*ParseError); ok {
		cErr = convertParseError(e, innerErr)
	} else if innerErr, ok := e.Err.(*openapi3.SchemaError); ok {
		cErr = convertSchemaError(e, innerErr)
	}

	if cErr != nil {
		enc.Encoder(ctx, cErr, w)
	} else {
		enc.Encoder(ctx, err, w)
	}
}

func convertRouteError(e *RouteError) *ValidationError {
	var cErr *ValidationError
	switch e.Reason {
	case "Path doesn't support the HTTP method":
		cErr = &ValidationError{Status: http.StatusMethodNotAllowed, Title: e.Reason}
	default:
		cErr = &ValidationError{Status: http.StatusNotFound, Title: e.Reason}
	}
	return cErr
}

func convertBasicRequestError(e *RequestError) *ValidationError {
	var cErr *ValidationError
	unsupportedContentType := "header 'Content-Type' has unexpected value: "
	if strings.HasPrefix(e.Reason, unsupportedContentType) {
		if strings.HasSuffix(e.Reason, `: ""`) {
			cErr = &ValidationError{
				Status: http.StatusUnsupportedMediaType,
				Title:  "header 'Content-Type' is required",
			}
		} else {
			cErr = &ValidationError{
				Status: http.StatusUnsupportedMediaType,
				Title:  "unsupported content type " + strings.TrimPrefix(e.Reason, unsupportedContentType),
			}
		}
	} else {
		cErr = &ValidationError{
			Status: http.StatusBadRequest,
			Title:  e.Error(),
		}
	}
	return cErr
}

func convertErrInvalidRequired(e *RequestError) *ValidationError {
	var cErr *ValidationError
	if e.Reason == ErrInvalidRequired.Error() && e.Parameter != nil {
		cErr = &ValidationError{
			Status: http.StatusBadRequest,
			Title:  fmt.Sprintf("Parameter '%s' in %s is required", e.Parameter.Name, e.Parameter.In),
		}
	} else {
		cErr = &ValidationError{
			Status: http.StatusBadRequest,
			Title:  e.Error(),
		}
	}
	return cErr
}

func convertParseError(e *RequestError, innerErr *ParseError) *ValidationError {
	var cErr *ValidationError
	// We treat path params of the wrong type like a 404 instead of a 400
	if innerErr.Kind == KindInvalidFormat && e.Parameter != nil && e.Parameter.In == "path" {
		cErr = &ValidationError{
			Status: http.StatusNotFound,
			Title:  fmt.Sprintf("Resource not found with '%s' value: %v", e.Parameter.Name, innerErr.Value),
		}
	} else if strings.HasPrefix(innerErr.Reason, "unsupported content type") {
		cErr = &ValidationError{
			Status: http.StatusUnsupportedMediaType,
			Title:  innerErr.Reason,
		}
	} else if innerErr.RootCause() != nil {
		if rootErr, ok := innerErr.Cause.(*ParseError); ok &&
			rootErr.Kind == KindInvalidFormat && e.Parameter.In == "query" {
			cErr = &ValidationError{
				Status: http.StatusBadRequest,
				Title: fmt.Sprintf("Parameter '%s' in %s is invalid: %v is %s",
					e.Parameter.Name, e.Parameter.In, rootErr.Value, rootErr.Reason),
			}
		} else {
			cErr = &ValidationError{
				Status: http.StatusBadRequest,
				Title:  innerErr.Reason,
			}
		}
	}
	return cErr
}

var propertyMissingNameRE = regexp.MustCompile(`Property '(?P<name>[^']*)' is missing`)

func convertSchemaError(e *RequestError, innerErr *openapi3.SchemaError) *ValidationError {
	cErr := &ValidationError{Title: innerErr.Reason}

	// Handle "Origin" error
	if originErr, ok := innerErr.Origin.(*openapi3.SchemaError); ok {
		cErr = convertSchemaError(e, originErr)
	}

	// Add http status code
	if e.Parameter != nil {
		cErr.Status = http.StatusBadRequest
	} else if e.RequestBody != nil {
		cErr.Status = http.StatusUnprocessableEntity
	}

	// Add error source
	if e.Parameter != nil {
		// We have a JSONPointer in the query param too so need to
		// make sure 'Parameter' check takes priority over 'Pointer'
		cErr.Source = &ValidationErrorSource{
			Parameter: e.Parameter.Name,
		}
	} else if innerErr.JSONPointer() != nil {
		pointer := innerErr.JSONPointer()

		cErr.Source = &ValidationErrorSource{
			Pointer: toJSONPointer(pointer),
		}
	}

	// Add details on allowed values for enums
	if innerErr.SchemaField == "enum" &&
		innerErr.Reason == "JSON value is not one of the allowed values" {
		enums := make([]string, 0, len(innerErr.Schema.Enum))
		for _, enum := range innerErr.Schema.Enum {
			enums = append(enums, fmt.Sprintf("%v", enum))
		}
		cErr.Detail = fmt.Sprintf("Value '%v' at %s must be one of: %s",
			innerErr.Value, toJSONPointer(innerErr.JSONPointer()), strings.Join(enums, ", "))
		value := fmt.Sprintf("%v", innerErr.Value)
		if e.Parameter != nil &&
			(e.Parameter.Explode == nil || *e.Parameter.Explode == true) &&
			(e.Parameter.Style == "" || e.Parameter.Style == "form") &&
			strings.Contains(value, ",") {
			parts := strings.Split(value, ",")
			cErr.Detail = cErr.Detail + "; " + fmt.Sprintf("perhaps you intended '?%s=%s'",
				e.Parameter.Name, strings.Join(parts, "&"+e.Parameter.Name+"="))
		}
	}
	return cErr
}

func toJSONPointer(reversePath []string) string {
	return "/" + strings.Join(reversePath, "/")
}
