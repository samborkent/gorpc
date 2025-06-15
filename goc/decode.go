package goc

import (
	"encoding"
	"errors"
	"fmt"
	"io"
	"reflect"
	"unsafe"
)

func decodeConcrete[T any](r io.Reader) (T, error) {
	// TODO: avoid allocation
	d := make([]byte, unsafe.Sizeof(*new(T)))

	n, err := r.Read(d)
	if n == 0 && err != nil && !errors.Is(err, io.EOF) {
		return *new(T), err
	}

	return *(*T)(unsafe.Pointer(&d[0])), nil
}

func decodeDecodeReader[T any](r io.Reader, decoder DecodeReader) (T, error) {
	if err := decoder.DecodeFrom(r); err != nil {
		return *new(T), fmt.Errorf("DecodeFrom: %w", err)
	}

	decoded, ok := decoder.(T)
	if !ok {
		return *new(T), errors.New("unable to cast DecodeReader to concrete type")
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
		return *new(T), errors.New("unable to cast Decoder to concrete type")
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
		return *new(T), errors.New("unable to cast BinaryUnmarshaler to concrete type")
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
