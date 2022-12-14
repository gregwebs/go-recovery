package recovery_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/gregwebs/go-recovery"
	"github.com/stretchr/testify/assert"
)

func TestCall(t *testing.T) {
	err := recovery.Call(func() error {
		return nil
	})
	assert.Nil(t, err)
	err = recovery.Call(func() error {
		return fmt.Errorf("return error")
	})
	assert.NotNil(t, err)
	err = recovery.Call(func() error {
		panic("panic")
	})
	assert.NotNil(t, err)
	assert.Equal(t, "panic", err.Error())
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
	assert.Equal(t, recovery.PanicError{Panic: "panic"}, errors.Unwrap(err))
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
