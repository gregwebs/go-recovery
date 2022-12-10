package recovery

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

// RecoveredCall is a helper function which allows you to easily recover from panics in the given function parameter "fn".
// If a panic occurs, RecoveredCall will convert it to an error and return it.
// If fn returns an error, that will also be returned
func RecoveredCall(fn func() error) (err error) {
	// the returned variable handles the case of panic(nil)
	returned := false
	defer func() {
		r := recover()
		if !returned && err == nil {
			// the case of panic(nil)
			if r == nil {
				r = PanicError{ Pancic: r }
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


// Convert panics to PanicError. A non-runtime error will be returned as is
// Please Note: nil is returned as nil
func RecoverToError(r interface{}) error {
        switch r := r.(type) {
	// A Go panic
        case runtime.Error:
		return PanicError{ Panic: r }
        case error:
		return r
	case nil:
		return nil
	default:
                return PanicError{ Panic: r }
        }
}

func CatchHandlePanic(errorHandler func(error), panicHandler func(v any)) {
        r := recover()
        if r == nil {
                return
        }
        if err := ErrorFromRecover(r); err != nil {
                errorHandler(err)
        } else {
                if panicHandler == nil {
                        panic(r)
                } else {
                        panicHandler(r)
                }
        }
}
