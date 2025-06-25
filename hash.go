package gorpc

import (
	"encoding/binary"
	"encoding/hex"
	"hash/fnv"
	"reflect"

	"github.com/samborkent/gorpc/internal/pool"
)

var hashPool = pool.NewBytesBuffer()

// Concatenate bytes of request name, request size, response name, and response size.
// Create 128-bit FNV-1a hash. Return hex encoding of hash.
func hashMethod[Request, Response any]() string {
	req := reflect.TypeFor[Request]()
	res := reflect.TypeFor[Response]()

	buf := hashPool.Get()

	// Write request name.
	_, _ = buf.WriteString(req.String())

	// Write request size.
	var d [4]byte
	binary.BigEndian.PutUint32(d[:], uint32(req.Size()))
	_, _ = buf.Write(d[:])

	// Write response name.
	_, _ = buf.WriteString(res.String())

	// Write response size.
	binary.BigEndian.PutUint32(d[:], uint32(res.Size()))
	_, _ = buf.Write(d[:])

	// Hash as 128-bit FNV-1a hash.
	hsh := fnv.New128a()
	_, _ = hsh.Write(buf.Bytes())
	hashPool.Put(buf)

	// Return as hex-encoded string.
	return hex.EncodeToString(hsh.Sum(nil))
}
