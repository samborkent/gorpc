package goc

import (
	"encoding"
	"fmt"
	"reflect"
)

var (
	reflectEncodeByteWriter   = reflect.TypeFor[EncodeByteWriter]()
	reflectDecodeByteReader   = reflect.TypeFor[DecodeByteReader]()
	reflectEncodeWriter       = reflect.TypeFor[EncodeWriter]()
	reflectDecodeReader       = reflect.TypeFor[DecodeReader]()
	reflectEncoder            = reflect.TypeFor[Encoder]()
	reflectDecoder            = reflect.TypeFor[Decoder]()
	reflectBinaryMarshaller   = reflect.TypeFor[encoding.BinaryMarshaler]()
	reflectBinaryUnmarshaller = reflect.TypeFor[encoding.BinaryUnmarshaler]()
	reflectError              = reflect.TypeFor[error]()
	reflectStringer           = reflect.TypeFor[fmt.Stringer]()
)
