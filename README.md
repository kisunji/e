# github.com/kisunji/e

![Go](https://github.com/kisunji/e/workflows/Go/badge.svg)

Inspired by [Failure is your Domain](https://middlemost.com/failure-is-your-domain/) and [Error handling in upspin](https://commandcenter.blogspot.com/2017/12/error-handling-in-upspin.html), `e` is designed to meet the needs of web applications by maintaining a clean separation between the error data consumed by end-users, clients, and operators. 

`e` focuses on three principles:


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

### Wrapping an non-pkg error type

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
