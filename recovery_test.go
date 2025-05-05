package recovery_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/gregwebs/go-recovery"
	"github.com/shoenig/test/must"
)

func TestCallNil(t *testing.T) {
	// return nil- no error
	err := recovery.Call(func() error {
		return nil
	})
	must.Nil(t, err)
}

func HasStack(err error) bool {
	if errWithStack, ok := err.(interface{ HasStack() bool }); ok {
		return errWithStack.HasStack()
	}
	return false
}

func TestCallError(t *testing.T) {
	var errOrig error
	// return a basic error
	err := recovery.Call(func() error {
		errOrig = fmt.Errorf("return error")
		return errOrig
	})
	must.NotNil(t, err)
	must.Eq(t, errOrig, err)
}

func TestCallPanicValue(t *testing.T) {
	// panic string
	err := recovery.Call(func() error {
		panic("panic")
	})
	must.NotNil(t, err)
	must.True(t, HasStack(err))
	must.Eq(t, "panic: panic", err.Error())

	// panic nil
	err = recovery.Call(func() error {
		panic(nil)
	})
	must.NotNil(t, err)
	must.True(t, HasStack(err))
	must.Eq(t, "panic: panic called with nil argument", err.Error())
}

var standardErr = fmt.Errorf("error standard")

func TestCallPanicError(t *testing.T) {
	// panic standard error
	err := recovery.Call(func() error {
		panic(standardErr)
	})
	must.NotNil(t, err)
	// There's no direct equivalent to test.Is in must, use type assertion
	_, ok := err.(recovery.PanicError)
	must.True(t, ok)
	must.True(t, HasStack(err))
	must.Eq(t, "panic: error standard", err.Error())

	// panic error
	err = recovery.Call(func() error {
		panic(errors.New("error with stack"))
	})
	_, ok = err.(recovery.PanicError)
	must.True(t, ok)
	must.NotNil(t, err)
	must.True(t, HasStack(err))
	must.Eq(t, "panic: error with stack", err.Error())
	fullPrint := fmt.Sprintf("%+v", err)
	must.StrContains(t, fullPrint, "recovery_test.go")
}

func TestCallThrown(t *testing.T) {
	thrown := fmt.Errorf("thrown error")
	err := recovery.Call(func() error {
		return thrown
	})
	must.NotNil(t, err)
	must.Eq(t, thrown, err)
	must.Eq(t, "thrown error", err.Error())
	err = recovery.Call(func() error {
		recovery.Throw(thrown)
		return nil
	})
	must.NotNil(t, err)
	must.Eq(t, thrown, errors.Unwrap(err))
	must.Eq(t, "thrown error", err.Error())

	err = recovery.Call(func() error {
		panic("panic")
	})
	must.NotNil(t, err)
	_, ok := err.(recovery.PanicError)
	must.True(t, ok)
	must.Eq(t, "panic: panic", err.Error())
}

func TestGoHandler(t *testing.T) {
	noError := func(err error) {
		must.Nil(t, err)
	}
	errHappened := func(err error) {
		must.NotNil(t, err)
	}
	recovery.GoHandler(noError, func() error {
		return nil
	})
	recovery.GoHandler(errHappened, func() error {
		panic("panic")
	})

	wait := make(chan struct{})
	go recovery.GoHandler(noError, func() error {
		wait <- struct{}{}
		return nil
	})
	go recovery.GoHandler(errHappened, func() error {
		defer func() { wait <- struct{}{} }()
		panic("panic")
	})
	<-wait
	<-wait
}

func TestGo(t *testing.T) {
	noError := func(err error) {
		must.Nil(t, err)
	}
	errHappened := func(err error) {
		must.NotNil(t, err)
	}

	recovery.ErrorHandler = noError
	recovery.Go(func() error {
		return nil
	})
	recovery.ErrorHandler = errHappened
	recovery.Go(func() error {
		panic("panic")
	})

	wait := make(chan struct{})
	recovery.ErrorHandler = noError
	go recovery.Go(func() error {
		wait <- struct{}{}
		return nil
	})
	<-wait
	recovery.ErrorHandler = errHappened
	go recovery.Go(func() error {
		defer func() { wait <- struct{}{} }()
		panic("panic")
	})
	<-wait
}
