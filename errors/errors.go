package errors

import (
	"fmt"

	"github.com/pkg/errors"
)

type withCause struct {
	stackedError error

	cause error
}

func (e *withCause) Error() string {
	return e.stackedError.Error()
}

func (e *withCause) Cause() error {
	return e.cause
}

func (e *withCause) Unwrap() error {
	return e.cause
}

func Is(err error, target error) bool {
	wc, ok := err.(*withCause)
	tc, ok2 := target.(*withCause)

	if ok && ok2 {
		return errors.Is(wc.cause, tc.cause)
	} else if ok {
		return errors.Is(wc.cause, target)
	} else if ok2 {
		return errors.Is(err, tc.cause)
	} else {
		return errors.Is(err, target)
	}
}

type ErrorMapper func(err error) error

func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}

	// try to cast to *error
	er, ok := err.(*withCause)

	// if the current error is not a wrapped error, wrap it and set it as cause
	if !ok {
		pkgErr := errors.Wrap(err, message)
		return &withCause{
			stackedError: pkgErr,
			cause:        err,
		}
	}

	// otherwise wrap the current error and set it's cause as cause
	pkgErr := errors.Wrap(err, message)
	return &withCause{
		stackedError: pkgErr,
		cause:        er.cause,
	}
}

func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	return Wrap(err, fmt.Sprintf(format, args...))
}

func WrapWithMapper(err error, mapper ErrorMapper, format string) error {
	if err == nil {
		return nil
	}

	mappedErr := mapper(err)

	return Wrap(mappedErr, format)
}

func WrapfWithMapper(err error, mapper ErrorMapper, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	return WrapWithMapper(err, mapper, fmt.Sprintf(format, args...))
}

func Errorf(format string, args ...interface{}) error {
	err := errors.Errorf(format, args...)

	return &withCause{
		stackedError: err,
		cause:        err,
	}
}

func New(message string) error {
	err := errors.New(message)
	return &withCause{
		stackedError: err,
		cause:        err,
	}
}

