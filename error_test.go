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

func TestErrors(t *testing.T) {
	tests := []struct {
		name string
		fn   func() error
		want string
	}{
		{
			name: "new constructs error",
			fn: func() error {
				const op = "Foo"

				return New(op, CodeDatabase, "cannot do something")
			},
			want: "Foo: [database_error] cannot do something",
		},
		{
			name: "wrap adds op",
			fn: func() error {
				const op = "Inner"
				err := New(op, CodeInternal, "cannot do something")

				const op2 = "Outer"
				return Wrap(op2, err)
			},
			want: "Outer: Inner: [internal_error] cannot do something",
		},
		{
			name: "wrap adds op and optionalInfo",
			fn: func() error {
				const op = "Inner"
				err := New(op, CodeInternal, "cannot do something")

				const op2 = "Outer"
				return Wrap(op2, err, "optional info here")
			},
			want: "Outer: (optional info here): Inner: [internal_error] cannot do something",
		},
		{
			name: "can wrap non-package errors",
			fn: func() error {
				const op = "Foo"
				err := errors.New("basic error")
				return Wrap(op, err)
			},
			want: "Foo: basic error",
		},
		{
			name: "can non-pkg wrap Error",
			fn: func() error {
				const op = "Inner"
				err := New(op, CodeInternal, "cannot do something")

				const op2 = "Outer"
				err = Wrap(op2, err)

				wrap := fmt.Errorf("not encouraged but compatible: %w", err)

				const op3 = "Outer2"
				return Wrap(op3, wrap)
			},
			want: "Outer2: not encouraged but compatible: Outer: Inner: [internal_error] cannot do something",
		},
		{
			name: "multiple SetCode display correctly",
			fn: func() error {
				const op = "Inner"
				err := New(op, CodeUnexpected, "unexpected error occurred")

				const op2 = "Outer"
				err2 := Wrap(op2, err)

				err3 := fmt.Errorf("%v%w", "BADWRAP", err2)

				const op3 = "Outer2"
				err4 := Wrap(op3, err3, "changed code to internal").SetCode(CodeInternal)

				err5 := fmt.Errorf("%v%w", "MOREBADWRAP", err4)

				const op4 = "Outer3"
				err6 := Wrap(op4, err5, "changed code to database").SetCode(CodeDatabase)
				return err6
			},
			want: "Outer3: [database_error] (changed code to database): MOREBADWRAPOuter2: [internal_error] (changed code to internal): BADWRAPOuter: Inner: [unexpected_error] unexpected error occurred",
		},
		{
			name: "wrap non-pkg err then set code",
			fn: func() error {
				const op = "Foo"
				err := errors.New("database error")
				return Wrap(op, err).SetCode(CodeDatabase)
			},
			want: "Foo: [database_error] database error",
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

func TestErrorMessage(t *testing.T) {
	tests := []struct {
		name string
		fn   func() string
		want string
	}{
		{
			name: "unset message returns blank",
			fn: func() string {
				op := "Foo"
				err := New(op, CodeUnexpected, "unexpected error occurred")
				return ErrorMessage(err)
			},
			want: "",
		},
		{
			name: "set message returns correctly",
			fn: func() string {
				op := "Foo"
				err := New(op, CodeUnexpected, "unexpected error occurred").SetMessage("oh no")
				return ErrorMessage(err)
			},
			want: "oh no",
		},
		{
			name: "multiple messages but outermost message is returned",
			fn: func() string {
				op := "Foo"
				err1 := New(op, CodeUnexpected, "bar").SetMessage("don't show this")

				op2 := "Foo2"
				err2 := Wrap(op2, err1).SetMessage("show this")

				return ErrorMessage(err2)
			},
			want: "show this",
		},
		{
			name: "works with non-pkg wrapping",
			fn: func() string {
				const op = "Inner"
				err := New(op, CodeInternal, "cannot do something")

				const op2 = "Outer"
				err = Wrap(op2, err).SetMessage("wrapped by fmt.Errorf")

				wrap := fmt.Errorf("not encouraged but compatible: %w", err)

				const op3 = "Outer2"
				wrap2 := Wrap(op3, wrap)

				return ErrorMessage(wrap2)
			},
			want: "wrapped by fmt.Errorf",
		},
		{
			name: "cleared message does not get returned",
			fn: func() string {
				const op = "Foo"
				err := New(op, CodeInternal, "fail fail fail").SetMessage("clear me!")

				const op2 = "Outer"
				err = Wrap(op2, err).SetMessage("clear me too!")

				const op3 = "Outer2"
				err = Wrap(op3, err).SetMessage("clear all of us!")

				err = err.ClearMessage()
				return ErrorMessage(err)
			},
			want: "",
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
				op := "Foo"
				err := New(op, "", "unexpected error occurred")
				return ErrorCode(err)
			},
			want: "",
		},
		{
			name: "set code with New() returns correctly",
			fn: func() string {
				op := "Foo"
				err := New(op, CodeUnexpected, "unexpected error occurred")
				return ErrorCode(err)
			},
			want: CodeUnexpected,
		},
		{
			name: "set code with Wrap() returns correctly",
			fn: func() string {
				op := "Foo"
				err := errors.New("db error occurred")
				wrap := Wrap(op, err).SetCode(CodeDatabase)
				return ErrorCode(wrap)
			},
			want: CodeDatabase,
		},
		{
			name: "setting multiple codes but last code is returned",
			fn: func() string {
				op := "Foo"
				err1 := New(op, CodeUnexpected, "bar")

				op2 := "Foo2"
				err2 := Wrap(op2, err1).SetCode(CodeInternal)

				return ErrorCode(err2)
			},
			want: CodeInternal,
		},
		{
			name: "returns outermost code",
			fn: func() string {
				op := "Foo"
				err1 := New(op, CodeUnexpected, "bar")

				op2 := "Foo2"
				err2 := Wrap(op2, err1).SetCode(CodeInternal)

				return ErrorCode(err2)
			},
			want: CodeInternal,
		},
		{
			name: "works with non-pkg wrapping",
			fn: func() string {
				const op = "Inner"
				err := New(op, CodeInternal, "cannot do something")

				const op2 = "Outer"
				err = Wrap(op2, err)

				wrap := fmt.Errorf("not encouraged but compatible: %w", err)

				const op3 = "Outer2"
				wrap2 := Wrap(op3, wrap)

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
