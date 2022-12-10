package recovery

import "fmt"

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

// RecoveredCall is a helper function which allows you to easily recover from panics in the given function parameter "fn".
// If fn returns an error, that will be returned.
// If a panic occurs, RecoveredCall will convert it to a PanicError and return it.
func RecoveredCall(fn func() error) (err error) {
	// the returned variable distinguishes the case of panic(nil)
	returned := false
	defer func() {
		r := recover()
		if !returned && err == nil {
			// the case of panic(nil)
			if r == nil {
				r = PanicError{Panic: r}
			}
			err = RecoverToError(r)
		}
	}()
	result := fn()
	returned = true
	return result
}

// Same as RecoveredCall but support returning 1 result in addition to the error.
func RecoveredCall1[T any](fn func() (T, error)) (T, error) {
	var t T
	return t, RecoveredCall(func() error {
		var err error
		t, err = fn()
		return err
	})
}

// Same as RecoveredCall but support returning 2 results in addition to the error.
func RecoveredCall2[T any, U any](fn func() (T, U, error)) (T, U, error) {
	var t T
	var u U
	return t, u, RecoveredCall(func() error {
		var err error
		t, u, err = fn()
		return err
	})
}

// Same as RecoveredCall but support returning 3 results in addition to the error.
func RecoveredCall3[T any, U any, V any](fn func() (T, U, V, error)) (T, U, V, error) {
	var t T
	var u U
	var v V
	return t, u, v, RecoveredCall(func() error {
		var err error
		t, u, v, err = fn()
		return err
	})
}

// GoRecovered allows you to easily handle panics when spawning go routines.
// Instead of your program crashing, the panic is converted to a PanicError and given to the errorHandler
func GoRecovered(errorHandler func(err error), fn func() error) {
	go func() {
		e := RecoveredCall(func() error {
			return fn()
		})
		if e != nil {
			errorHandler(e)
		}
	}()
}

// Wrap panic values in a PanicError.
// nil is returned as nil so this function can be called direclty with the result of recover()
// A ThrownError or a PanicError are returned as is.
func RecoverToError(r interface{}) error {
	switch r := r.(type) {
	// A Go panic
	case PanicError:
		return r
	case ThrownError:
		return r
	case nil:
		return nil
	default:
		return PanicError{Panic: r}
	}
}
