package recovery_test

import (
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

func TestGo(t *testing.T) {
	noError := func(err error) {
		assert.Nil(t, err)
	}
	errHappened := func(err error) {
		assert.NotNil(t, err)
	}
	recovery.Go(noError, func() error {
		return nil
	})
	recovery.Go(errHappened, func() error {
		panic("panic")
	})

	wait := make(chan struct{})
	go recovery.Go(noError, func() error {
		wait <- struct{}{}
		return nil
	})
	go recovery.Go(errHappened, func() error {
		defer func() { wait <- struct{}{} }()
		panic("panic")
	})
	<-wait
	<-wait
}

func TestCallThrown(t *testing.T) {
	thrown := recovery.ThrownError{Err: fmt.Errorf("thrown error")}
	err := recovery.Call(func() error {
		return thrown
	})
	assert.NotNil(t, err)
	assert.Equal(t, thrown, err)
	assert.Equal(t, "thrown error", err.Error())
	err = recovery.Call(func() error {
		panic(thrown)
	})
	assert.NotNil(t, err)
	assert.Equal(t, thrown, err)
	assert.Equal(t, "thrown error", err.Error())

	err = recovery.Call(func() error {
		panic("panic")
	})
	assert.NotNil(t, err)
	assert.Equal(t, recovery.PanicError{Panic: "panic"}, err)
}
