package goc

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"math/bits"
	"reflect"
)

// TODO: implement unsafe decoding

func decodeReader[T any](in *bytes.Reader, out *T, allowUnsafe bool) error {
	v := reflect.ValueOf(out)
	return decodeValueWithInterfaces(in, v, v.Type(), allowUnsafe)
}

var (
	reflectDecodeByteReader   = reflect.TypeFor[DecodeByteReader]()
	reflectDecodeReader       = reflect.TypeFor[DecodeReader]()
	reflectDecoder            = reflect.TypeFor[Decoder]()
	reflectBinaryUnmarshaller = reflect.TypeFor[encoding.BinaryUnmarshaler]()
)

func decodeValueWithInterfaces(r *bytes.Reader, v reflect.Value, t reflect.Type, allowUnsafe bool) error {
	switch {
	case t.Implements(reflectDecodeByteReader):
		return v.Interface().(DecodeByteReader).DecodeByteRead(r)
	case t.Implements(reflectDecodeReader):
		return v.Interface().(DecodeReader).DecodeRead(r)
	case t.Implements(reflectDecoder):
		buf := decodingPool.Get()
		defer decodingPool.Put(buf)

		_, err := r.WriteTo(buf)
		if err != nil {
			return err
		}

		return v.Interface().(Decoder).Decode(buf.Bytes())
	case t.Implements(reflectBinaryUnmarshaller):
		buf := decodingPool.Get()
		defer decodingPool.Put(buf)

		_, err := r.WriteTo(buf)
		if err != nil {
			return err
		}

		return v.Interface().(encoding.BinaryUnmarshaler).UnmarshalBinary(buf.Bytes())
	default:
		return decodeValue(r, v, t, allowUnsafe)
	}
}

func decodeValue(r *bytes.Reader, v reflect.Value, t reflect.Type, allowUnsafe bool) error {
	if !v.IsValid() {
		return ErrInvalidValue
	}

	indirections, err := numIndirections(t)
	if err != nil {
		return err
	}

	for range indirections {
		v = reflect.Indirect(v)
	}

	if indirections > 0 {
		t = v.Type()
	}

	var data []byte

	// Read concrete type.
	switch t.Kind() {
	case reflect.Bool,
		reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128:
		data = make([]byte, t.Size())

		n, err := r.Read(data)
		if n == 0 && err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("reading %s: %w", t.String(), err)
		}
	}

	switch v.Kind() {
	case reflect.Bool:
		v.SetBool(decodeBool(data[0]))
		return nil
	case reflect.Int:
		var d [8]byte

		_, err := r.Read(d[:])
		if err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("reading %s: %w", t.String(), err)
		}

		decoded := decodeInt64(d[:])

		if decoded < (-1<<(bits.UintSize-1)) || decoded > (1<<(bits.UintSize-1)-1) {
			return ErrOverflowInt
		}

		v.SetInt(decoded)
		return nil
	case reflect.Int8:
		v.SetInt(int64(int8(data[0])))
		return nil
	case reflect.Int16:
		v.SetInt(int64(decodeInt16(data)))
		return nil
	case reflect.Int32:
		v.SetInt(int64(decodeInt32(data)))
		return nil
	case reflect.Int64:
		v.SetInt(decodeInt64(data))
		return nil
	case reflect.Uint, reflect.Uintptr:
		var d [8]byte

		_, err := r.Read(d[:])
		if err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("reading %s: %w", t.String(), err)
		}

		decoded := decodeUint64(d[:])

		if decoded > (1<<bits.UintSize - 1) {
			return ErrOverflowUint
		}

		v.SetUint(decoded)
		return nil
	case reflect.Uint8:
		v.SetUint(uint64(data[0]))
		return nil
	case reflect.Uint16:
		v.SetUint(uint64(decodeUint16(data)))
		return nil
	case reflect.Uint32:
		v.SetUint(uint64(decodeUint32(data)))
		return nil
	case reflect.Uint64:
		v.SetUint(decodeUint64(data))
		return nil
	case reflect.Float32:
		v.SetFloat(float64(decodeFloat32(data)))
		return nil
	case reflect.Float64:
		v.SetFloat(decodeFloat64(data))
		return nil
	case reflect.Complex64:
		v.SetComplex(complex128(decodeComplex64(data)))
		return nil
	case reflect.Complex128:
		v.SetComplex(decodeComplex128(data))
		return nil
	case reflect.String:
		var d [4]byte

		_, err := r.Read(d[:])
		if err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("decoding string len: %w", err)
		}

		length := decodeUint32(d[:])

		if length == 0 {
			v.SetString("")
			return nil
		}

		// TODO: sync.Pool
		data = make([]byte, int(length))

		n, err := r.Read(data)
		if n == 0 && err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("decoding string: %w", err)
		}

		// TODO: avoid allocation?
		v.SetString(string(data))
		return nil
	case reflect.Struct:
		for i := range v.NumField() {
			f := v.Field(i)

			if err := decodeValue(r, f, f.Type(), allowUnsafe); err != nil {
				return fmt.Errorf("decoding struct field %d of type %s: %w", i, v.Field(i).Type().String(), err)
			}
		}

		return nil
	case reflect.Array, reflect.Slice:
		var d [4]byte

		_, err := r.Read(d[:])
		if err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("decoding %s len: %w", t.Kind(), err)
		}

		length32 := decodeUint32(d[:])

		if length32 == 0 {
			v.Set(reflect.MakeSlice(t, 0, 0))
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
			v.Set(reflect.MakeSlice(t, length, length))
		}

		// Decode slice with underlying type of variable size.
		for i := range length {
			elem := v.Index(i)

			if err := decodeValue(r, elem, elemType, allowUnsafe); err != nil {
				return fmt.Errorf("decoding %s index %d of type %s: %w", elem.Kind().String(), i, elemType.String(), err)
			}
		}

		return nil
	case reflect.Map:
		var d [4]byte

		_, err := r.Read(d[:])
		if err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("decoding map len: %w", err)
		}

		length32 := decodeUint32(d[:])

		if length32 == 0 {
			v.Set(reflect.MakeMap(t))
			return nil
		}

		length := int(length32)

		v.Set(reflect.MakeMapWithSize(t, length))

		key := reflect.New(t.Key())
		val := reflect.New(t.Elem())

		for range length {
			keyElem := key.Elem()

			if err := decodeValue(r, keyElem, keyElem.Type(), allowUnsafe); err != nil {
				return fmt.Errorf("decoding map key: %w", err)
			}

			valElem := val.Elem()

			if err := decodeValue(r, valElem, valElem.Type(), allowUnsafe); err != nil {
				return fmt.Errorf("decoding map value: %w", err)
			}

			v.SetMapIndex(keyElem, valElem)
		}

		return nil
	default:
		return fmt.Errorf("decoding of type %s is not supported", t.String())
	}
}

func decodeBool(b byte) bool {
	return b != 0
}

func decodeInt16(b []byte) int16 {
	return int16(decodeUint16(b))
}

func decodeUint16(b []byte) uint16 {
	_ = b[1] // bounds check hint for compiler
	return binary.LittleEndian.Uint16(b[:2])
}

func decodeInt32(b []byte) int32 {
	return int32(decodeUint32(b))
}

func decodeUint32(b []byte) uint32 {
	_ = b[3] // bounds check hint for compiler
	return binary.LittleEndian.Uint32(b[:4])
}

func decodeInt64(b []byte) int64 {
	return int64(decodeUint64(b))
}

func decodeUint64(b []byte) uint64 {
	_ = b[7] // bounds check hint for compiler
	return binary.LittleEndian.Uint64(b[:8])
}

func decodeFloat32(b []byte) float32 {
	return math.Float32frombits(decodeUint32(b))
}

func decodeFloat64(b []byte) float64 {
	return math.Float64frombits(decodeUint64(b))
}

func decodeComplex64(b []byte) complex64 {
	_ = b[7] // bounds check hint for compiler
	r := decodeFloat32(b[:4])
	i := decodeFloat32(b[4:8])
	return complex(r, i)
}

func decodeComplex128(b []byte) complex128 {
	_ = b[15] // bounds check hint for compiler
	r := decodeFloat64(b[:8])
	i := decodeFloat64(b[8:16])
	return complex(r, i)
}
