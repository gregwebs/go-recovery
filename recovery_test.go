package recovery_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/gregwebs/go-recovery"
	"github.com/stretchr/testify/assert"
)

func TestCallNil(t *testing.T) {
	// return nil- no error
	err := recovery.Call(func() error {
		return nil
	})
	assert.Nil(t, err)
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
	assert.NotNil(t, err)
	assert.Equal(t, errOrig, err)
}

func TestCallPanicValue(t *testing.T) {
	// panic string
	err := recovery.Call(func() error {
		panic("panic")
	})
	assert.NotNil(t, err)
	assert.True(t, HasStack(err))
	assert.Equal(t, "panic: panic", err.Error())

	// panic nil
	err = recovery.Call(func() error {
		panic(nil)
	})
	assert.NotNil(t, err)
	assert.True(t, HasStack(err))
	assert.Equal(t, "panic: panic called with nil argument", err.Error())
}

var standardErr = fmt.Errorf("error standard")

func TestCallPanicError(t *testing.T) {
	// panic standard error
	err := recovery.Call(func() error {
		panic(standardErr)
	})
	assert.NotNil(t, err)
	assert.IsType(t, recovery.PanicError{}, err)
	assert.True(t, HasStack(err))
	assert.Equal(t, "panic: error standard", err.Error())

	// panic error
	err = recovery.Call(func() error {
		panic(errors.New("error with stack"))
	})
	assert.IsType(t, recovery.PanicError{}, err)
	assert.NotNil(t, err)
	assert.True(t, HasStack(err))
	assert.Equal(t, "panic: error with stack", err.Error())
	fullPrint := fmt.Sprintf("%+v", err)
	assert.Contains(t, fullPrint, "recovery_test.go")
}

func TestCallThrown(t *testing.T) {
	thrown := fmt.Errorf("thrown error")
	err := recovery.Call(func() error {
		return thrown
	})
	assert.NotNil(t, err)
	assert.Equal(t, thrown, err)
	assert.Equal(t, "thrown error", err.Error())
	err = recovery.Call(func() error {
		recovery.Throw(thrown)
		return nil
	})
	assert.NotNil(t, err)
	assert.Equal(t, thrown, errors.Unwrap(err))
	assert.Equal(t, "thrown error", err.Error())

	err = recovery.Call(func() error {
		panic("panic")
	})
	assert.NotNil(t, err)
	assert.IsType(t, recovery.PanicError{}, err)
	assert.Equal(t, "panic: panic", err.Error())
}

func TestGoHandler(t *testing.T) {
	noError := func(err error) {
		assert.Nil(t, err)
	}
	errHappened := func(err error) {
		assert.NotNil(t, err)
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
		assert.Nil(t, err)
	}
	errHappened := func(err error) {
		assert.NotNil(t, err)
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
