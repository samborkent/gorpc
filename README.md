# goRPC

Like gRPC, but then only for Go.

goRPC is a dependency free, HTTP/2 based RPC protocol that makes use of goc encoding.

## goc

goc is a Go-only encoding method inspired by gob and Protobuf.

# Ideas

* Optimizations:
    - Encode map as []key + []val, to more efficiently encode/decode values.
* Implement unsafe support to enable super fast optimizations.
* Small values option.
    - In this option max len is set to uint16 instead.
    - Deflate/gzip compression is enabled by default.
