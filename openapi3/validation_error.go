package openapi3

// ValidationError is the typed form of an error returned by the document
// validation walker (T.Validate, Info.Validate, Paths.Validate, etc.).
//
// Today most of those validators return plain errors.New(...) values, so
// downstream code that wants to react to a specific validation failure has
// to compare on the human-readable message string. That's brittle: any
// upstream rewording silently breaks the consumer. ValidationError gives
// each failure a stable Code that callers can switch on via errors.As,
// while keeping the same Error() string so existing string-matching code
// continues to work unchanged.
//
// Codes follow a kebab-case subject-action shape (e.g. "info-version-required",
// "path-leading-slash-missing"), modeled after the rule-ID convention used
// by oasdiff and similar tools. The Code field is a bare string rather than
// a typed constant so validators in any package can introduce new codes
// without a central registry; callers that want compile-time safety can
// declare their own const block.
//
// Backward compatibility: this type is purely additive. Every site that
// today returns errors.New(msg) can be migrated to return
// &ValidationError{Code: "...", Message: msg} without changing the
// observed Error() output. Callers that don't use errors.As see no
// difference.
type ValidationError struct {
	Code    string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
