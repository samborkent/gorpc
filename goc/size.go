package goc

import "reflect"

// Size returns the encoded size in bytes of a [reflect.Value].
func Size(v reflect.Value) int {
	if !v.IsValid() {
		return 0
	}

	indirections, err := numIndirections(v.Type())
	if err != nil {
		return 0
	}

	for range indirections {
		v = reflect.Indirect(v)
	}

	switch v.Kind() {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128:
		return int(v.Type().Size())
	case reflect.Array, reflect.Slice:
		size := 4

		for i := range v.Len() {
			size += Size(v.Index(i))
		}

		return size
	case reflect.Map:
		size := 4
		iter := v.MapRange()

		for iter.Next() {
			size += Size(iter.Key())
			size += Size(iter.Value())
		}

		return size
	case reflect.String:
		return 4 + v.Len()
	case reflect.Struct:
		size := 0

		for i := range v.NumField() {
			size += Size(v.Field(i))
		}

		return size
	default:
		return 0
	}
}
