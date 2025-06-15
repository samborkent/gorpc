package goc

import (
	cryptorand "crypto/rand"
	"math"
	"math/rand/v2"
	"testing"
)

// TODO: test non-comparable structs with pointer fields
// TODO: test nested arrays/slices

type ComparableStruct struct {
	Bool       bool
	Int8       int8
	Int16      int16
	Int32      int32
	Int64      int64
	Uint8      uint8
	Uint16     uint16
	Uint32     uint32
	Uint64     uint64
	Float32    float32
	Float64    float64
	Complex64  complex64
	Complex128 complex128
	String     string
	Struct     struct {
		A int64
		B string
	}
}

func TestEncodeDecode(t *testing.T) {
	t.Parallel()

	t.Run("bool", func(t *testing.T) {
		t.Parallel()

		want := true

		encodeDecodeComparable(t, want)
	})
	t.Run("int", func(t *testing.T) {
		t.Parallel()

		want := rand.Int()

		encodeDecodeComparable(t, want)
	})
	// t.Run("int8", func(t *testing.T) {
	// 	t.Parallel()

	// 	want := int8(rand.IntN(math.MaxUint8) - math.MaxInt8)

	// 	encodeDecodeComparable(t, want)
	// })
	// t.Run("int16", func(t *testing.T) {
	// 	t.Parallel()

	// 	want := int16(rand.IntN(math.MaxUint16) - math.MaxInt16)

	// 	encodeDecodeComparable(t, want)
	// })
	// t.Run("int32", func(t *testing.T) {
	// 	t.Parallel()

	// 	want := rand.Int32() - math.MaxInt32/2

	// 	encodeDecodeComparable(t, want)
	// })
	// t.Run("int64", func(t *testing.T) {
	// 	t.Parallel()

	// 	want := rand.Int64() - math.MaxInt64/2

	// 	encodeDecodeComparable(t, want)
	// })
	// t.Run("uint", func(t *testing.T) {
	// 	t.Parallel()

	// 	want := rand.Uint()

	// 	encodeDecodeComparable(t, want)
	// })
	// t.Run("uintptr", func(t *testing.T) {
	// 	t.Parallel()

	// 	want := uintptr(rand.Uint())

	// 	encodeDecodeComparable(t, want)
	// })
	// t.Run("uint8", func(t *testing.T) {
	// 	t.Parallel()

	// 	want := uint8(rand.IntN(math.MaxUint8))

	// 	encodeDecodeComparable(t, want)
	// })
	// t.Run("uint16", func(t *testing.T) {
	// 	t.Parallel()

	// 	want := uint16(rand.IntN(math.MaxUint16))

	// 	encodeDecodeComparable(t, want)
	// })
	// t.Run("uint32", func(t *testing.T) {
	// 	t.Parallel()

	// 	want := rand.Uint32()

	// 	encodeDecodeComparable(t, want)
	// })
	// t.Run("uint64", func(t *testing.T) {
	// 	t.Parallel()

	// 	want := rand.Uint64()

	// 	encodeDecodeComparable(t, want)
	// })
	// t.Run("float32", func(t *testing.T) {
	// 	t.Parallel()

	// 	want := 2*rand.Float32() - 1

	// 	encodeDecodeComparable(t, want)
	// })
	// t.Run("float64", func(t *testing.T) {
	// 	t.Parallel()

	// 	want := 2*rand.Float64() - 1

	// 	encodeDecodeComparable(t, want)
	// })
	// t.Run("complex64", func(t *testing.T) {
	// 	t.Parallel()

	// 	want := complex64(complex(2*rand.Float64()-1, 2*rand.Float64()-1))

	// 	encodeDecodeComparable(t, want)
	// })
	// t.Run("complex128", func(t *testing.T) {
	// 	t.Parallel()

	// 	want := complex(2*rand.Float64()-1, 2*rand.Float64()-1)

	// 	encodeDecodeComparable(t, want)
	// })
	// t.Run("array", func(t *testing.T) {
	// 	t.Parallel()

	// 	want := [4]string{cryptorand.Text(), cryptorand.Text(), cryptorand.Text(), cryptorand.Text()}

	// 	encodeDecodeComparable(t, want)
	// })
	// t.Run("string", func(t *testing.T) {
	// 	t.Parallel()

	// 	want := cryptorand.Text()

	// 	encodeDecodeComparable(t, want)
	// })
	// t.Run("struct", func(t *testing.T) {
	// 	t.Parallel()

	// 	want := makeComparableStruct(t)

	// 	encodeDecodeComparable(t, want)
	// })
	// t.Run("slice", func(t *testing.T) {
	// 	t.Parallel()

	// 	t.Run("bool", func(t *testing.T) {
	// 		t.Parallel()

	// 		want := make([]bool, rand.IntN(math.MaxInt8))
	// 		for i := range want {
	// 			if i%(1+rand.IntN(7)) == 0 {
	// 				want[i] = true
	// 			}
	// 		}

	// 		encodeDecodeComparableSlice(t, want)
	// 	})
	// 	t.Run("int", func(t *testing.T) {
	// 		t.Parallel()

	// 		want := make([]int, rand.IntN(math.MaxInt8))
	// 		for i := range want {
	// 			want[i] = rand.Int()
	// 		}

	// 		encodeDecodeComparableSlice(t, want)
	// 	})
	// 	t.Run("int8", func(t *testing.T) {
	// 		t.Parallel()

	// 		want := make([]int8, rand.IntN(math.MaxInt8))
	// 		for i := range want {
	// 			want[i] = int8(rand.IntN(math.MaxUint8) - math.MaxInt8)
	// 		}

	// 		encodeDecodeComparableSlice(t, want)
	// 	})
	// 	t.Run("int16", func(t *testing.T) {
	// 		t.Parallel()

	// 		want := make([]int16, rand.IntN(math.MaxInt8))
	// 		for i := range want {
	// 			want[i] = int16(rand.IntN(math.MaxUint16) - math.MaxInt16)
	// 		}

	// 		encodeDecodeComparableSlice(t, want)
	// 	})
	// 	t.Run("int32", func(t *testing.T) {
	// 		t.Parallel()

	// 		want := make([]int32, rand.IntN(math.MaxInt8))
	// 		for i := range want {
	// 			want[i] = rand.Int32() - math.MaxInt32/2
	// 		}

	// 		encodeDecodeComparableSlice(t, want)
	// 	})
	// 	t.Run("int64", func(t *testing.T) {
	// 		t.Parallel()

	// 		want := make([]int64, rand.IntN(math.MaxInt8))
	// 		for i := range want {
	// 			want[i] = rand.Int64() - math.MaxInt64/2
	// 		}

	// 		encodeDecodeComparableSlice(t, want)
	// 	})
	// 	t.Run("uint", func(t *testing.T) {
	// 		t.Parallel()

	// 		want := make([]uint, rand.IntN(math.MaxInt8))
	// 		for i := range want {
	// 			want[i] = rand.Uint()
	// 		}

	// 		encodeDecodeComparableSlice(t, want)
	// 	})
	// 	t.Run("uint8", func(t *testing.T) {
	// 		t.Parallel()

	// 		want := make([]uint8, rand.IntN(math.MaxInt8))
	// 		for i := range want {
	// 			want[i] = uint8(rand.IntN(math.MaxUint8))
	// 		}

	// 		encodeDecodeComparableSlice(t, want)
	// 	})
	// 	t.Run("uint16", func(t *testing.T) {
	// 		t.Parallel()

	// 		want := make([]uint16, rand.IntN(math.MaxInt8))
	// 		for i := range want {
	// 			want[i] = uint16(rand.IntN(math.MaxUint16))
	// 		}

	// 		encodeDecodeComparableSlice(t, want)
	// 	})
	// 	t.Run("uint32", func(t *testing.T) {
	// 		t.Parallel()

	// 		want := make([]uint32, rand.IntN(math.MaxInt8))
	// 		for i := range want {
	// 			want[i] = rand.Uint32()
	// 		}

	// 		encodeDecodeComparableSlice(t, want)
	// 	})
	// 	t.Run("uint64", func(t *testing.T) {
	// 		t.Parallel()

	// 		want := make([]uint64, rand.IntN(math.MaxInt8))
	// 		for i := range want {
	// 			want[i] = rand.Uint64()
	// 		}

	// 		encodeDecodeComparableSlice(t, want)
	// 	})
	// 	t.Run("float32", func(t *testing.T) {
	// 		t.Parallel()

	// 		want := make([]float32, rand.IntN(math.MaxInt8))
	// 		for i := range want {
	// 			want[i] = 2*rand.Float32() - 1
	// 		}

	// 		encodeDecodeComparableSlice(t, want)
	// 	})
	// 	t.Run("float64", func(t *testing.T) {
	// 		t.Parallel()

	// 		want := make([]float64, rand.IntN(math.MaxInt8))
	// 		for i := range want {
	// 			want[i] = 2*rand.Float64() - 1
	// 		}

	// 		encodeDecodeComparableSlice(t, want)
	// 	})
	// 	t.Run("complex64", func(t *testing.T) {
	// 		t.Parallel()

	// 		want := make([]complex64, rand.IntN(math.MaxInt8))
	// 		for i := range want {
	// 			want[i] = complex64(complex(2*rand.Float64()-1, 2*rand.Float64()-1))
	// 		}

	// 		encodeDecodeComparableSlice(t, want)
	// 	})
	// 	t.Run("complex128", func(t *testing.T) {
	// 		t.Parallel()

	// 		want := make([]complex128, rand.IntN(math.MaxInt8))
	// 		for i := range want {
	// 			want[i] = complex(2*rand.Float64()-1, 2*rand.Float64()-1)
	// 		}

	// 		encodeDecodeComparableSlice(t, want)
	// 	})
	// 	t.Run("string", func(t *testing.T) {
	// 		t.Parallel()

	// 		want := make([]string, rand.IntN(math.MaxInt8))
	// 		for i := range want {
	// 			want[i] = cryptorand.Text()
	// 		}

	// 		encodeDecodeComparableSlice(t, want)
	// 	})
	// 	t.Run("struct", func(t *testing.T) {
	// 		t.Parallel()

	// 		want := make([]ComparableStruct, rand.IntN(math.MaxInt8))
	// 		for i := range want {
	// 			want[i] = makeComparableStruct(t)
	// 		}

	// 		encodeDecodeComparableSlice(t, want)
	// 	})
	// })
	// t.Run("map", func(t *testing.T) {
	// 	t.Parallel()

	// 	want := make(map[uint64]float32, 10)

	// 	for range 10 {
	// 		want[rand.Uint64()] = rand.Float32()
	// 	}

	// 	d, err := Encode(want)
	// 	if err != nil {
	// 		t.Fatalf("Encode: %s", err.Error())
	// 	}

	// 	got, err := Decode[map[uint64]float32](d)
	// 	if err != nil {
	// 		t.Fatalf("Decode: %s", err.Error())
	// 	}

	// 	if len(got) != len(want) {
	// 		t.Errorf("got len %d, want len %d", len(got), len(want))
	// 	}

	// 	for k, g := range got {
	// 		w, ok := want[k]
	// 		if !ok {
	// 			t.Errorf("missing key %d", k)
	// 		}

	// 		if g != w {
	// 			t.Errorf("key %d: got %+v, want %+v", k, g, w)
	// 		}
	// 	}
	// })
}

func encodeDecodeComparable[T comparable](t *testing.T, want T) {
	t.Helper()

	d, err := Encode(want)
	if err != nil {
		t.Fatalf("Encode: %s", err.Error())
	}

	got, err := Decode[T](d)
	if err != nil {
		t.Fatalf("Decode: %s", err.Error())
	}

	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func encodeDecodeComparableSlice[T comparable](t *testing.T, want []T) {
	t.Helper()

	d, err := Encode(want)
	if err != nil {
		t.Fatalf("Encode: %s", err.Error())
	}

	got, err := Decode[[]T](d)
	if err != nil {
		t.Fatalf("Decode: %s", err.Error())
	}

	if len(got) != len(want) {
		t.Errorf("got len %d, want len %d", len(got), len(want))
	}

	for i, g := range got {
		if g != want[i] {
			t.Errorf("index %d: got %+v, want %+v", i, got, want)
		}
	}
}

func makeComparableStruct(t *testing.T) ComparableStruct {
	t.Helper()

	return ComparableStruct{
		Bool:       true,
		Int8:       int8(int16(rand.IntN(math.MaxUint8)) - math.MaxInt8),
		Int16:      int16(int32(rand.IntN(math.MaxUint16)) - math.MaxInt16),
		Int32:      rand.Int32() - math.MaxInt32/2,
		Int64:      rand.Int64() - math.MaxInt64/2,
		Uint8:      uint8(rand.UintN(math.MaxUint8)),
		Uint16:     uint16(rand.IntN(math.MaxUint16)),
		Uint32:     rand.Uint32(),
		Uint64:     rand.Uint64(),
		Float32:    2*rand.Float32() - 1,
		Float64:    2*rand.Float64() - 1,
		Complex64:  complex64(complex(2*rand.Float64()-1, 2*rand.Float64()-1)),
		Complex128: complex(2*rand.Float64()-1, 2*rand.Float64()-1),
		String:     cryptorand.Text(),
		Struct: struct {
			A int64
			B string
		}{
			A: rand.Int64() - math.MaxInt64/2,
			B: cryptorand.Text(),
		},
	}
}
