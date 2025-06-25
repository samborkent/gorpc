package goc

import "github.com/samborkent/gorpc/internal/pool"

var (
	encodingPool = pool.NewBytesBuffer()
	decodingPool = pool.NewBytesBuffer()
)
