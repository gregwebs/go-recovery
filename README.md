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
        switch r := err.(type) {
	// A Go panic
	case PanicError:
		log.Printf("go routine panic %v", r.Panic)
	default:
		log.Printf("go routine error %+v", err)
	}
}

recovery.GoHandler(errHandler, func() error {
		panic("panic")
		return nil
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

There are semgrep rules to help with upgrading existing go routines:

	for rule in ../go-recovery/codemod/semgrep/* ; do ; semgrep --config $rule ; done

However, there is an issue with the upgrade rules that could introduce defects.
For non-anonymous function and method calls, it delays the evaluation of the arguments

Additionally, it doesn't upgrade anonymous functions that are called with arguments, e.g. `go func(x int) {}(1)`.
