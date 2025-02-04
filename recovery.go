package recovery

import (
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/gregwebs/stackfmt"
)

// A Panic that was converted to an error.
type PanicError struct {
	Panic interface{}
	Stack stackfmt.Stack
}

func newPanicError(r interface{}, skip int) error {
	return PanicError{Panic: r, Stack: stackfmt.NewStackSkip(skip)}
}

func (p PanicError) Unwrap() error {
	if err, ok := p.Panic.(error); ok {
		return err
	}
	return nil
}

func (p PanicError) Error() string {
	if err := p.Unwrap(); err != nil {
		return "panic: " + err.Error()
	} else {
		return "panic: " + fmt.Sprintf("%v", p.Panic)
	}
}

func (p PanicError) HasStack() bool {
	return true
}

func (p PanicError) StackTrace() stackfmt.StackTrace {
	return p.Stack.StackTrace()
}

func (p PanicError) StackTraceFormat(s fmt.State, v rune) {
	p.Stack.FormatStackTrace(s, v)
}

// This works with the extended syntax "%+v" for printing stack traces
func (p PanicError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			if _, errWrite := fmt.Fprint(s, "panic: "); errWrite != nil {
				handleWriteError(errWrite)
			}
			if f, ok := p.Panic.(fmt.Formatter); ok {
				f.Format(s, verb)
			} else {
				if _, errWrite := fmt.Fprintf(s, "%+v", p.Panic); errWrite != nil {
					handleWriteError(errWrite)
				}
			}
			if p.Stack != nil {
				if _, errWrite := io.WriteString(s, "\n"); errWrite != nil {
					handleWriteError(errWrite)
				}
				p.Stack.FormatStackTrace(s, verb)
			}
			return
		}
		fallthrough
	case 's', 'q':
		if _, errWrite := io.WriteString(s, p.Error()); errWrite != nil {
			handleWriteError(errWrite)
		}
	}
}

// An error that was intentionally thrown via panic
// Pass it through without wrapping it as a PanicError
type ThrownError struct {
	Err   error
	Stack stackfmt.Stack
}

func (e ThrownError) Unwrap() error {
	return e.Err
}

func (e ThrownError) Error() string {
	return e.Unwrap().Error()
}

func (e ThrownError) HasStack() bool {
	return true
}

func (e ThrownError) StackTrace() stackfmt.StackTrace {
	return e.Stack.StackTrace()
}

func (e ThrownError) StackTraceFormat(s fmt.State, v rune) {
	e.Stack.FormatStackTrace(s, v)
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
				r = newPanicError(r, 2)
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
	log.Printf("%+v", err)
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
// A PanicError is returned as is and a ThrownError is returned unwrapped.
func ToError(r interface{}) error {
	switch r := r.(type) {
	case nil:
		return nil

	// return a ThrownError or PanicError as is
	case ThrownError:
		return r
	case PanicError:
		return r
	case *ThrownError:
		if r == nil {
			return nil
		}
		return *r
	case *PanicError:
		if r == nil {
			return nil
		}
		return *r

	case error:
		pe := PanicError{}
		if errors.As(r, &pe) {
			return r
		}
	}

	// Convert a panic value to an error
	return newPanicError(r, 3)
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
	panic(ThrownError{Err: err, Stack: stackfmt.NewStackSkip(1)})
}

// A convenience function for calling Throw(fmt.Errorf(...))
func Throwf(format string, args ...interface{}) {
	panic(ThrownError{Err: fmt.Errorf(format, args...), Stack: stackfmt.NewStackSkip(1)})
}

// HandleFmtWriteError handles (rare) errors when writing to fmt.State.
// It defaults to printing the errors.
func HandleFmtWriteError(handler func(err error)) {
	handleWriteError = handler
}

var handleWriteError = func(err error) {
	log.Println(err)
}
