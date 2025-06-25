package goc

import (
	"bytes"
	"io"
)

// Encoder implements a goc-encoder which returns the encoded object.
type Encoder interface {
	Encode() (out []byte, err error)
}

// EncodeWriter implements a goc-encoder which writes the encoded object to [io.Writer].
type EncodeWriter interface {
	EncodeWrite(out io.Writer) (n int, err error)
}

// EncodeWriter implements a goc-encoder which writes the encoded object to a [*bytes.Buffer].
type EncodeByteWriter interface {
	EncodeBuffer(out *bytes.Buffer) error
}

// Encode goc-encodes an object and returns the encoded object.
func Encode[T any](in T, options ...Option) (out []byte, err error) {
	var cfg config
	for _, option := range options {
		option(&cfg)
	}

	buf := encodingPool.Get()
	defer encodingPool.Put(buf)

	if err := encodeBuffer(buf, in, cfg.allowUnsafe); err != nil {
		return nil, err
	}

	// Create a copy to so we can safely return *bytes.Buffer to pool.
	encoded := make([]byte, buf.Len())
	copy(encoded, buf.Bytes())

	return encoded, nil
}

// EncodeWrite goc-encodes an object and writes it to [io.Writer].
func EncodeWrite[T any](out io.Writer, in T, options ...Option) (n int, err error) {
	var cfg config
	for _, option := range options {
		option(&cfg)
	}

	buf := encodingPool.Get()
	defer encodingPool.Put(buf)

	if err := encodeBuffer(buf, in, cfg.allowUnsafe); err != nil {
		return 0, err
	}

	return out.Write(buf.Bytes())
}

// EncodeByteWrite goc-encodes an object and writes it to [*bytes.Buffer].
func EncodeByteWrite[T any](out *bytes.Buffer, in T, options ...Option) error {
	var cfg config
	for _, option := range options {
		option(&cfg)
	}

	return encodeBuffer(out, in, cfg.allowUnsafe)
}
