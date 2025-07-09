package pool

import "sync"

type ByteSlice struct {
	pool sync.Pool
	init bool
}

func NewByteSlice(size int) *ByteSlice {
	return &ByteSlice{
		pool: sync.Pool{
			New: func() any {
				b := make([]byte, 0, size)
				return &b
			},
		},
		init: true,
	}
}

func (p *ByteSlice) Get() *[]byte {
	return p.pool.Get().(*[]byte)
}

func (p *ByteSlice) Put(b *[]byte) {
	buf := *b
	*b = buf[:0]
	p.pool.Put(b)
}
