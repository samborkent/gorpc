package goc

import (
	"bytes"
	"encoding"
	"fmt"
	"io"
	"math"
	"reflect"

	"github.com/samborkent/gorpc/internal/convert"
)

type EncodeWriter interface {
	EncodeTo(w io.Writer) error
}

type Encoder interface {
	Encode() ([]byte, error)
}

func Encode[T any](val T) ([]byte, error) {
	buf := new(bytes.Buffer)

	if err := EncodeTo(buf, val); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func EncodeTo[T any](w io.Writer, val T) error {
	// Try to encode through interface implementation.
	switch encoder := any(val).(type) {
	case EncodeWriter:
		if err := encoder.EncodeTo(w); err != nil {
			return fmt.Errorf("EncodeWriter: %w", err)
		}

		return nil
	case Encoder:
		encoded, err := encoder.Encode()
		if err != nil {
			return fmt.Errorf("Encoder: %w", err)
		}

		_, err = w.Write(encoded)
		if err != nil {
			return fmt.Errorf("Encoder: write: %w", err)
		}
	case encoding.BinaryMarshaler:
		encoded, err := encoder.MarshalBinary()
		if err != nil {
			return fmt.Errorf("BinaryMarshaler: %w", err)
		}

		_, err = w.Write(encoded)
		if err != nil {
			return fmt.Errorf("BinaryMarshaler: write: %w", err)
		}
	}

	v := reflect.ValueOf(val)

	// Try to encode concrete type.
	switch v.Kind() {
	case reflect.Bool,
		reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128:
		if err := encodeConcrete(w, val); err != nil {
			return fmt.Errorf("encoding %s: %w", v.Type().String(), err)
		}

		return nil
	case reflect.String:
		_, err := w.Write(convert.StringToBytes(unsafeCast[string](val)))
		if err != nil {
			return fmt.Errorf("encoding string: %w", err)
		}

		return nil
	}

	// Encode through reflection.
	return encodeValue(w, v)
}

var (
	reflectEncodeWriter     = reflect.TypeFor[EncodeWriter]()
	reflectEncoder          = reflect.TypeFor[Encoder]()
	reflectBinaryMarshaller = reflect.TypeFor[encoding.BinaryMarshaler]()
)

func EncodeValue(w io.Writer, v reflect.Value) error {
	if v.Type().Implements(reflectEncodeWriter) {
		if err := reflect.TypeAssert[EncodeWriter](v).EncodeTo(w); err != nil {
			return fmt.Errorf("EncodeWriter: %w", err)
		}

		return nil
	} else if v.Type().Implements(reflectEncoder) {
		encoded, err := reflect.TypeAssert[Encoder](v).Encode()
		if err != nil {
			return fmt.Errorf("Encoder: %w", err)
		}

		_, err = w.Write(encoded)
		if err != nil {
			return fmt.Errorf("Encoder: write: %w", err)
		}
	} else if v.Type().Implements(reflectBinaryMarshaller) {
		encoded, err := reflect.TypeAssert[encoding.BinaryMarshaler](v).MarshalBinary()
		if err != nil {
			return fmt.Errorf("BinaryMarshaler: %w", err)
		}

		_, err = w.Write(encoded)
		if err != nil {
			return fmt.Errorf("BinaryMarshaler: write: %w", err)
		}
	}

	return encodeValue(w, v)
}

func encodeValue(w io.Writer, v reflect.Value) error {
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

	switch v.Kind() {
	case reflect.Bool:
		if err := encodeConcrete(w, v.Bool()); err != nil {
			return fmt.Errorf("encoding bool: %w", err)
		}

		return nil
	case reflect.Int:
		// TODO: test
		switch v.Type().Size() {
		case 4:
			if err := encodeConcrete(w, uint8(4)); err != nil {
				return fmt.Errorf("encoding 32-bit int header: %w", err)
			}

			if err := encodeConcrete(w, int32(v.Int())); err != nil {
				return fmt.Errorf("encoding 32-bit int: %w", err)
			}

			return nil
		case 8:
			if err := encodeConcrete(w, uint8(8)); err != nil {
				return fmt.Errorf("encoding 64-bit int header: %w", err)
			}

			if err := encodeConcrete(w, v.Int()); err != nil {
				return fmt.Errorf("encoding 64-bit int: %w", err)
			}

			return nil
		default:
			return fmt.Errorf("unknown int size %d encountered", v.Type().Size())
		}
	case reflect.Int8:
		if err := encodeConcrete(w, int8(v.Int())); err != nil {
			return fmt.Errorf("encoding int8: %w", err)
		}

		return nil
	case reflect.Int16:
		if err := encodeConcrete(w, int16(v.Int())); err != nil {
			return fmt.Errorf("encoding int16: %w", err)
		}

		return nil
	case reflect.Int32:
		if err := encodeConcrete(w, int32(v.Int())); err != nil {
			return fmt.Errorf("encoding int32: %w", err)
		}

		return nil
	case reflect.Int64:
		if err := encodeConcrete(w, v.Int()); err != nil {
			return fmt.Errorf("encoding int64: %w", err)
		}

		return nil
	case reflect.Uint, reflect.Uintptr:
		// TODO: decode
		// TODO: test
		switch v.Type().Size() {
		case 4:
			if err := encodeConcrete(w, uint8(4)); err != nil {
				return fmt.Errorf("encoding 32-bit %s header: %w", v.Kind().String(), err)
			}

			if err := encodeConcrete(w, uint32(v.Uint())); err != nil {
				return fmt.Errorf("encoding 32-bit %s: %w", v.Type().String(), err)
			}

			return nil
		case 8:
			if err := encodeConcrete(w, uint8(8)); err != nil {
				return fmt.Errorf("encoding 64-bit %s header: %w", v.Kind().String(), err)
			}

			if err := encodeConcrete(w, v.Uint()); err != nil {
				return fmt.Errorf("encoding 64-bit %s: %w", v.Type().String(), err)
			}

			return nil
		default:
			return fmt.Errorf("unknown int size %d encountered", v.Type().Size())
		}
	case reflect.Uint8:
		if err := encodeConcrete(w, uint8(v.Uint())); err != nil {
			return fmt.Errorf("encoding uint8: %w", err)
		}

		return nil
	case reflect.Uint16:
		if err := encodeConcrete(w, uint16(v.Uint())); err != nil {
			return fmt.Errorf("encoding uint16: %w", err)
		}

		return nil
	case reflect.Uint32:
		if err := encodeConcrete(w, uint32(v.Uint())); err != nil {
			return fmt.Errorf("encoding uint32: %w", err)
		}

		return nil
	case reflect.Uint64:
		if err := encodeConcrete(w, v.Uint()); err != nil {
			return fmt.Errorf("encoding uint64: %w", err)
		}

		return nil
	case reflect.Float32:
		if err := encodeConcrete(w, float32(v.Float())); err != nil {
			return fmt.Errorf("encoding float32: %w", err)
		}

		return nil
	case reflect.Float64:
		if err := encodeConcrete(w, v.Float()); err != nil {
			return fmt.Errorf("encoding float64: %w", err)
		}

		return nil
	case reflect.Complex64:
		if err := encodeConcrete(w, complex64(v.Complex())); err != nil {
			return fmt.Errorf("encoding complex64: %w", err)
		}

		return nil
	case reflect.Complex128:
		if err := encodeConcrete(w, v.Complex()); err != nil {
			return fmt.Errorf("encoding complex128: %w", err)
		}

		return nil
	case reflect.String:
		if v.Len() > math.MaxInt32 {
			return fmt.Errorf("maximum string size of %d bytes exceeded", math.MaxInt32)
		}

		if err := encodeConcrete(w, uint32(v.Len())); err != nil {
			return fmt.Errorf("encoding string len: %w", err)
		}

		// _, err := w.Write(convert.StringToBytes(v.String()))
		_, err := w.Write([]byte(v.String()))
		if err != nil {
			return fmt.Errorf("encoding string: %w", err)
		}

		return nil
	case reflect.Struct:
		for i := range v.NumField() {
			if err := encodeValue(w, v.Field(i)); err != nil {
				return fmt.Errorf("encoding struct field %d of type %s: %w", i, v.Field(i).Type().String(), err)
			}
		}

		return nil
	case reflect.Array, reflect.Slice:
		if err := encodeConcrete(w, uint32(v.Len())); err != nil {
			return fmt.Errorf("encoding slice len: %w", err)
		}

		if v.Len() == 0 {
			return nil
		}

		elemType := v.Type().Elem()

		// Calculate number of indirection for slice's underlying type.
		indirections, err := numIndirections(elemType)
		if err != nil {
			return err
		}

		// Indirect the underlying slice type.
		for range indirections {
			elemType = elemType.Elem()
		}

		// Encode slice of fixed-size types.
		if v.Kind() == reflect.Slice {
			switch elemType.Kind() {
			case reflect.Bool:
				if err := encodeConcreteSlice(w, reflect.TypeAssert[[]bool](v)); err != nil {
					return fmt.Errorf("encoding []%s: %w", elemType.String(), err)
				}

				return nil
			case reflect.Int8:
				if err := encodeConcreteSlice(w, reflect.TypeAssert[[]int8](v)); err != nil {
					return fmt.Errorf("encoding []%s: %w", elemType.String(), err)
				}

				return nil
			case reflect.Int16:
				if err := encodeConcreteSlice(w, reflect.TypeAssert[[]int16](v)); err != nil {
					return fmt.Errorf("encoding []%s: %w", elemType.String(), err)
				}

				return nil
			case reflect.Int32:
				if err := encodeConcreteSlice(w, reflect.TypeAssert[[]int32](v)); err != nil {
					return fmt.Errorf("encoding []%s: %w", elemType.String(), err)
				}

				return nil
			case reflect.Int64:
				if err := encodeConcreteSlice(w, reflect.TypeAssert[[]int64](v)); err != nil {
					return fmt.Errorf("encoding []%s: %w", elemType.String(), err)
				}

				return nil
			case reflect.Uint8:
				if _, err := w.Write(v.Bytes()); err != nil {
					return fmt.Errorf("encoding []%s: %w", elemType.String(), err)
				}

				return nil
			case reflect.Uint16:
				if err := encodeConcreteSlice(w, reflect.TypeAssert[[]uint16](v)); err != nil {
					return fmt.Errorf("encoding []%s: %w", elemType.String(), err)
				}

				return nil
			case reflect.Uint32:
				if err := encodeConcreteSlice(w, reflect.TypeAssert[[]uint32](v)); err != nil {
					return fmt.Errorf("encoding []%s: %w", elemType.String(), err)
				}

				return nil
			case reflect.Uint64:
				if err := encodeConcreteSlice(w, reflect.TypeAssert[[]uint64](v)); err != nil {
					return fmt.Errorf("encoding []%s: %w", elemType.String(), err)
				}

				return nil
			case reflect.Float32:
				if err := encodeConcreteSlice(w, reflect.TypeAssert[[]float32](v)); err != nil {
					return fmt.Errorf("encoding []%s: %w", elemType.String(), err)
				}

				return nil
			case reflect.Float64:
				if err := encodeConcreteSlice(w, reflect.TypeAssert[[]float64](v)); err != nil {
					return fmt.Errorf("encoding []%s: %w", elemType.String(), err)
				}

				return nil
			case reflect.Complex64:
				if err := encodeConcreteSlice(w, reflect.TypeAssert[[]complex64](v)); err != nil {
					return fmt.Errorf("encoding []%s: %w", elemType.String(), err)
				}

				return nil
			case reflect.Complex128:
				if err := encodeConcreteSlice(w, reflect.TypeAssert[[]complex128](v)); err != nil {
					return fmt.Errorf("encoding []%s: %w", elemType.String(), err)
				}

				return nil
			}
		}

		// Encode slice with underlying type of variable size.
		for i := range v.Len() {
			if err := encodeValue(w, v.Index(i)); err != nil {
				return fmt.Errorf("encoding %s index %d of type %s: %w", v.Kind().String(), i, v.Index(i).Type().String(), err)
			}
		}

		return nil
	case reflect.Map:
		if err := encodeConcrete(w, uint32(v.Len())); err != nil {
			return fmt.Errorf("encoding map len: %w", err)
		}

		if v.Len() == 0 {
			return nil
		}

		iter := v.MapRange()

		for iter.Next() {
			key := iter.Key()

			if err := encodeValue(w, key); err != nil {
				return fmt.Errorf("encoding map key: %w", err)
			}

			value := iter.Value()

			if err := encodeValue(w, value); err != nil {
				return fmt.Errorf("encoding map value: %w", err)
			}
		}

		return nil
	default:
		return fmt.Errorf("encoding of type %s if not supported", v.Type().String())
	}
}
