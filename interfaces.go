package e

import "errors"

// The following interfaces can be easily implemented by existing custom error types
// to maintain compatibility with package e.

// ClientFacing allows custom error types to be used with utility functions
// ErrorCode() and ErrorMessage().
type ClientFacing interface {

	// ClientCode returns the a short string representing the type of error, such as
	// "database_error", to be used by a client or an application.
	//
	// Note: ErrorCode() should be used to retrieve the topmost Code()
	ClientCode() string

	// ClientMessage returns a user-friendly error message (if any) which is logically
	// separate from the error cause.
	//
	// Note: ErrorMessage() should be used to retrieve the topmost Message().
	ClientMessage() string
}

// ErrorCode returns the first unwrapped Code of an error which implements
// ClientFacing interface. Otherwise returns an empty string.
func ErrorCode(err error) string {
	for err != nil {
		if e, ok := err.(ClientFacing); ok && e.ClientCode() != "" {
			return e.ClientCode()
		}
		err = errors.Unwrap(err)
	}
	return ""
}

// ErrorMessage returns the first unwrapped Message of an error which implements
// ClientFacing interface. Otherwise returns an empty string.
func ErrorMessage(err error) string {
	for err != nil {
		if e, ok := err.(ClientFacing); ok && e.ClientMessage() != "" {
			return e.ClientMessage()
		}
		err = errors.Unwrap(err)
	}
	return ""
}

// HasStacktrace allows custom error types to be used with utility function
// ErrorStacktrace().
type HasStacktrace interface {

	// Stacktrace returns the innermost stacktrace, if any.
	Stacktrace() string
}

// ErrorStacktrace returns the innermost Stack of an error which implements
// HasStacktrace interface. Otherwise returns an empty string.
func ErrorStacktrace(err error) string {
	var stack string
	for err != nil {
		if e, ok := err.(HasStacktrace); ok && e.Stacktrace() != "" {
			stack = e.Stacktrace()
		}
		err = errors.Unwrap(err)
	}
	return stack
}
