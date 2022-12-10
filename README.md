# Go panic recovery

Helpers for catching panics, including in goroutines.
Unify panics as an `error`, but retain the ability to distinguish it as a `PanicError`.

# Go routine panic recovery

recovery.Go launches a goroutine.
Note that the function given to recovery.Go returns an error, whereas a normal go routine does not.
Panics are returned to the error handling function as a `PanicError`.

``` go
errHandler := func(err error) {
        switch r := err.(type) {
	// A Go panic
	case PanicError:
		log.Printf("go routine panic %v", r.Panic)
	default:
		log.Printf("go routine error %+v", err)
	}
}

recovery.Go(errHandler, func() error {
		panic("panic")
		return nil
})
```

# Function panic recovery

Recover a panic for a normal function, no go routines.

``` go
err = recovery.Call(func() error {
		panic("panic")
})
```

There are also variants that allow the called function to return results in addition to an error: Call1, Call2, Call3.
