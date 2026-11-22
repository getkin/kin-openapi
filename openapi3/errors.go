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
		// branch list stays readable. Empty lines stay empty so the message
		// never carries trailing whitespace.
		lines := strings.Split(branchBriefError(branch), "\n")
		for j, line := range lines {
			if j > 0 && line != "" {
				lines[j] = "    " + line
			}
		}
		fmt.Fprintf(&sb, "\n  %s: %s", e.labels[i], strings.Join(lines, "\n"))
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
	case *oneOfNoMatchError, *oneOfConflictError:
		// Their own rendering is already a compact branch list.
		return e.Error()
	case *messageError:
		// Its message was precomputed to be detail-free.
		return e.Error()
	case multiErrorForOneOf:
		return joinBriefErrors(e, " Or ")
	case multiErrorForAllOf:
		return joinBriefErrors(e, " And ")
	case MultiError:
		return joinBriefErrors(e, " | ")
	default:
		// Unwrap wrappers (e.g. fmt.Errorf("...: %w", sub)) so nested
		// *SchemaError values render briefly too, keeping the wrapper's own
		// message as a prefix when it can be recovered.
		if sub := errors.Unwrap(err); sub != nil {
			if _, ok := sub.(*SchemaError); ok {
				if msg, subMsg := err.Error(), sub.Error(); strings.HasSuffix(msg, subMsg) {
					return strings.TrimSuffix(msg, subMsg) + briefError(sub)
				}
			}
		}
		return err.Error()
	}
}

// joinBriefErrors renders each of errs briefly, joined with sep.
func joinBriefErrors(errs []error, sep string) string {
	msgs := make([]string, 0, len(errs))
	for _, sub := range errs {
		msgs = append(msgs, briefError(sub))
	}
	return strings.Join(msgs, sep)
}

// branchBriefError renders a oneOf branch's error for the oneOf branch list:
// like briefError, but a branch whose Origin wraps a MultiError (e.g. an allOf
// failure reported as "doesn't match schema due to: ...") renders as just the
// multi's brief list. The wrapper's literal prefix cannot be recovered (its
// message is frozen at wrap time, while a *SchemaError's message keeps gaining
// path segments as it propagates), and the branch label already carries that
// context.
func branchBriefError(err error) string {
	if se, ok := err.(*SchemaError); ok && se.Origin != nil {
		switch se.Origin.(type) {
		case *oneOfNoMatchError, *oneOfConflictError:
			// Already compact; briefError renders it as is.
		default:
			if sub := errors.Unwrap(se.Origin); sub != nil {
				switch sub.(type) {
				case multiErrorForOneOf, multiErrorForAllOf, MultiError:
					return se.briefWithOriginBrief(briefError(sub))
				}
			}
		}
	}
	return briefError(err)
}
