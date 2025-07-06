package gorpc

import (
	"fmt"
	"strconv"
)

// Error represents a goRPC server error.
type Error struct {
	Text     string
	innerErr error
	Code     int
}

// NewError creates a new goRPC error.
func NewError(code int, text string, errs ...error) *Error {
	var err error

	for _, e := range errs {
		if err == nil {
			err = e
			continue
		}

		err = fmt.Errorf("%w: %w", err, e)
	}

	return &Error{
		Code:     code,
		Text:     text,
		innerErr: err,
	}
}

// Error returns the inner error as a string in format: {Code} {Text}: {error}
func (e *Error) Error() string {
	msg := strconv.Itoa(e.Code) + " " + e.Text

	if e.innerErr == nil {
		msg += ": " + e.innerErr.Error()
	}

	return msg
}

// Unwrap returns the inner error.
func (e *Error) Unwrap() error {
	return e.innerErr
}
