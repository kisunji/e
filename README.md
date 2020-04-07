# kisunji/e

![Go](https://github.com/kisunji/e/workflows/Go/badge.svg)

Inspired by [Failure is your Domain](https://middlemost.com/failure-is-your-domain/) and [Error handling in upspin](https://commandcenter.blogspot.com/2017/12/error-handling-in-upspin.html) (comparison [here](#comparisons-with-other-approaches)), `kisunji/e` is designed to meet the needs of web applications by maintaining a clean separation between the error data consumed by end-users, clients, and operators. 

`kisunji/e` focuses on three principles:


* **Simplicity**: Small, fluent interface for developer ergonomics

* **Safety**: Explicitly separate end-user messages from internal errors

* **Compatibility**: Follows error-wrapping conventions and can be adopted incrementally


## Usage

### Creating a new Error
```go
const CodeInvalidError = "invalid_error"

func Foo(bar string) error {
    const op = "Foo"

    if bar == nil {
        return e.New(op, CodeDbError, "bar cannot be nil")
        // "Foo: [invalid_error] bar cannot be nil"
    }
    return nil
} 
```

### Wrapping an existing Error
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

work in progress!

## Comparisons with other approaches

### Upspin

`kisunji/e` adopts the `const op = "funcName"` paradigm introduced by Rob Pike in [Upspin](https://commandcenter.blogspot.com/2017/12/error-handling-in-upspin.html) to build a logical stacktrace.

In place of Upspin's multi-purpose `func E(args ...interface{}) error`, `kisunji/e` uses the familiar verbs `New()` and `Wrap()` to provide better type safety and simpler implementation.

Upspin did not have a clear separation between messages for end-users and the error stacktrace, making it unsuitable for a web application which needs to hide internal details.

### Ben Johnson's Failure is your Domain

`kisunji/e` is heavily influenced by Ben's approach to error-handling outlined in his blog post [Failure is your Domain](https://middlemost.com/failure-is-your-domain/). These are some key differences:

* `kisunji/e` does not rely on struct initialization to create errors and instead uses `New()`, `Wrap()` and other helper methods to guarantee valid internal state
  
* `kisunji/e` does not require the error stack to have uniform type, and is compatible with other error types

* `kisunji/e` keeps `message` logically distinct from the error stacktrace, making it suitable for displaying to an end-user