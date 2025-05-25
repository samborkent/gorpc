package gorpc

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"hash/fnv"
	"reflect"
)

// Concatenate bytes of request name, request size, response name, and response size.
// Create 128-bit FNV-1a hash. Return hex encoding of hash.
func hashMethod[Request, Response any]() string {
	req := reflect.TypeOf(*new(Request))
	res := reflect.TypeOf(*new(Response))

	buf := new(bytes.Buffer)

	_, _ = buf.WriteString(req.Name())

	var d [4]byte
	binary.BigEndian.PutUint32(d[:], uint32(req.Size()))
	_, _ = buf.Write(d[:])

	_, _ = buf.WriteString(res.Name())

	binary.BigEndian.PutUint32(d[:], uint32(res.Size()))
	_, _ = buf.Write(d[:])

	hsh := fnv.New128a()
	_, _ = hsh.Write(buf.Bytes())

	return hex.EncodeToString(hsh.Sum(nil))
}
