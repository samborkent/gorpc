package unsafecast

import (
	"unsafe"
)

func BoolByte(v bool) byte {
	return *(*byte)(unsafe.Pointer(&v))
}

type FixedSize interface {
	~uint16 | ~int16 | ~uint32 | ~int32 | ~float32 | ~uint64 | ~int64 | ~float64 | ~complex64 | ~complex128
}

func FixedToBytes[T FixedSize](v T) []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(&v)), unsafe.Sizeof(v))
}

func FixedFromBytes[T FixedSize](v []byte) T {
	if len(v) != int(unsafe.Sizeof(*new(T))) {
		panic("FromBytes: number of bytes does not map to given type")
	}

	return *(*T)(unsafe.Pointer(unsafe.SliceData(v)))
}
