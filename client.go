package gorpc

import (
	"bytes"
	"context"
	"fmt"
	"hash/maphash"
	"net/http"
	"strings"
	"weak"

	"github.com/samborkent/gorpc/goc"
	isync "github.com/samborkent/gorpc/internal/sync"
)

type Client[Request, Response any] struct {
	client                  *http.Client
	addr, hash              string
	cache                   isync.Map[uint64, weak.Pointer[Response]]
	seed                    maphash.Seed
	cacheResponse, validate bool
}

func NewClient[Request, Response any](addr string, options ...ClientOption) (*Client[Request, Response], error) {
	cfg := clientConfig{}
	for _, option := range options {
		if err := option(&cfg); err != nil {
			return nil, err
		}
	}

	hash := hashMethod[Request, Response]()

	var client *http.Client

	if cfg.withHTTPClient {
		client = cfg.client
	} else {
		client = &http.Client{
			Transport: httpRoundTripper,
		}
	}

	return &Client[Request, Response]{
		client:        client,
		addr:          strings.TrimRight(addr, "/") + "/" + hash,
		hash:          hash,
		seed:          maphash.MakeSeed(),
		cacheResponse: cfg.cacheResponse,
		validate:      cfg.validate,
	}, nil
}

func (c *Client[Request, Response]) Do(ctx context.Context, req *Request) (*Response, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if c.validate {
		return ValidationRoundTripper(c.do)(ctx, req)
	}

	return c.do(ctx, req)
}

func (c *Client[Request, Response]) do(ctx context.Context, req *Request) (*Response, error) {
	// TODO: use []byte pool
	data, err := goc.Encode(req)
	if err != nil {
		return nil, fmt.Errorf("encoding request: %w", err)
	}

	var (
		cachedResponse weak.Pointer[Response]
		payloadHash    uint64
	)

	if c.cacheResponse {
		payloadHash = maphash.Bytes(c.seed, data)

		var ok bool

		cachedResponse, ok = c.cache.Load(payloadHash)
		if ok && cachedResponse.Value() != nil {
			// TODO: resolve race-condition
			return cachedResponse.Value(), nil
		}
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.addr, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("initializing request: %w", err)
	}

	httpReq.Header.Add(HeaderAccept, MIMEType)
	httpReq.Header.Add(HeaderContentType, MIMEType)
	httpReq.Header.Add(HeaderMethodHash, c.hash)
	httpReq.ContentLength = int64(len(data))

	httpRes, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}

	if httpRes.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("http error: %s", httpRes.Status)
	}

	res, err := goc.DecodeFrom[Response](httpRes.Body)
	_ = httpRes.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	if c.cacheResponse && payloadHash != 0 {
		_ = c.cache.CompareAndSwap(payloadHash, cachedResponse, weak.Make(&res))
	}

	return &res, nil
}
