package gorpc

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"sync"

	"github.com/samborkent/gorpc/goc"
)

type Client[Request, Response any] struct {
	client     *http.Client
	pool       sync.Pool
	addr, hash string
}

func NewClient[Request, Response any](addr string) *Client[Request, Response] {
	protocols := new(http.Protocols)
	protocols.SetHTTP1(true)
	protocols.SetHTTP2(true)
	protocols.SetUnencryptedHTTP2(true)

	hash := hashMethod[Request, Response]()

	return &Client[Request, Response]{
		client: &http.Client{
			Transport: &http.Transport{
				// ForceAttemptHTTP2: true,
				Protocols: protocols,
			},
		},
		pool: sync.Pool{
			New: func() any {
				return bytes.NewBuffer(make([]byte, goc.Size(reflect.ValueOf(*new(Request)))))
			},
		},
		addr: strings.TrimRight(addr, "/") + "/" + hash,
		hash: hash,
	}
}

func (c *Client[Request, Response]) Do(ctx context.Context, req *Request) (*Response, error) {
	data, err := goc.Encode(req)
	if err != nil {
		panic(err)
	}

	// // Retrieve buffer from client pool, and return in upon return.
	// buf := c.pool.Get().(*bytes.Buffer)
	// defer func() {
	// 	buf.Reset()
	// 	c.pool.Put(buf)
	// }()

	// slog.InfoContext(ctx, "request", slog.Any("req", req))

	// if err := goc.EncodeTo(buf, req); err != nil {
	// 	return nil, fmt.Errorf("encoding request: %w", err)
	// }

	// data, err := io.ReadAll(buf)
	// if err != nil {
	// 	panic(err)
	// }

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.addr, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("initializing request: %w", err)
	}

	httpReq.Header.Add(HeaderContentType, MIMEType)
	httpReq.Header.Add(HeaderMethodHash, c.hash)
	// httpReq.ContentLength = int64(buf.Len())
	httpReq.ContentLength = int64(len(data))

	httpRes, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}

	defer func() { _ = httpRes.Body.Close() }()

	if httpRes.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("http error: %s", httpRes.Status)
	}

	res, err := goc.DecodeFrom[Response](httpRes.Body)
	if err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &res, nil
}
