package driver

import (
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type RetryableError interface {
	Retryable()
}

type RetryableErrorImpl struct {
	Err error
}

func (RetryableErrorImpl) Retryable()      {}
func (e RetryableErrorImpl) Error() string { return e.Err.Error() }

type Retrier interface {
	Retry(func() error) error
	RetryComplex(func() error, int, time.Duration) error
}

type RetrierImpl struct{}

func (r RetrierImpl) Retry(actionFunc func() error) error {
	return r.RetryComplex(actionFunc, 30, 2*time.Second)
}

func (RetrierImpl) RetryComplex(actionFunc func() error, times int, sleep time.Duration) error {
	var lastErr error

	for i := 0; i < times; i++ {
		lastErr = actionFunc()
		if lastErr == nil {
			return nil
		}

		if _, ok := lastErr.(RetryableError); !ok {
			return bosherr.WrapError(lastErr, "Encountered non-retryable error")
		}

		time.Sleep(sleep)
	}

	return bosherr.WrapErrorf(lastErr, "Retried '%d' times", times)
}
