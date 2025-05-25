# goc

goc is a Go-only encoding method inspired by gob and Protobuf.

goc works in a similar way to gob, but it is not self-describing. Meaning both the sender and the receiver need to be aware of the sturcture of the data.
This makes it ideal to work with the strictly Go-typed RPC method: goRPC.
