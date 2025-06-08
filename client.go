package gorpc

import (
	"bytes"
	"context"
	"fmt"
	"hash/maphash"
	"net/http"
	"strings"
	"sync"
	"weak"

	"github.com/samborkent/gorpc/goc"
)

type Client[Request, Response any] struct {
	client     *http.Client
	addr, hash string
	// TODO: use sync.Map?
	cache map[uint64]weak.Pointer[Response]
	cacheLock *sync.RWMutex
	seed maphash.Seed
	cacheResponse, validate bool
}

func NewClient[Request, Response any](addr string, options ...ClientOption) *Client[Request, Response] {
	cfg := clientConfig{}
	for _, option := range options {
		option(&cfg)
	}

	hash := hashMethod[Request, Response]()

	return &Client[Request, Response]{
		client: &http.Client{
			// TODO: fix this, this should be a *http.Transport
			Transport: &httpDefaultTransport,
		},
		addr: strings.TrimRight(addr, "/") + "/" + hash,
		hash: hash,
		cache: m,
		cacheLock: new(sync.RWMutex),
		seed: maphash.MakeSeed(),
		cacheResponse: cfg.cacheResponse,
		validate: cfg.validate,
	}
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

	var payloadHash uint64

	if c.cacheResponse {
		payloadHash = maphash.Bytes(c.seed, data)
	
		// TODO: get rid of lock
		c.cacheLock.RLock()
		res, ok := c.cache[payloadHash]
		c.cacheLock.RUnlock()

		if ok && res.Value != nil {
			// TODO: resolve race-condition
			return res.Value, nil
		}
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.addr, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("initializing request: %w", err)
	}

	httpReq.Header.Add(HeaderContentType, MIMEType)
	httpReq.Header.Add(HeaderMethodHash, c.hash)
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

	if c.cacheResponse && payloadHash != 0 {
		weakRef := weak.Make(res)
	
		c.cacheLock.Lock()
		c.cache[payloadHash] = weakRef
		c.cacheLock.Unlock()
	}

	return &res, nil
}
