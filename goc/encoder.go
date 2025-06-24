package goc

import (
	"bytes"
	"io"
)

// Encoder implements a goc-encoder which returns the encoded object.
type Encoder interface {
	Encode() (encoded []byte, err error)
}

// EncodeWriter implements a goc-encoder which writes the encoded object to [io.Writer].
type EncodeWriter interface {
	EncodeWrite(w io.Writer) (n int, err error)
}

// EncodeWriter implements a goc-encoder which writes the encoded object to a [*bytes.Buffer].
type BufferedEncoder interface {
	EncodeBuffer(buf *bytes.Buffer) error
}

// Encode goc-encodes an object and returns the encoded object.
func Encode[T any](in T) (out []byte, err error) {
	buf := bytesBufferPool.Get().(*bytes.Buffer)
	defer func() {
		bytesBufferPool.Put(buf)
	}()

	if err := encodeBuffer(buf, in); err != nil {
		return nil, err
	}

	// Create a copy to so we can safely return *bytes.Buffer to pool.
	encoded := make([]byte, buf.Len())
	copy(encoded, buf.Bytes())

	return encoded, nil
}

// EncodeWrite goc-encodes an object and writed it to [io.Writer].
func EncodeWrite[T any](out io.Writer, in T) (n int, err error) {
	buf := bytesBufferPool.Get().(*bytes.Buffer)
	defer func() {
		bytesBufferPool.Put(buf)
	}()

	if err := encodeBuffer(buf, in); err != nil {
		return 0, err
	}

	return out.Write(buf.Bytes())
}

// EncodeBuffer goc-encodes an object and writes it to [*bytes.Buffer].
func EncodeBuffer[T any](out *bytes.Buffer, in T) error {
	return encodeBuffer(out, in)
}
