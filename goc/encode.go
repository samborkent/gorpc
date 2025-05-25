package goc

import (
	"io"
	"unsafe"
)

func encodeConcrete[T any](w io.Writer, v T) error {
	_, err := w.Write(unsafe.Slice((*byte)(unsafe.Pointer(&v)), unsafe.Sizeof(v)))
	if err != nil {
		return err
	}

	return nil
}

type FixedSized interface {
	~bool | ~int8 | ~int16 | ~int32 | ~int64 | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64 | ~complex64 | ~complex128
}

func encodeConcreteSlice[T FixedSized](w io.Writer, src []T) error {
	_, err := w.Write(unsafe.Slice((*byte)(unsafe.Pointer(&src[0])), len(src)*int(unsafe.Sizeof(src[0]))))
	if err != nil {
		return err
	}

	return nil
}
