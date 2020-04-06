package errs

import (
	"errors"
	"fmt"
	"strings"
)

// Error is a standard application error
// This type should always have a non-nil nested err
// and therefore cannot itself be the root of an error stack
type Error struct {
	// For use by application or clients
	// e.g. "unexpected_error", "database_error", "not_exists" etc.
	code string

	// User-friendly localized error message
	message string

	// Operation (function name) and nested error
	op  string
	err error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}

	var sb strings.Builder
	if e.op != "" {
		sb.WriteString(fmt.Sprintf("%s: ", e.op))
	}

	// If nested err wraps an error, write its Error() message.
	// Otherwise write the root's Error() and code
	if errors.Unwrap(e.err) != nil {
		sb.WriteString(e.err.Error())
	} else {
		sb.WriteString(e.err.Error())
		if e.code != "" {
			sb.WriteString(fmt.Sprintf(" [%s]", e.code))
		}
	}
	if ErrorMessage(e.err) != "" {
		sb.WriteString(" " + e.message)
	}

	return sb.String()
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.err
}

// OverwriteCode will replace the root *Error's code
func (e *Error) OverwriteCode(code string) *Error {
	if e == nil {
		return nil
	}
	for err := e.err; err != nil; err = errors.Unwrap(err) {
		if e, ok := err.(*Error); ok && e.code != "" {
			e.code = code
		}
	}
	return e
}

// SetClientMsg adds a user-friendly message to *Error, overwriting any existing messages
// Important: ensure the string is localized for the end-user
func (e *Error) SetClientMsg(localizedMsg string) *Error {
	_ = e.ClearClientMsg()
	if e != nil {
		e.message = localizedMsg
	}
	return e
}

// ClearClientMsg unsets message through the entire error stack
func (e *Error) ClearClientMsg() *Error {
	if e == nil {
		return nil
	}
	e.message = ""
	for err := e.err; err != nil; err = errors.Unwrap(err) {
		if e, ok := err.(*Error); ok && e.message != "" {
			e.message = ""
		}
	}
	return e
}

// Constructs a new Error struct
// cause is used to create a nested error
func New(op, code, cause string) *Error {
	return &Error{
		op:   op,
		code: code,
		err:  errors.New(cause),
	}
}

// Wrap adds op to the logical stacktrace
// OptionalInfo can be passed to insert more context at the wrap site
// Only the first OptionalInfo string will be used
//
// Basic Usage:
// 		err := foo()
//		if err != nil {
// 			return e.Wrap(op, err)
// 		}
//
// Adding Info:
// 		err := foo()
//		if err != nil {
// 			return e.Wrap(op, err, fmt.Sprintf("cannot find id: %v", id))
// 		}
//
// Optionally, function can be modified to provide a Wrapf behaviour
// There will be extra effort to convert optionalInfo
// from `...string` to `...interface{}` and check type safety
// You will also lose format linting/checking in IDEs
// Assuming Basic Usage will be the common use case, I am against the feature
func Wrap(op string, err error, optionalInfo ...string) *Error {
	innerErr := err
	if len(optionalInfo) > 0 {
		innerErr = fmt.Errorf("(%v): %w", optionalInfo[0], err)
	}
	return &Error{
		op:  op,
		err: innerErr,
	}
}

func ErrorCode(err error) string {
	if err == nil {
		return ""
	} else if e, ok := err.(*Error); ok && e.code != "" {
		return e.code
	} else if ok && e.err != nil {
		return ErrorCode(e.err)
	}
	return "unexpected_error"
}

func ErrorMessage(err error) string {
	for err != nil {
		if e, ok := err.(*Error); ok && e.message != ""{
			return e.message
		}
		err = errors.Unwrap(err)
	}
	return ""
}
