package goc

import "errors"

var (
	ErrInvalidValue  = errors.New("invalid value encountered")
	ErrTypeAssertion = errors.New("type assertion failed")
	ErrOverflowInt   = errors.New("decoded int overflows 32-bit int value")
	ErrOverflowUint  = errors.New("decoded uint or uintptr overflows 32-bit uint value")
)
