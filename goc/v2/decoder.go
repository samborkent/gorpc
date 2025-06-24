package goc

import (
	"bytes"
	"io"
)

// Decoder implements a goc-decoder which reads the byte slice and decodes into the [Decoder] object.
type Decoder interface {
	Decode(in []byte) error
}

// DecodeReader implements a goc-decoder which reads from the [io.Reader] decodes into the [DecodeReader] object.
type DecodeReader interface {
	DecodeRead(in io.Reader) error
}

// DecodeByteReader implements a goc-decoder which reads from the [*bytes.Reader] decodes into the [DecodeReader] object.
type DecodeByteReader interface {
	DecodeByteRead(in *bytes.Reader) error
}

// Decode goc-decodes a byte slice into an object.
func Decode[T any](in []byte, out *T) error {
	return decodeReader(bytes.NewReader(in), out)
}

// DecodeRead reads from an [io.Reader] and goc-decodes into an object.
func DecodeRead[T any](in io.Reader, out *T) error {
	switch t := in.(type) {
	case *bytes.Reader:
		return decodeReader(t, out)
	case *bytes.Buffer:
		return decodeReader(bytes.NewReader(t.Bytes()), out)
	default:
		buf := bytesBufferPool.Get().(*bytes.Buffer)
		defer func() {
			bytesBufferPool.Put(buf)
		}()

		_, err := buf.ReadFrom(in)
		if err != nil {
			return err
		}

		return decodeReader(bytes.NewReader(buf.Bytes()), out)
	}
}

// DecodeByteRead reads from an [*bytes.Reader] and goc-decodes into an object.
func DecodeByteRead[T any](in *bytes.Reader, out *T) error {
	return decodeReader(in, out)
}
