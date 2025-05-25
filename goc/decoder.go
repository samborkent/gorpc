package goc

import (
	"bytes"
	"encoding"
	"errors"
	"fmt"
	"io"
	"reflect"
	"unsafe"

	"github.com/samborkent/gorpc/internal/convert"
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
	// Try to decode through interface implementation.
	switch decoder := any(new(T)).(type) {
	case DecodeReader:
		return decodeDecodeReader[T](r, decoder)
	case Decoder:
		return decodeDecoder[T](r, decoder)
	case encoding.BinaryUnmarshaler:
		return decodeBinaryUnmarshaler[T](r, decoder)
	}

	switch decoder := any(*new(T)).(type) {
	case DecodeReader:
		return decodeDecodeReader[T](r, decoder)
	case Decoder:
		return decodeDecoder[T](r, decoder)
	case encoding.BinaryUnmarshaler:
		return decodeBinaryUnmarshaler[T](r, decoder)
	}

	// Try to decode concrete type.
	switch reflect.TypeOf(*new(T)).Kind() {
	case reflect.Bool,
		reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128:
		val, err := decodeConcrete[T](r)
		if err != nil {
			return *new(T), fmt.Errorf("decoding %s: %w", reflect.TypeOf(*new(T)).String(), err)
		}

		return val, nil
	case reflect.String:
		str, err := io.ReadAll(r)
		if err != nil {
			return *new(T), fmt.Errorf("reading encoded string: %w", err)
		}

		return unsafeCast[T](convert.BytesToString(str)), nil
	}

	val := new(T)

	// Decode through reflection.
	if err := DecodeValue(r, reflect.ValueOf(val)); err != nil {
		return *new(T), fmt.Errorf("DecodeValue: %w", err)
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
		// TODO: sync.Pool
		d = make([]byte, t.Size())

		n, err := r.Read(d)
		if n == 0 && err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("reading %s: %w", t.String(), err)
		}
	}

	switch v.Kind() {
	case reflect.Bool:
		v.SetBool(unsafeSliceCast[bool](d))
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

			v.SetInt(int64(unsafeSliceCast[int32](d)))
			return nil
		case 8:
			d = make([]byte, 8)

			n, err = r.Read(d)
			if n == 0 && err != nil && !errors.Is(err, io.EOF) {
				return fmt.Errorf("reading %s: %w", t.String(), err)
			}

			v.SetInt(unsafeSliceCast[int64](d))
			return nil
		default:
			return fmt.Errorf("unknown int size %d encountered", header[0])
		}
	case reflect.Int8:
		v.SetInt(int64(unsafeSliceCast[int8](d)))
		return nil
	case reflect.Int16:
		v.SetInt(int64(unsafeSliceCast[int16](d)))
		return nil
	case reflect.Int32:
		v.SetInt(int64(unsafeSliceCast[int32](d)))
		return nil
	case reflect.Int64:
		v.SetInt(unsafeSliceCast[int64](d))
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

			v.SetUint(uint64(unsafeSliceCast[uint32](d)))
			return nil
		case 8:
			d = make([]byte, 8)

			n, err = r.Read(d)
			if n == 0 && err != nil && !errors.Is(err, io.EOF) {
				return fmt.Errorf("reading %s: %w", t.String(), err)
			}

			v.SetUint(unsafeSliceCast[uint64](d))
			return nil
		default:
			return fmt.Errorf("unknown %s size %d encountered", t.Kind(), header[0])
		}
	case reflect.Uint8:
		v.SetUint(uint64(unsafeSliceCast[uint8](d)))
		return nil
	case reflect.Uint16:
		v.SetUint(uint64(unsafeSliceCast[uint16](d)))
		return nil
	case reflect.Uint32:
		v.SetUint(uint64(unsafeSliceCast[uint32](d)))
		return nil
	case reflect.Uint64:
		v.SetUint(unsafeSliceCast[uint64](d))
		return nil
	case reflect.Float32:
		v.SetFloat(float64(unsafeSliceCast[float32](d)))
		return nil
	case reflect.Float64:
		v.SetFloat(unsafeSliceCast[float64](d))
		return nil
	case reflect.Complex64:
		v.SetComplex(complex128(unsafeSliceCast[complex64](d)))
		return nil
	case reflect.Complex128:
		v.SetComplex(unsafeSliceCast[complex128](d))
		return nil
	case reflect.String:
		length, err := decodeConcrete[uint32](r)
		if err != nil {
			return fmt.Errorf("decoding string length: %w", err)
		}

		// TODO: sync.Pool
		d := make([]byte, length)

		n, err := r.Read(d)
		if n == 0 && err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("reading encoded string: %w", err)
		}

		v.SetString(convert.BytesToString(d))
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

		// Decode slice of fixed-size types.
		if v.Kind() == reflect.Slice {
			switch elemType.Kind() {
			case reflect.Bool,
				reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
				reflect.Float32, reflect.Float64,
				reflect.Complex64, reflect.Complex128:
				// TODO: sync.Pool
				d := make([]byte, length*int(elemType.Size()))

				n, err := r.Read(d)
				if n == 0 && err != nil && !errors.Is(err, io.EOF) {
					return fmt.Errorf("reading []%s: %w", elemType.String(), err)
				}

				v.Set(reflect.SliceAt(elemType, unsafe.Pointer(&d[0]), length))
				return nil
			}

			v.Set(reflect.MakeSlice(v.Type(), length, length))
		}

		// Decode slice with underlying type of variable size.
		for i := range length {
			if err := decodeValue(r, v.Index(i)); err != nil {
				return fmt.Errorf("decoding %s index %d of type %s: %w", v.Kind(), i, v.Index(i).Type().String(), err)
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

			if err := decodeValue(r, key); err != nil {
				return fmt.Errorf("decoding map key: %w", err)
			}

			value := reflect.New(v.Type().Elem())

			if err := decodeValue(r, value); err != nil {
				return fmt.Errorf("decoding map value: %w", err)
			}

			v.SetMapIndex(key.Elem(), value.Elem())
		}

		return nil
	default:
		return fmt.Errorf("decoding of type %s is not supported", v.Type().String())
	}
}
