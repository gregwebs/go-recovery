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
// If a panic occurs, RecoveredCall will convert it to an error and return it.
// If fn returns an error, that will also be returned
func RecoveredCall(fn func() error) (err error) {
	// the returned variable distinguishes the case of panic(nil)
	returned := false
	defer func() {
		r := recover()
		if !returned && err == nil {
			// the case of panic(nil)
			if r == nil {
				r = PanicError{ Panic: r }
			}
			err = RecoverToError(r)
		}
	}()
	result := fn()
	returned = true
	return result
}

// GoRecovered allows you to easily handle panics when using the "go" keyword.
// Instead of your program crashing, give the panic to the errorHandler
// Panics are converted to an error and given to the "errorHandler" function.
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


// Convert panics values to PanicError.
// nil is returned as nil so this can be called direclty with the result of recover()
// A ThrownError or an existing PanicError is returned as is
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
                return PanicError{ Panic: r }
        }
}
