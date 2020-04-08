# e

![Go](https://github.com/kisunji/e/workflows/Go/badge.svg)

[pkg.go.dev link](https://pkg.go.dev/github.com/kisunji/e?tab=doc)

Inspired by [Failure is your Domain](https://middlemost.com/failure-is-your-domain/) and [Error handling in upspin](https://commandcenter.blogspot.com/2017/12/error-handling-in-upspin.html) (comparison [here](#comparisons-with-other-approaches)), package `e` is designed to meet the needs of web applications by maintaining a clean separation between the error data consumed by end-users, clients, and operators.

`e` focuses on three principles:

- **Simplicity**: Small, fluent interface for developer ergonomics

- **Safety**: Explicitly separate end-user messages from internal errors

- **Compatibility**: Follows error-wrapping conventions and can be adopted incrementally

## Usage

### Creating a new Error

```go
const CodeInvalidError = "invalid_error"

func Foo(bar string) error {
    const op = "Foo"

    if bar == nil {
        return e.New(op, CodeInvalidError, "bar cannot be nil")
        // "Foo: [invalid_error] bar cannot be nil"
    }
    return nil
}
```

### Wrapping an existing Error

The `const op = "FuncName"` pattern reduces the burden on developers of providing context to the error and helps keep the error stack free of redundant or unhelpful messages.

```go
func Foo(bar string) error {
    const op = "Foo"

    err := db.GetBar(bar) // "GetBar: [database_error] cannot find bar"
    if err != nil {
        return e.Wrap(op, err)
        // "Foo: GetBar: [database_error] cannot find bar"
    }
    return nil
}
```

`Wrap()` can take an `optionalInfo` param to inject additional context into the error stack

```go
func Foo(bar string) error {
    const op = "Foo"

    err := db.GetBar(bar) // "GetBar: [database_error] cannot find bar"
    if err != nil {
        return e.Wrap(op, err, fmt.Sprintf("bar id: %s", bar.Id))
        // "Foo: (bar id: 2hs8qh9): GetBar: [database_error] cannot find bar"
    }
    return nil
}
```

### Wrapping a different error type

`Wrap()` can be chained with `SetCode()` to provide a new code for errors from another package (or default go errors)

```go
// db.GetBar returns sentinel error:
//     var ErrNotFound = errors.New("not found")

const CodeInternalError = "internal_error"

func Foo(bar string) error {
    const op = "Foo"

    err := db.GetBar(bar) // "not found"
    if err != nil {
        return e.Wrap(op, err).SetCode(CodeInternalError)
        // "Foo: [internal_error] not found
    }
}
```

## Handling Errors

### End-user

`ErrorMessage()` is used to display a user-friendly error message to the end-user. `New()` and `Wrap()` do not have a `message` param (intentional design). `SetMessage()` should be called to assign an intentional, meaningful message.

```go
func doSomething(r *http.Request) error {
    const op = "doSomething"

    err := Foo()
    if err != nil {
        return e.Wrap(op, err).SetMessage("Oh no! An Error has occurred.")
        // this message will not show in Error() and can only be retrieved
        // by e.Message() or ErrorMessage()
    }
    return nil
}
```

This behaviour should be combined with a default error message at the handler level to make `message` an optional field that should only be set to give additional context to the end-user.

```go
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    bytes, err := doSomething(r)
    if err != nil {
        // Log full error stack
        logger.Error(err)

        // Extract the first error message from the stack
        userMsg := e.ErrorMessage(err)
        if userMsg == "" {
            userMsg = "Unexpected error has occurred"
        }

        http.Error(w, userMsg, http.StatusInternalServerError)
        return
    }
    w.Write(bytes)
}
```

### Client

`ErrorCode()` is intended for use by any clients such as front-end applications, other libraries, and even callers within your own application. In the context of a go codebase, `code` provides an alternative way of introspecting error types without comparing `Error()` strings or using type assertions.

```go
const (
    CodeInvalidError = "invalid_error"
    CodeDatabaseError = "database_error"
)

func Foo(bar string) error {
    const op = "Foo"

    if bar == nil {
        return e.New(op, CodeInvalidError, "bar cannot be nil")
    }
    err := db.GetBar(bar)
    if err != nil {
        return e.Wrap(op, err).SetCode(CodeDatabaseError)
        // SetCode() may not be necessary if we control the err
        // returned from db.GetBar().
    }
    return nil
}

func doSomething(r *http.Request) error {
    const op = "doSomething"

    err := Foo(r.FormValue("bar"))
    if err != nil {
        var info string
        switch e.ErrorCode(err) {
        case CodeInvalidError:
            info = "Invalid Request"
        case CodeDatabaseError:
            err2 := RetryFoo(r.FormValue("bar"))
            if err2 != nil {
                info = "Database error has occurred. Please try again."
            } else {
                return nil
            }
        default:
            info = "Unexpected error has occurred. Please try again"
        }
        return e.Wrap(op, err).SetMessage(info)
    }
    return nil
}
```

### Operator

Operators are usually the developers or application support who are concerned with logging the logical stack trace of the error. By using the `const op = "FuncName"` pattern we can easily build a chain of functions that were called down the stack to the error site.

Logging `Error()` is sufficient to write the stack trace in this format:

```
op: (additionalInfo): [code] cause
```

Examples:

```
Foo: cannot find bar

DoSomething: Foo: [database_error] cannot find bar

Handler.ServeHTTP: DoSomething: Foo: (cannot find bar 2hjk7d): GetBar: getBarById: [database_error] cannot find bar
```

## Comparisons with other approaches

### Upspin

Package `e` adopts the `const op = "FuncName"` pattern introduced by Rob Pike and Andrew Gerrard in [Upspin](https://commandcenter.blogspot.com/2017/12/error-handling-in-upspin.html) to build a logical stacktrace.

In place of Upspin's multi-purpose `E(args ...interface{})` function, `e` uses the familiar verbs `New()` and `Wrap()` to provide better type safety and simpler implementation.

Upspin did not have a clear separation between messages for end-users and the error stacktrace, making it unsuitable for a web application which needs to hide internal details.

### Ben Johnson's Failure is your Domain

Package `e` is heavily influenced by Ben's approach to error-handling outlined in his blog post [Failure is your Domain](https://middlemost.com/failure-is-your-domain/). These are some key differences:

- `e` does not rely on struct initialization to create errors and instead uses `New()`, `Wrap()` and other helper methods to guarantee valid internal state

- `e` does not require the error stack to have uniform type, and is compatible with other error types

- `e` keeps `message` logically distinct from the error stacktrace, making it suitable for displaying to an end-user
