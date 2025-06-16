package goc

import (
	"bytes"
	"encoding"
	"errors"
	"fmt"
	"io"
	"reflect"
)

type DecodeReader interface {
	DecodeFrom(io.Reader) error
}

type Decoder interface {
	Decode([]byte) error
}

func Decode[T any](b []byte) (T, error) {
	return DecodeFrom[T](bytes.NewReader(b))
}

func DecodeFrom[T any](r io.Reader) (T, error) {
	val := new(T)

	// Try to decode through interface implementation.
	switch decoder := any(val).(type) {
	case DecodeReader:
		return decodeDecodeReader[T](r, decoder)
	case Decoder:
		return decodeDecoder[T](r, decoder)
	case encoding.BinaryUnmarshaler:
		return decodeBinaryUnmarshaler[T](r, decoder)
	}

	switch decoder := any(*val).(type) {
	case DecodeReader:
		return decodeDecodeReader[T](r, decoder)
	case Decoder:
		return decodeDecoder[T](r, decoder)
	case encoding.BinaryUnmarshaler:
		return decodeBinaryUnmarshaler[T](r, decoder)
	}

	var zero T

	// Try to decode concrete type.
	switch reflect.TypeOf(val).Kind() {
	case reflect.Bool,
		reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128:
		val, err := decodeConcrete[T](r)
		if err != nil {
			return zero, fmt.Errorf("decoding %s: %w", reflect.TypeOf(zero).String(), err)
		}

		return val, nil
	case reflect.String:
		strBytes, err := io.ReadAll(r)
		if err != nil {
			return zero, fmt.Errorf("reading encoded string: %w", err)
		}

		// TODO: avoid allocation
		return castGeneric[T](string(strBytes))
	}

	// Decode through reflection.
	if err := DecodeValue(r, reflect.ValueOf(val)); err != nil {
		return zero, fmt.Errorf("DecodeValue: %w", err)
	}

	return *val, nil
}

var (
	reflectDecodeReader       = reflect.TypeFor[DecodeReader]()
	reflectDecoder            = reflect.TypeFor[Decoder]()
	reflectBinaryUnmarshaller = reflect.TypeFor[encoding.BinaryUnmarshaler]()
)

func DecodeValue(r io.Reader, v reflect.Value) error {
	if v.CanAddr() {
		if v.Type().Implements(reflectDecodeReader) {
			return decodeValueDecodeReader(r, v)
		} else if v.Type().Implements(reflectDecoder) {
			return decodeValueDecoder(r, v)
		} else if v.Type().Implements(reflectBinaryUnmarshaller) {
			return decodeValueBinaryUnmarshaler(r, v)
		}
	}

	return decodeValue(r, v)
}

func decodeValue(r io.Reader, v reflect.Value) error {
	if !v.IsValid() {
		return ErrInvalidValue
	}

	indirections, err := numIndirections(v.Type())
	if err != nil {
		return err
	}

	for range indirections {
		v = reflect.Indirect(v)
	}

	t := v.Type()

	var d []byte

	// Read concrete type.
	switch t.Kind() {
	case reflect.Bool,
		reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128:
		d = make([]byte, t.Size())

		n, err := r.Read(d)
		if n == 0 && err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("reading %s: %w", t.String(), err)
		}
	}

	switch v.Kind() {
	case reflect.Bool:
		v.SetBool(decodeBool(d[0]))
		return nil
	case reflect.Int:
		// TODO: test
		var header [1]byte

		n, err := r.Read(header[:])
		if n == 0 && err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("reading int header: %w", err)
		}

		switch header[0] {
		case 4:
			d = make([]byte, 4)

			n, err = r.Read(d)
			if n == 0 && err != nil && !errors.Is(err, io.EOF) {
				return fmt.Errorf("reading %s: %w", t.String(), err)
			}

			v.SetInt(int64(decodeInt32(d)))
			return nil
		case 8:
			d = make([]byte, 8)

			n, err = r.Read(d)
			if n == 0 && err != nil && !errors.Is(err, io.EOF) {
				return fmt.Errorf("reading %s: %w", t.String(), err)
			}

			v.SetInt(decodeInt64(d))
			return nil
		default:
			return fmt.Errorf("unknown int size %d encountered", header[0])
		}
	case reflect.Int8:
		v.SetInt(int64(int8(d[0])))
		return nil
	case reflect.Int16:
		v.SetInt(int64(decodeInt16(d)))
		return nil
	case reflect.Int32:
		v.SetInt(int64(decodeInt32(d)))
		return nil
	case reflect.Int64:
		v.SetInt(decodeInt64(d))
		return nil
	case reflect.Uint, reflect.Uintptr:
		// TODO: test
		var header [1]byte

		n, err := r.Read(header[:])
		if n == 0 && err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("reading %s header: %w", t.Kind(), err)
		}

		switch header[0] {
		case 4:
			d = make([]byte, 4)

			n, err = r.Read(d)
			if n == 0 && err != nil && !errors.Is(err, io.EOF) {
				return fmt.Errorf("reading %s: %w", t.String(), err)
			}

			v.SetUint(uint64(decodeUint32(d)))
			return nil
		case 8:
			d = make([]byte, 8)

			n, err = r.Read(d)
			if n == 0 && err != nil && !errors.Is(err, io.EOF) {
				return fmt.Errorf("reading %s: %w", t.String(), err)
			}

			v.SetUint(decodeUint64(d))
			return nil
		default:
			return fmt.Errorf("unknown %s size %d encountered", t.Kind(), header[0])
		}
	case reflect.Uint8:
		v.SetUint(uint64(d[0]))
		return nil
	case reflect.Uint16:
		v.SetUint(uint64(decodeUint16(d)))
		return nil
	case reflect.Uint32:
		v.SetUint(uint64(decodeUint32(d)))
		return nil
	case reflect.Uint64:
		v.SetUint(decodeUint64(d))
		return nil
	case reflect.Float32:
		v.SetFloat(float64(decodeFloat32(d)))
		return nil
	case reflect.Float64:
		v.SetFloat(decodeFloat64(d))
		return nil
	case reflect.Complex64:
		v.SetComplex(complex128(decodeComplex64(d)))
		return nil
	case reflect.Complex128:
		v.SetComplex(decodeComplex128(d))
		return nil
	case reflect.String:
		length, err := decodeConcrete[uint32](r)
		if err != nil {
			return fmt.Errorf("decoding string length: %w", err)
		}

		if length == 0 {
			v.SetString("")
			return nil
		}

		// TODO: sync.Pool
		d := make([]byte, length)

		n, err := r.Read(d)
		if n == 0 && err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("reading encoded string: %w", err)
		}

		// TODO: avoid allocation?
		v.SetString(string(d))
		return nil
	case reflect.Struct:
		for i := range v.NumField() {
			if err := decodeValue(r, v.Field(i)); err != nil {
				return fmt.Errorf("decoding struct field %d of type %s: %w", i, v.Field(i).Type().String(), err)
			}
		}

		return nil
	case reflect.Array, reflect.Slice:
		length32, err := decodeConcrete[uint32](r)
		if err != nil {
			return fmt.Errorf("decoding %s length: %w", v.Kind(), err)
		}

		if length32 == 0 {
			return nil
		}

		length := int(length32)
		elemType := t.Elem()

		// Calculate number of indirection for slice's underlying type.
		indirections, err := numIndirections(elemType)
		if err != nil {
			return err
		}

		// Indirect the underlying slice type.
		for range indirections {
			elemType = elemType.Elem()
		}

		// Allocate underlying slice.
		if v.Kind() == reflect.Slice {
			v.Grow(length)
			v.SetLen(length)
		}

		// Decode slice with underlying type of variable size.
		for i := range length {
			if err := decodeValue(r, v.Index(i)); err != nil {
				return fmt.Errorf("decoding %s index %d of type %s: %w", v.Kind().String(), i, elemType.String(), err)
			}
		}

		return nil
	case reflect.Map:
		length32, err := decodeConcrete[uint32](r)
		if err != nil {
			return fmt.Errorf("decoding map length: %w", err)
		}

		if length32 == 0 {
			return nil
		}

		length := int(length32)

		v.Set(reflect.MakeMapWithSize(v.Type(), length))

		for range length {
			key := reflect.New(v.Type().Key())

			// As
			if err := decodeValue(r, key); err != nil {
				return fmt.Errorf("decoding map key: %w", err)
			}

			// value := reflect.New(v.Type().Elem())

			if err := decodeValue(r, v.MapIndex(key)); err != nil {
				return fmt.Errorf("decoding map value: %w", err)
			}

			// v.SetMapIndex(key.Elem(), value.Elem())
		}

		return nil
	default:
		return fmt.Errorf("decoding of type %s is not supported", v.Type().String())
	}
}
