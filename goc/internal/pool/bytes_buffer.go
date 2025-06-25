package pool

import (
	"bytes"
	"sync"
)

type BytesBuffer struct {
	pool sync.Pool
	init bool
}

func NewBytesBuffer() *BytesBuffer {
	return &BytesBuffer{
		pool: sync.Pool{
			New: func() any {
				return new(bytes.Buffer)
			},
		},
		init: true,
	}
}

func (p *BytesBuffer) Get() *bytes.Buffer {
	if p == nil || !p.init {
		return new(bytes.Buffer)
	}

	return p.pool.Get().(*bytes.Buffer)
}

func (p *BytesBuffer) Put(b *bytes.Buffer) {
	if b == nil {
		return
	}

	b.Reset()

	if p == nil || !p.init {
		return
	}

	p.pool.Put(b)
}
