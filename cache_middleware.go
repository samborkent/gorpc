package gorpc

import (
	"context"
	"hash/maphash"
)

func CacheMiddleware[Request, Response any](next HandlerFunc[Request, Response]) HandlerFunc[Request, Response] {
	seed := maphash.MakeSeed()

	return func(ctx context.Context, req *Request) (*Response, error) {
		
	}
}
