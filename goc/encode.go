package goc

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
)

// TODO: implement unsafe encoding

func encodeBuffer[T any](out *bytes.Buffer, in T, allowUnsafe bool) error {
	v := reflect.ValueOf(in)
	return encodeValueWithInterfaces(out, v, v.Type(), allowUnsafe)
}

var (
	reflectEncodeByteWriter = reflect.TypeFor[EncodeByteWriter]()
	reflectEncodeWriter     = reflect.TypeFor[EncodeWriter]()
	reflectEncoder          = reflect.TypeFor[Encoder]()
	reflectBinaryMarshaller = reflect.TypeFor[encoding.BinaryMarshaler]()
)

func encodeValueWithInterfaces(b *bytes.Buffer, v reflect.Value, t reflect.Type, allowUnsafe bool) error {
	switch {
	case t.Implements(reflectEncodeByteWriter):
		return v.Interface().(EncodeByteWriter).EncodeBuffer(b)
	case t.Implements(reflectEncodeWriter):
		_, err := v.Interface().(EncodeWriter).EncodeWrite(b)
		return err
	case t.Implements(reflectEncoder):
		d, err := v.Interface().(Encoder).Encode()
		if err != nil {
			return err
		}

		_, err = b.Write(d)
		return err
	case t.Implements(reflectBinaryMarshaller):
		d, err := v.Interface().(encoding.BinaryMarshaler).MarshalBinary()
		if err != nil {
			return err
		}

		_, err = b.Write(d)
		return err
	default:
		return encodeValue(b, v, t, allowUnsafe)
	}
}

func encodeValue(b *bytes.Buffer, v reflect.Value, t reflect.Type, allowUnsafe bool) error {
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

	switch t.Kind() {
	case reflect.Bool:
		if err := b.WriteByte(encodeBool(v.Bool())); err != nil {
			return fmt.Errorf("encoding %s: %w", t.String(), err)
		}

		return nil
	case reflect.Int:
		data = encodeInt64(v.Int())
	case reflect.Int8:
		if err := b.WriteByte(byte(v.Int())); err != nil {
			return fmt.Errorf("encoding %s: %w", t.String(), err)
		}

		return nil
	case reflect.Int16:
		data = encodeInt16(int16(v.Int()))
	case reflect.Int32:
		data = encodeInt32(int32(v.Int()))
	case reflect.Int64:
		data = encodeInt64(v.Int())
	case reflect.Uint, reflect.Uintptr:
		data = encodeUint64(v.Uint())
	case reflect.Uint8:
		if err := b.WriteByte(byte(v.Uint())); err != nil {
			return fmt.Errorf("encoding %s: %w", t.String(), err)
		}

		return nil
	case reflect.Uint16:
		data = encodeUint16(uint16(v.Uint()))
	case reflect.Uint32:
		data = encodeUint32(uint32(v.Uint()))
	case reflect.Uint64:
		data = encodeUint64(v.Uint())
	case reflect.Float32:
		data = encodeFloat32(float32(v.Float()))
	case reflect.Float64:
		data = encodeFloat64(v.Float())
	case reflect.Complex64:
		data = encodeComplex64(complex64(v.Complex()))
	case reflect.Complex128:
		data = encodeComplex128(v.Complex())
	case reflect.String:
		strLen := v.Len()

		if strLen > math.MaxUint32 {
			return fmt.Errorf("maximum string size of %d bytes exceeded", math.MaxUint32)
		}

		// Empty string.
		if strLen == 0 {
			data = []byte{0, 0}
			break
		}

		_, err := b.Write(encodeUint32(uint32(v.Len())))
		if err != nil {
			return fmt.Errorf("encoding string len: %w", err)
		}

		_, err = b.WriteString(v.String())
		if err != nil {
			return fmt.Errorf("encoding string: %w", err)
		}

		return nil
	case reflect.Struct:
		for i := range v.NumField() {
			f := v.Field(i)

			if err := encodeValueWithInterfaces(b, f, f.Type(), allowUnsafe); err != nil {
				return fmt.Errorf("encoding struct field %d of type %s: %w", i, f.Type().String(), err)
			}
		}

		return nil
	case reflect.Array, reflect.Slice:
		sliceLen := v.Len()

		if sliceLen > math.MaxUint32 {
			return fmt.Errorf("maximum array/slice size of %d bytes exceeded", math.MaxUint32)
		}

		// Empty slice.
		if v.Len() == 0 {
			data = []byte{0, 0}
			break
		}

		_, err = b.Write(encodeUint32(uint32(sliceLen)))
		if err != nil {
			return fmt.Errorf("encoding slice len: %w", err)
		}

		elemType := v.Index(0).Type()

		// Encode slice with underlying type of variable size.
		if elemType.Implements(reflectEncodeByteWriter) ||
			elemType.Implements(reflectEncodeWriter) ||
			elemType.Implements(reflectEncoder) ||
			elemType.Implements(reflectBinaryMarshaller) {
			for i := range v.Len() {
				f := v.Index(i)

				if err := encodeValueWithInterfaces(b, f, f.Type(), allowUnsafe); err != nil {
					return fmt.Errorf("encoding %s index %d of type %s: %w", v.Kind().String(), i, f.Type().String(), err)
				}
			}
		} else {
			for i := range v.Len() {
				f := v.Index(i)

				if err := encodeValue(b, f, f.Type(), allowUnsafe); err != nil {
					return fmt.Errorf("encoding %s index %d of type %s: %w", v.Kind().String(), i, f.Type().String(), err)
				}
			}
		}

		return nil
	case reflect.Map:
		mapLen := v.Len()

		if mapLen > math.MaxUint32 {
			return fmt.Errorf("maximum map size of %d bytes exceeded", math.MaxUint32)
		}

		// Empty map.
		if v.Len() == 0 {
			data = []byte{0, 0}
			break
		}

		_, err = b.Write(encodeUint32(uint32(mapLen)))
		if err != nil {
			return fmt.Errorf("encoding map len: %w", err)
		}

		iter := v.MapRange()

		for iter.Next() {
			key := iter.Key()

			if err := encodeValue(b, key, key.Type(), allowUnsafe); err != nil {
				return fmt.Errorf("encoding map key of type %s: %w", key.Type().String(), err)
			}

			val := iter.Value()

			if err := encodeValue(b, val, val.Type(), allowUnsafe); err != nil {
				return fmt.Errorf("encoding map value of type %s: %w", val.Type().String(), err)
			}
		}

		return nil
	default:
		return fmt.Errorf("encoding of type %s is not supported", t.String())
	}

	if len(data) > 0 {
		_, err = b.Write(data)
		if err != nil {
			return fmt.Errorf("encoding %s: %w", t.String(), err)
		}
	}

	return nil
}

func encodeBool(v bool) byte {
	if v {
		return 1
	} else {
		return 0
	}
}

func encodeInt16(v int16) []byte {
	return encodeUint16(uint16(v))
}

func encodeUint16(v uint16) []byte {
	var d [2]byte
	binary.LittleEndian.PutUint16(d[:], v)
	return d[:]
}

func encodeInt32(v int32) []byte {
	return encodeUint32(uint32(v))
}

func encodeUint32(v uint32) []byte {
	var d [4]byte
	binary.LittleEndian.PutUint32(d[:], v)
	return d[:]
}

func encodeInt64(v int64) []byte {
	return encodeUint64(uint64(v))
}

func encodeUint64(v uint64) []byte {
	var d [8]byte
	binary.LittleEndian.PutUint64(d[:], v)
	return d[:]
}

func encodeFloat32(v float32) []byte {
	return encodeUint32(math.Float32bits(v))
}

func encodeFloat64(v float64) []byte {
	return encodeUint64(math.Float64bits(v))
}

func encodeComplex64(v complex64) []byte {
	var d [8]byte
	copy(d[:4], encodeFloat32(real(v)))
	copy(d[4:], encodeFloat32(imag(v)))
	return d[:]
}

func encodeComplex128(v complex128) []byte {
	var d [16]byte
	copy(d[:8], encodeFloat64(real(v)))
	copy(d[8:], encodeFloat64(imag(v)))
	return d[:]
}
