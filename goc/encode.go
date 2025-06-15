package goc

import (
	"encoding/binary"
	"io"
	"math"
	"reflect"
)

func encodeConcrete[T any](w io.Writer, v T) error {
	// TODO: use sync.Pool
	d := make([]byte, 0, reflect.TypeFor[T]().Size())

	var zero T

	// TODO: add int, uint, uintptr?
	switch t := any(zero).(type) {
	case bool:
		d = append(d, encodeBool(t))
	case int8:
		d = append(d, byte(t))
	case int16:
		d = append(d, encodeInt16(t)...)
	case int32:
		d = append(d, encodeInt32(t)...)
	case int64:
		d = append(d, encodeInt64(t)...)
	case uint8:
		d = append(d, t)
	case uint16:
		d = append(d, encodeUint16(t)...)
	case uint32:
		d = append(d, encodeUint32(t)...)
	case uint64:
		d = append(d, encodeUint64(t)...)
	case float32:
		d = append(d, encodeFloat32(t)...)
	case float64:
		d = append(d, encodeFloat64(t)...)
	case complex64:
		d = append(d, encodeComplex64(t)...)
	case complex128:
		d = append(d, encodeComplex128(t)...)
	default:
		panic("encodeConcrete received a non-concrete type: " + reflect.TypeFor[T]().String())
	}

	_, err := w.Write(d)
	if err != nil {
		return err
	}

	return nil
}

func encodeBool(b bool) byte {
	if b {
		return 1
	} else {
		return 0
	}
}

func encodeInt16(v int16) []byte {
	return encodeUint16(uint16(v))
}

func encodeUint16(v uint16) []byte {
	var b [2]byte
	binary.LittleEndian.PutUint16(b[:], v)
	return b[:]
}

func encodeInt32(v int32) []byte {
	return encodeUint32(uint32(v))
}

func encodeUint32(v uint32) []byte {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], v)
	return b[:]
}

func encodeInt64(v int64) []byte {
	return encodeUint64(uint64(v))
}

func encodeUint64(v uint64) []byte {
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], v)
	return b[:]
}

func encodeFloat32(v float32) []byte {
	return encodeUint32(math.Float32bits(v))
}

func encodeFloat64(v float64) []byte {
	return encodeUint64(math.Float64bits(v))
}

func encodeComplex64(v complex64) []byte {
	// TODO: use sync.Pool
	d := make([]byte, 0, 8)

	d = append(d, encodeFloat32(real(v))...)
	d = append(d, encodeFloat32(imag(v))...)

	return d
}

func encodeComplex128(v complex128) []byte {
	// TODO: use sync.Pool
	d := make([]byte, 0, 8)

	d = append(d, encodeFloat64(real(v))...)
	d = append(d, encodeFloat64(imag(v))...)

	return d
}
