# Go panic recovery

Helpers for catching panics, including in goroutines.
Unify panics as an `error`, but retain the ability to distinguish it as a `PanicError`.

# Usage

## Go routine panic recovery

recovery.Go helps handle goroutine panics and errors.
Note that the function given to recovery.Go returns an error, whereas a normal go routine does not.

``` go
// This will crash your program
go func() {
	panic("panic")
}

// This will not crash your program
// By default it logs the panic
go recovery.Go(func() error {
	panic("panic")
})
```

Panics are given to the error handling function as a `PanicError`.
The global error handling function can be set with the variable recovery.ErrorHandler or can be set locally by using GoHandler

```
errHandler := func(err error) {
	sentry.CaptureException(err)
}

// global
recovery.ErrorHandler = errHandler

// local
recovery.GoHandler(errHandler, func() error {
		panic("panic")
})
```

## Function panic recovery

Recover a panic for a normal function call (not a goroutine) and return the error.

``` go
err = recovery.Call(func() error {
	panic("panic")
})
```

There are also variants that allow the called function to return results in addition to an error: `Call1`, `Call2`, `Call3`.
There are two helpers `Throw` and `Throwf` for panicing an error that will avoid wrapping it as a `PanicError`.

# Codemod

There are comby rules to help with upgrading existing go routines:

	comby -config ../go-recovery/codemod/comby/upgrade.toml -f myfile.go
	goimports -w myfile.go
	gofmt -w myfile.go

Existing `return` statements in go routines should be manually changed to `return nil`.

The upgrades are conservative and will insert function wrapping that is sometimes unnecesssary.
In a go statement the arguments to a function, including the method receiver are evaluated immediately.
In `go fn(x)`, x is evaluated immediately. The function wrapping with `go func(x){}(x)` maintains the property of immediate evaluation.
If the variable is immutable or copied before changed, then immediate evaluation is unnecessary, but you will need to inspect this manually.

