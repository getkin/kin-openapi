package openapi3

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

// MultiError is a collection of errors, intended for when
// multiple issues need to be reported upstream
type MultiError []error

func (me MultiError) Error() string {
	return spliceErr(" | ", me)
}

func spliceErr(sep string, errs []error) string {
	buff := &bytes.Buffer{}
	for i, e := range errs {
		buff.WriteString(e.Error())
		if i != len(errs)-1 {
			buff.WriteString(sep)
		}
	}
	return buff.String()
}

// Is allows you to determine if a generic error is in fact a MultiError using `errors.Is()`
// It will also return true if any of the contained errors match target
func (me MultiError) Is(target error) bool {
	if _, ok := target.(MultiError); ok {
		return true
	}
	for _, e := range me {
		if errors.Is(e, target) {
			return true
		}
	}
	return false
}

// As allows you to use `errors.As()` to set target to the first error within the multi error that matches the target type
func (me MultiError) As(target any) bool {
	for _, e := range me {
		if errors.As(e, target) {
			return true
		}
	}
	return false
}

type multiErrorForOneOf MultiError

func (meo multiErrorForOneOf) Error() string {
	return spliceErr(" Or ", meo)
}

func (meo multiErrorForOneOf) Unwrap() error {
	return MultiError(meo)
}

type multiErrorForAllOf MultiError

func (mea multiErrorForAllOf) Error() string {
	return spliceErr(" And ", mea)
}

func (mea multiErrorForAllOf) Unwrap() error {
	return MultiError(mea)
}

// oneOfNoMatchError is the Origin of the *SchemaError reported when a value
// matches none of a schema's oneOf branches. Its message names the oneOf and
// lists every attempted branch's error, labeled by the branch's index (and
// its $ref or title when it has one), so the reasons the value failed are
// visible at a glance instead of buried in one branch's schema dump.
// Unwrap exposes the branch errors as a multiErrorForOneOf, and through it as
// a MultiError.
type oneOfNoMatchError struct {
	branches multiErrorForOneOf
	// labels describes each branch in the message, parallel to branches.
	labels []string
}

func (e *oneOfNoMatchError) Error() string {
	var sb strings.Builder
	sb.WriteString(`value doesn't match any schema from "oneOf"`)
	if len(e.branches) != 0 {
		sb.WriteString(":")
	}
	for i, branch := range e.branches {
		// Indent continuation lines of nested combinator errors so the
		// branch list stays readable.
		msg := strings.ReplaceAll(briefError(branch), "\n", "\n    ")
		fmt.Fprintf(&sb, "\n  %s: %s", e.labels[i], msg)
	}
	return sb.String()
}

// Unwrap returns the branch errors, so errors.As/errors.Is reach them.
func (e *oneOfNoMatchError) Unwrap() error { return e.branches }

// oneOfConflictError is the Origin of the *SchemaError reported when a value
// matches more than one branch of a schema's oneOf. Its message names the
// matching branches. Unwrap exposes ErrOneOfConflict.
type oneOfConflictError struct {
	indices []int
}

func (e *oneOfConflictError) Error() string {
	return fmt.Sprintf(`value matches more than one schema from "oneOf" (matches schemas at indices %v)`, e.indices)
}

// Unwrap exposes ErrOneOfConflict, so errors.Is(err, ErrOneOfConflict) holds.
func (e *oneOfConflictError) Unwrap() error { return ErrOneOfConflict }

// briefError renders err without the "Schema:"/"Value:" details that
// *SchemaError.Error appends, for listing one branch's error among several in
// a combinator error message, where repeating the details of every branch
// would drown the actual reasons.
func briefError(err error) string {
	switch e := err.(type) {
	case *SchemaError:
		return e.brief()
	case multiErrorForOneOf:
		return briefError(MultiError(e))
	case multiErrorForAllOf:
		return briefError(MultiError(e))
	case MultiError:
		msgs := make([]string, 0, len(e))
		for _, sub := range e {
			msgs = append(msgs, briefError(sub))
		}
		return strings.Join(msgs, " | ")
	default:
		return err.Error()
	}
}
