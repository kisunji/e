package e

import (
	"errors"
	"fmt"
	"strings"
)

// Error represents a standard application error.
// Error should always have a non-nil nested err and therefore the type cannot
// itself be the root of an error stack.
type Error struct {
	// Represents the error type to be used by client or application.
	// Should only exist at the root error (SetCode() enforces this).
	// e.g. "unexpected_error", "database_error", "not_exists" etc.
	code string

	// A user-friendly error message.
	// Does not get printed with Error().
	message string

	// Operation being performed--usually a function/method name.
	op string

	// Nested error for building an Error() stacktrace. Should not be nil.
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
	if e.code != "" {
		sb.WriteString(fmt.Sprintf("[%s] ", e.code))
	}
	sb.WriteString(e.err.Error())

	return sb.String()
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.err
}

func (e *Error) Code() string {
	if e == nil {
		return ""
	}
	if e.code != "" {
		return e.code
	}
	for err := e.err; err != nil; err = errors.Unwrap(err) {
		if e, ok := err.(*Error); ok && e.code != "" {
			return e.code
		}
	}
	return ""
}

// SetCode adds an error type *Error
func (e *Error) SetCode(code string) *Error {
	if e != nil {
		e.code = code
	}
	return e
}

func (e *Error) Message() string {
	if e == nil {
		return ""
	}
	if e.message != "" {
		return e.message
	}
	for err := e.err; err != nil; err = errors.Unwrap(err) {
		if e, ok := err.(*Error); ok && e.message != "" {
			return e.message
		}
	}
	return ""
}

// SetMessage adds a user-friendly message to *Error.
// Important: ensure the string is localized for the end-user.
func (e *Error) SetMessage(message string) *Error {
	if e != nil {
		e.message = message
	}
	return e
}

// ClearMessage unsets all messages from *Error's stack.
func (e *Error) ClearMessage() *Error {
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

// New constructs a new *Error.
//
// Usage:
// 		func Foo() error {
// 			const op = "Foo"
// 			if bar != nil {
// 				return e.New(op, "unexpected_error", "bar not nil")
// 			}
//			...
//			return nil
//		}
//
func New(op, code, cause string) *Error {
	return &Error{
		op:   op,
		code: code,
		err:  errors.New(cause),
	}
}

// Wrap adds op to the logical stacktrace. Recommended to chain with SetCode()
// if wrapping an external error type that does not implement ClientFacing.
//
// OptionalInfo can be passed to insert more context at the wrap site.
// Only the first OptionalInfo string will be used.
//
// Basic usage:
// 		err := Foo()
//		if err != nil {
// 			return e.Wrap(op, err)
// 		}
//
// Wrapping an external error:
//		err := db.Get()
// 		if err != nil {
//			return e.Wrap(op, err).SetCode("database_error")
//		}
//
// Adding additional info:
// 		err := Foo()
//		if err != nil {
// 			return e.Wrap(op, err, fmt.Sprintf("cannot find id: %v", id))
// 		}
//
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

// ClientFacing allows custom error types to be used with utility functions
// ErrorCode() and ErrorMessage().
type ClientFacing interface {
	Code() string
	Message() string
}

// ErrorCode returns the first unwrapped Code of an error which implements
// ClientFacing interface. Otherwise returns an empty string.
func ErrorCode(err error) string {
	for err != nil {
		if e, ok := err.(ClientFacing); ok && e.Code() != "" {
			return e.Code()
		}
		err = errors.Unwrap(err)
	}
	return ""
}

// ErrorMessage returns the first unwrapped Message of an error which implements
// ClientFacing interface. Otherwise returns an empty string.
func ErrorMessage(err error) string {
	for err != nil {
		if e, ok := err.(ClientFacing); ok && e.Message() != "" {
			return e.Message()
		}
		err = errors.Unwrap(err)
	}
	return ""
}
