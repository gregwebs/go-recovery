package recovery_test

import (
	"fmt"
	"testing"

	"github.com/gregwebs/recovery"
	"github.com/gregwebs/try/assert"
)

func TestRecoveredCall(t *testing.T) {
	err := recovery.RecoveredCall(func() error {
			return nil
	})
	assert.Nil(t, err)
	err = recovery.RecoveredCall(func() error {
			return fmt.Errorf("return error")
	})
	assert.NotNil(t, err)
	err = recovery.RecoveredCall(func() error {
			panic("panic")
	})
	assert.NotNil(t, err)
	assert.Equal(t, "panic", err.Error())

}

func TestGoRecovered(t *testing.T) {
	noError := func(err error) {
			assert.Nil(t, err)
	}
	errHappened := func(err error) {
			assert.NotNil(t, err)
	}
	recovery.GoRecovered(noError, func() error {
			return nil
	})
	recovery.GoRecovered(errHappened, func() error {
			panic("panic")
	})

	wait := make(chan struct{})
	go recovery.GoRecovered(noError, func() error {
			wait <- struct{}{}
			return nil
	})
	go recovery.GoRecovered(errHappened, func() error {
			defer func() { wait <- struct{}{} }()
			panic("panic")
	})
	<-wait
	<-wait
}
