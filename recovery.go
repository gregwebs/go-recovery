package recovery

import (
	"fmt"
	"log"
)

// A Panic that was converted to an error.
type PanicError struct {
	Panic any
}

func (p PanicError) Unwrap() error {
	if err, ok := p.Panic.(error); ok {
		return err
	}
	return nil
}

func (p PanicError) Error() string {
	if err := p.Unwrap(); err != nil {
		return err.Error()
	} else {
		return fmt.Sprintf("%v", p.Panic)
	}
}

// An error that was intentionally thrown via panic
// Pass it through without wrapping it as a PanicError
type ThrownError struct {
	Err error
}

func (e ThrownError) Unwrap() error {
	return e.Err
}

func (e ThrownError) Error() string {
	return e.Unwrap().Error()
}

// Call is a helper function which allows you to easily recover from panics in the given function parameter "fn".
// If fn returns an error, that will be returned.
// If a panic occurs, Call will convert it to a PanicError and return it.
func Call(fn func() error) (err error) {
	// the returned variable distinguishes the case of panic(nil)
	returned := false
	defer func() {
		r := recover()
		if !returned && err == nil {
			// the case of panic(nil)
			if r == nil {
				r = PanicError{Panic: r}
			}
			err = ToError(r)
		}
	}()
	result := fn()
	returned = true
	return result
}

// Same as Call but support returning 1 result in addition to the error.
func Call1[T any](fn func() (T, error)) (T, error) {
	var t T
	return t, Call(func() error {
		var err error
		t, err = fn()
		return err
	})
}

// Same as Call but support returning 2 results in addition to the error.
func Call2[T any, U any](fn func() (T, U, error)) (T, U, error) {
	var t T
	var u U
	return t, u, Call(func() error {
		var err error
		t, u, err = fn()
		return err
	})
}

// Same as Call but support returning 3 results in addition to the error.
func Call3[T any, U any, V any](fn func() (T, U, V, error)) (T, U, V, error) {
	var t T
	var u U
	var v V
	return t, u, v, Call(func() error {
		var err error
		t, u, v, err = fn()
		return err
	})
}

// The default ErrorHandler is DefaultErrorHandler
var ErrorHandler func(error) = DefaultErrorHandler

// The DefaultErrorHandler prints the error with log.Printf
func DefaultErrorHandler(err error) {
	log.Printf("go routine error: %+v", err)
}

// Go is designed to use as an entry point to a go routine.
//
//	go recovery.Go(func() error { ... })
//
// Instead of your program crashing, the panic is converted to a PanicError.
// The panic or a returned error is given to the global ErrorHandler function.
// Change the behavior globally by setting the ErrorHandler package variable.
// Or use GoHandler to set the error handler on a local basis.
func Go(fn func() error) {
	GoHandler(ErrorHandler, fn)
}

// GoHandler is designed to be used when creating go routines to handle panics and errors.
// Instead of your program crashing, a panic is converted to a PanicError.
// The panic or a returned error is given to the errorHandler function.
//
//	go GoHandler(handler, func() error { ... })
func GoHandler(errorHandler func(err error), fn func() error) {
	e := Call(func() error {
		return fn()
	})
	if e != nil {
		errorHandler(e)
	}
}

// Wrap panic values in a PanicError.
// nil is returned as nil so this function can be called direclty with the result of recover()
// A ThrownError or a PanicError are returned as is.
func ToError(r interface{}) error {
	switch r := r.(type) {
	// A Go panic
	case PanicError:
		return r
	case ThrownError:
		return r.Err
	case nil:
		return nil
	default:
		return PanicError{Panic: r}
	}
}

// Throw will panic an error as a ThrownError.
// The error will be returned by Call* functions as a normal error rather than a PanicError.
// Useful with CallX functions to avoid writing out zero values when prototyping.
//
//	func (x int) (int, error) {
//		recovery.Call1(func() (int, error) {
//			recovery.Throw(err)
//		}
//	}
func Throw(err error) {
	panic(ThrownError{Err: err})
}

// A convenience function for calling Throw(fmt.Errorf(...))
func Throwf(format string, args ...interface{}) {
	Throw(fmt.Errorf(format, args...))
}
