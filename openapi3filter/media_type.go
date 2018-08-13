package openapi3filter

// Definition of MediaType to be interpreted as JSON
var DefaultJSONMediaTypes = []string{
	"application/json",
}

func isMediaTypeJSON(mediaType string) bool {
	for _, mt := range DefaultJSONMediaTypes {
		if mediaType == mt {
			return true
		}
	}
	return false
}
