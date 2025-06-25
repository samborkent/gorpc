package goc

import (
	"math"
	"unsafe"

	"github.com/samborkent/gorpc/goc/internal/unsafecast"
)

func unsafeEncodeBool(v bool) byte {
	return unsafecast.BoolByte(v)
}

func unsafeEncodeInt16(v int16) []byte {
	return unsafeEncodeUint16(uint16(v))
}

func unsafeEncodeUint16(v uint16) []byte {
	return unsafecast.FixedToBytes(v)
}

func unsafeEncodeInt32(v int32) []byte {
	return unsafeEncodeUint32(uint32(v))
}

func unsafeEncodeUint32(v uint32) []byte {
	return unsafecast.FixedToBytes(v)
}

func unsafeEncodeInt64(v int64) []byte {
	return unsafeEncodeUint64(uint64(v))
}

func unsafeEncodeUint64(v uint64) []byte {
	return unsafecast.FixedToBytes(v)
}

func unsafeEncodeFloat32(v float32) []byte {
	return encodeUint32(math.Float32bits(v))
}

func unsafeEncodeFloat64(v float64) []byte {
	return encodeUint64(math.Float64bits(v))
}

func unsafeEncodeComplex64(v complex64) []byte {
	return unsafecast.FixedToBytes(v)
}

func unsafeEncodeComplex128(v complex128) []byte {
	return unsafecast.FixedToBytes(v)
}

func unsafeEncodeString(v string) []byte {
	return unsafe.Slice(unsafe.StringData(v), len(v))
}
