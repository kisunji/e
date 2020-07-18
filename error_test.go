package e

import (
	"errors"
	"fmt"
	"testing"
)

const (
	CodeUnexpected = "unexpected_error"
	CodeDatabase   = "database_error"
	CodeInternal   = "internal_error"
)

var errSentinel = NewError(CodeInternal, "sentinel error")

func TestErrors(t *testing.T) {
	tests := []struct {
		name string
		fn   func() error
		want string
	}{
		{
			name: "new constructs error",
			fn:   Foo,
			want: "Foo: [database_error] cannot foo",
		},
		{
			name: "wrap adds op",
			fn:   Bar,
			want: "Bar: Foo: [database_error] cannot foo",
		},
		{
			name: "wrapf adds formatted text",
			fn:   Barf,
			want: "Barf: (test id: 123): Foo: [database_error] cannot foo",
		},
		{
			name: "wrap adds op and optionalInfo",
			fn:   Fizz,
			want: "Fizz: (failed to fizz): Foo: [database_error] cannot foo",
		},
		{
			name: "can wrap non-package errors",
			fn:   Buzz,
			want: "Buzz: basic error",
		},
		{
			name: "can non-pkg wrap Error",
			fn:   FizzBuzz,
			want: "FizzBuzz: not encouraged but compatible: Foo: [database_error] cannot foo",
		},
		{
			name: "multiple SetCode display correctly",
			fn:   FizzBuzzWhiz,
			want: "FizzBuzzWhiz: [database_error] (changed code to database): badWrapper: [internal_error] (changed code to internal): BADWRAPBar: Foo: [database_error] cannot foo",
		},
		{
			name: "wrap non-pkg err then set code",
			fn:   FooBuzz,
			want: "FooBuzz: [database_error] database error",
		},
		{
			name: "works with lambdas",
			fn: func() error {
				return NewError(CodeInternal, "called from lambda")
			},
			want: "TestErrors.func1: [internal_error] called from lambda",
		},
		{
			name: "works with goroutines",
			fn: func() error {
				errChan := make(chan error)
				go func() {
					// force index of next goroutine to 2 (sanity check)
				}()
				go func(c chan<- error) {
					c <- NewError(CodeInternal, "called from lambda")
				}(errChan)
				return <-errChan
			},
			want: "TestErrors.func2.2: [internal_error] called from lambda",
		},
		{
			name: "NewErrorf formats string",
			fn:   Foof,
			want: "Foof: [database_error] id: 13",
		},
		{
			name: "sentinel error works", // not recommended though!
			fn: func() error {
				return errSentinel
			},
			want: "init: [internal_error] sentinel error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.fn(); err != nil && err.Error() != tt.want {
				t.Errorf("\ngot:  %q\nwant: %q", err, tt.want)
			}
		})
	}
}

func Foo() error {
	return NewError(CodeDatabase, "cannot foo")
}

func Bar() error {
	err := Foo()
	if err != nil {
		return Wrap(err)
	}
	return nil
}

func Barf() error {
	err := Foo()
	if err != nil {
		return Wrapf(err, "test id: %v", 123)
	}
	return nil
}

func Fizz() error {
	err := Foo()
	if err != nil {
		return Wrap(err, "failed to fizz")
	}
	return nil
}

func Buzz() error {
	err := errors.New("basic error")
	return Wrap(err)
}

func FizzBuzz() error {
	err := Foo()
	wrap := fmt.Errorf("not encouraged but compatible: %w", err)
	return Wrap(wrap)
}

func FizzBuzzWhiz() error {
	err := badWrapper()
	return Wrap(err, "changed code to database").SetCode(CodeDatabase)
}

func badWrapper() error {
	err := Bar()
	err2 := fmt.Errorf("%v%w", "BADWRAP", err)
	return Wrap(err2, "changed code to internal").SetCode(CodeInternal)
}

func FooBuzz() error {
	err := errors.New("database error")
	return Wrap(err).SetCode(CodeDatabase)
}

func Foof() error {
	format := 13
	return NewErrorf(CodeDatabase, "id: %d", format)
}

func TestErrorMessage(t *testing.T) {
	tests := []struct {
		name string
		fn   func() string
		want string
	}{
		{
			name: "unset message returns blank",
			fn: func() string {
				err := NewError(CodeUnexpected, "unexpected error occurred")
				return ErrorMessage(err)
			},
			want: "",
		},
		{
			name: "set message returns correctly",
			fn: func() string {
				err := NewError(CodeUnexpected, "unexpected error occurred").SetMessage("oh no")
				return ErrorMessage(err)
			},
			want: "oh no",
		},
		{
			name: "multiple messages but outermost message is returned",
			fn: func() string {
				err1 := NewError(CodeUnexpected, "bar").SetMessage("don't show this")

				err2 := Wrap(err1).SetMessage("show this")

				return ErrorMessage(err2)
			},
			want: "show this",
		},
		{
			name: "works with non-pkg wrapping",
			fn: func() string {
				err := NewError(CodeInternal, "cannot do something")
				err = Wrap(err).SetMessage("wrapped by fmt.Errorf")

				wrap := fmt.Errorf("not encouraged but compatible: %w", err)

				wrap2 := Wrap(wrap)

				return ErrorMessage(wrap2)
			},
			want: "wrapped by fmt.Errorf",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fn(); got != tt.want {
				t.Errorf("\ngot:  %q\nwant: %q", got, tt.want)
			}
		})
	}
}

func TestErrorCode(t *testing.T) {
	tests := []struct {
		name string
		fn   func() string
		want string
	}{
		{
			name: "unset code returns blank",
			fn: func() string {
				err := NewError("", "unexpected error occurred")
				return ErrorCode(err)
			},
			want: "",
		},
		{
			name: "set code with NewError() returns correctly",
			fn: func() string {
				err := NewError(CodeUnexpected, "unexpected error occurred")
				return ErrorCode(err)
			},
			want: CodeUnexpected,
		},
		{
			name: "set code with Wrap() returns correctly",
			fn: func() string {
				err := errors.New("db error occurred")
				wrap := Wrap(err).SetCode(CodeDatabase)
				return ErrorCode(wrap)
			},
			want: CodeDatabase,
		},
		{
			name: "setting multiple codes but last code is returned",
			fn: func() string {
				err1 := NewError(CodeUnexpected, "bar")

				err2 := Wrap(err1).SetCode(CodeInternal)

				return ErrorCode(err2)
			},
			want: CodeInternal,
		},
		{
			name: "returns outermost code",
			fn: func() string {
				err1 := NewError(CodeUnexpected, "bar")

				err2 := Wrap(err1).SetCode(CodeInternal)

				return ErrorCode(err2)
			},
			want: CodeInternal,
		},
		{
			name: "works with non-pkg wrapping",
			fn: func() string {
				err := NewError(CodeInternal, "cannot do something")

				err = Wrap(err)

				wrap := fmt.Errorf("not encouraged but compatible: %w", err)

				wrap2 := Wrap(wrap)

				return ErrorCode(wrap2)
			},
			want: CodeInternal,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fn(); got != tt.want {
				t.Errorf("\ngot:  %q\nwant: %q", got, tt.want)
			}
		})
	}
}

func TestErrorStack(t *testing.T) {
	t.Run("ErrorStacktrace returns something", func(t *testing.T) {
		err := NewError("", "unexpected error occurred")
		if ErrorStacktrace(err) == "" {
			t.Fatalf("expected stacktrace from ErrorStacktrace() but got none")
		}
	})
	t.Run("ErrorStacktrace returns inner stacktrace", func(t *testing.T) {
		err := NewError("", "unexpected error occurred")
		badError := errorImpl{
			op:         "BAD",
			code:       "BAD",
			message:    "BAD",
			err:        err,
			stacktrace: "BAD",
		}
		if ErrorStacktrace(badError) == "BAD" {
			t.Fatalf("expected inner stacktrace from ErrorStacktrace() but got outer")
		}
	})
}

func Benchmark_getCallingFunc(b *testing.B) {
	for n := 0; n < b.N; n++ {
		getCallingFunc(0)
	}
}

func Test_getCallingFunc(t *testing.T) {
	tests := []struct {
		name        string
		frameOffset int
		want        string
	}{
		{
			name:        "gets itself",
			frameOffset: 0,
			want:        "getCallingFunc",
		},
		{
			name:        "any negative number returns Callers",
			frameOffset: -1000,
			want:        "Callers",
		},
		{
			name:        "super high number returns unknown",
			frameOffset: 9999,
			want:        "unknown",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getCallingFunc(tt.frameOffset); got != tt.want {
				t.Errorf("getCallingFunc() = %v, want %v", got, tt.want)
			}
		})
	}
}