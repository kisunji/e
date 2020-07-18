// Package e is an error-handling package designed to be simple, safe, and
// compatible.
package e

import (
	"errors"
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
)

// Error represents a standard application error.
// Implements ClientFacing and HasStacktrace so it can be introspected
// with functions like ErrorCode, ErrorMessage, and ErrorStacktrace.
type Error interface {
	error
	ClientFacing
	HasStacktrace

	Unwrap() error

	// SetCode adds an error type to a non-nil Error such as "unexpected_error",
	// "database_error", "not_exists", etc.
	//
	// Will panic when used with a nil Error receiver.
	SetCode(code string) Error

	// SetMessage adds a user-friendly message to a non-nil Error.
	// Message will not be printed with Error() and should be retrieved with ErrorMessage().
	//
	// Will panic when used with a nil Error receiver.
	SetMessage(message string) Error
}

// NewError constructs a new Error. code should be a short, single string
// describing the type of error (typically a pre-defined const). cause is used
// to create the nested error which will act as the root of the error stack.
//
// Usage:
// 		func Foo(bar *Bar) error {
// 			if bar == nil {
// 				return e.New("unexpected_error", "bar is nil")
// 			}
//			return doFoo(bar)
//		}
//
func NewError(code, cause string) Error {
	return errorImpl{
		op:         getCallingFunc(2),
		code:       code,
		err:        errors.New(cause),
		stacktrace: string(debug.Stack()),
	}
}

// NewErrorf constructs a new Error with formatted string. code should be a short,
// single string describing the type of error (typically a pre-defined const).
// cause is used to create the nested error which will act as the root of the error stack.
//
// Usage:
// 		func Foo(bar Bar) error {
// 			done := doFoo(bar)
// 			if !done {
// 				return e.NewErrorf("unexpected_error", "cannot process bar: %v", bar)
// 			}
//			return nil
//		}
//
func NewErrorf(code, fmtCause string, args ...interface{}) Error {
	return errorImpl{
		op:         getCallingFunc(2),
		code:       code,
		err:        fmt.Errorf(fmtCause, args...),
		stacktrace: string(debug.Stack()),
	}
}

// Wrap adds the name of the calling function to the wrapped error.
// OptionalInfo can be passed to insert more context at the wrap site.
// Only the first OptionalInfo string will be used.
//
// Basic usage:
// 		err := Foo()
//		if err != nil {
// 			return e.Wrap(err)
// 		}
//
// Adding additional info:
// 		err := Foo()
//		if err != nil {
// 			return e.Wrap(err, fmt.Sprintf("cannot find id: %v", id))
// 		}
//
func Wrap(err error, optionalInfo ...string) Error {
	if err == nil {
		return nil
	}

	innerErr := err
	if len(optionalInfo) > 0 {
		innerErr = fmt.Errorf("(%v): %w", optionalInfo[0], err) // localizer.Ignore
	}

	wrapped := errorImpl{
		op:         getCallingFunc(2),
		err:        innerErr,
		stacktrace: ErrorStacktrace(err),
	}

	if wrapped.stacktrace == "" {
		wrapped.stacktrace = string(debug.Stack())
	}

	return wrapped
}

// Wrapf adds the name of the calling function and a formatted message
// to the wrapped error.
//
// Basic usage:
// 		err := Foo(bar)
//		if err != nil {
// 			return e.Wrapf(err, "bar not found: %v", bar.Id)
// 		}
//
func Wrapf(err error, fmtInfo string, args ...interface{}) Error {
	if err == nil {
		return nil
	}

	wrapped := errorImpl{
		op:         getCallingFunc(2),
		err:        fmt.Errorf("(%v): %w", fmt.Sprintf(fmtInfo, args...), err), // localizer.Ignore
		stacktrace: ErrorStacktrace(err),
	}

	if wrapped.stacktrace == "" {
		wrapped.stacktrace = string(debug.Stack())
	}

	return wrapped
}

// errorImpl should always have a non-nil nested err and therefore this type
// cannot by itself be the true root of an error stack.
type errorImpl struct {
	// Operation being performed--populated at runtime automagically
	op string

	// Represents the error type to be used by client or application.
	// e.g. "unexpected_error", "database_error", "not_exists" etc.
	// Use ErrorCode(err) to retrieve the outermost code.
	code string

	// A user-friendly error message. Does not get printed with Error().
	// Use ErrorMessage(err) to retrieve the outermost message.
	message string

	// Nested error for building an error stacktrace. Should not be nil.
	err error

	// Internal stacktrace for logging. Does not get printed with Error().
	// Use ErrorStacktrace(err) to retrieve the innermost stacktrace.
	stacktrace string
}

func (e errorImpl) Error() string {
	var sb strings.Builder
	if e.op != "" {
		sb.WriteString(fmt.Sprintf("%s: ", e.op))
	}
	if e.code != "" {
		sb.WriteString(fmt.Sprintf("[%s] ", e.code)) // localizer.Ignore
	}
	sb.WriteString(e.err.Error())

	return sb.String()
}

func (e errorImpl) Unwrap() error {
	return e.err
}

func (e errorImpl) ClientCode() string {
	return e.code
}

func (e errorImpl) ClientMessage() string {
	return e.message
}

func (e errorImpl) SetCode(code string) Error {
	e.code = code
	return e
}

func (e errorImpl) SetMessage(message string) Error {
	e.message = message
	return e
}

func (e errorImpl) Stacktrace() string {
	return e.stacktrace
}

// getCallingFunc returns the name of the calling function N levels
// above getCallingFunc (e.g. 0 for `getCallingFunc` itself)
func getCallingFunc(frameOffset int) string {
	// only need len = 1 to contain the calling function
	programCounters := make([]uintptr, 1)
	// base offset is 1 to skip `runtime.Callers` itself
	n := runtime.Callers(1+frameOffset, programCounters)
	if n == 0 {
		return "unknown"
	}
	frames := runtime.CallersFrames(programCounters)
	frame, _ := frames.Next()

	// Remove package name (too verbose)
	ss := strings.Split(frame.Function, "/")
	funcname := ss[len(ss)-1]
	return strings.SplitAfterN(funcname, ".", 2)[1]
}
