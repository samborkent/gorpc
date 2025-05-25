package gorpc

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/samborkent/gorpc/goc"
)

var reqPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

type Request[Req, Res any] struct {
	r *http.Request
}

func NewRequest[Req, Res any](ctx context.Context, addr string, req *Req) (*Request[Req, Res], error) {
	// Retrieve buffer from client pool, and return in upon return.
	buf := reqPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		reqPool.Put(buf)
	}()

	if err := goc.EncodeTo(buf, req); err != nil {
		return nil, fmt.Errorf("encoding request: %w", err)
	}

	methodHash := hashMethod[Req, Res]()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, addr+"/"+methodHash, buf)
	if err != nil {
		return nil, fmt.Errorf("initializing request: %w", err)
	}

	httpReq.Header.Add(HeaderMethodHash, methodHash)

	return &Request[Req, Res]{
		r: httpReq,
	}, nil
}

func Do[Req, Res any](r *Request[Req, Res]) (*Res, error) {
	httpRes, err := http.DefaultClient.Do(r.r)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}

	defer func() {
		_ = httpRes.Body.Close()
	}()

	if httpRes.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http error: %s", httpRes.Status)
	}

	res, err := goc.DecodeFrom[Res](httpRes.Body)
	if err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &res, nil
}
