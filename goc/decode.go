package goc

import (
	"encoding"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
)

func decodeConcrete[T any](r io.Reader) (T, error) {
	var zero T

	// TODO: avoid allocation
	d := make([]byte, reflect.TypeFor[T]().Size())

	n, err := r.Read(d)
	if n == 0 && err != nil && !errors.Is(err, io.EOF) {
		return zero, err
	}

	// TODO: add int, uint, uintptr?
	switch any(zero).(type) {
	case bool:
		return castGeneric[T](decodeBool(d[0]))
	case int8:
		return castGeneric[T](int8(d[0]))
	case int16:
		return castGeneric[T](decodeInt16(d))
	case int32:
		return castGeneric[T](decodeInt32(d))
	case int64:
		return castGeneric[T](decodeInt64(d))
	case uint8:
		return castGeneric[T](d[0])
	case uint16:
		return castGeneric[T](decodeUint16(d))
	case uint32:
		return castGeneric[T](decodeUint32(d))
	case uint64:
		return castGeneric[T](decodeUint64(d))
	case float32:
		return castGeneric[T](decodeFloat32(d))
	case float64:
		return castGeneric[T](decodeFloat64(d))
	case complex64:
		return castGeneric[T](decodeComplex64(d))
	case complex128:
		return castGeneric[T](decodeComplex128(d))
	default:
		panic("decodeConcrete received a non-concrete type: " + reflect.TypeFor[T]().String())
	}
}

func decodeBool(b byte) bool {
	return b != 0
}

func decodeInt16(b []byte) int16 {
	return int16(decodeUint16(b))
}

func decodeUint16(b []byte) uint16 {
	return binary.LittleEndian.Uint16(b[:2])
}

func decodeInt32(b []byte) int32 {
	return int32(decodeUint32(b))
}

func decodeUint32(b []byte) uint32 {
	return binary.LittleEndian.Uint32(b[:4])
}

func decodeInt64(b []byte) int64 {
	return int64(decodeUint64(b))
}

func decodeUint64(b []byte) uint64 {
	return binary.LittleEndian.Uint64(b[:8])
}

func decodeFloat32(b []byte) float32 {
	return math.Float32frombits(decodeUint32(b))
}

func decodeFloat64(b []byte) float64 {
	return math.Float64frombits(decodeUint64(b))
}

func decodeComplex64(b []byte) complex64 {
	r := decodeFloat32(b[:4])
	i := decodeFloat32(b[4:8])
	return complex(r, i)
}

func decodeComplex128(b []byte) complex128 {
	r := decodeFloat64(b[:8])
	i := decodeFloat64(b[8:16])
	return complex(r, i)
}

func decodeDecodeReader[T any](r io.Reader, decoder DecodeReader) (T, error) {
	if err := decoder.DecodeFrom(r); err != nil {
		return *new(T), fmt.Errorf("DecodeFrom: %w", err)
	}

	decoded, ok := decoder.(T)
	if !ok {
		return *new(T), errors.New("unable to encode DecodeReader to concrete type")
	}

	return decoded, nil
}

func decodeDecoder[T any](r io.Reader, decoder Decoder) (T, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return *new(T), fmt.Errorf("ReadAll: %w", err)
	}

	if err := decoder.Decode(b); err != nil {
		return *new(T), fmt.Errorf("Decode: %w", err)
	}

	decoded, ok := decoder.(T)
	if !ok {
		return *new(T), errors.New("unable to encode Decoder to concrete type")
	}

	return decoded, nil
}

func decodeBinaryUnmarshaler[T any](r io.Reader, decoder encoding.BinaryUnmarshaler) (T, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return *new(T), fmt.Errorf("ReadAll: %w", err)
	}

	if err := decoder.UnmarshalBinary(b); err != nil {
		return *new(T), fmt.Errorf("UnmarshalBinary: %w", err)
	}

	decoded, ok := decoder.(T)
	if !ok {
		return *new(T), errors.New("unable to encode BinaryUnmarshaler to concrete type")
	}

	return decoded, nil
}

func decodeValueDecodeReader(r io.Reader, v reflect.Value) error {
	decodeReader, ok := reflect.TypeAssert[DecodeReader](v)
	if !ok {
		return ErrTypeAssertion
	}

	if err := decodeReader.DecodeFrom(r); err != nil {
		return fmt.Errorf("DecodeFrom: %w", err)
	}

	return nil
}

func decodeValueDecoder(r io.Reader, v reflect.Value) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("ReadAll: %w", err)
	}

	decoder, ok := reflect.TypeAssert[Decoder](v)
	if !ok {
		return ErrTypeAssertion
	}

	if err := decoder.Decode(b); err != nil {
		return fmt.Errorf("Decode: %w", err)
	}

	return nil
}

func decodeValueBinaryUnmarshaler(r io.Reader, v reflect.Value) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("ReadAll: %w", err)
	}

	binaryUnmarshaler, ok := reflect.TypeAssert[encoding.BinaryUnmarshaler](v)
	if !ok {
		return ErrTypeAssertion
	}

	if err := binaryUnmarshaler.UnmarshalBinary(b); err != nil {
		return fmt.Errorf("UnmarshalBinary: %w", err)
	}

	return nil
}
