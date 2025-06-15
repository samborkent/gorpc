package goc

import "errors"

var (
	ErrInvalidValue  = errors.New("invalid value encountered")
	ErrTypeAssertion = errors.New("type assertion failed")
)
