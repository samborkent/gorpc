package gorpc

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"hash/maphash"
	"io"
	"net/http"
	"strings"
	"weak"

	"github.com/samborkent/gorpc/goc"
	"github.com/samborkent/gorpc/internal/pool"
	isync "github.com/samborkent/gorpc/internal/sync"
)

var clientPool = pool.NewBytesBuffer()

type Client[Request, Response any] struct {
	cache                           isync.Map[uint64, weak.Pointer[Response]]
	client                          *http.Client
	addr, hash                      string
	method                          Method
	seed                            maphash.Seed
	cacheResponse, useGob, validate bool
}

func NewClient[Request, Response any](addr string, options ...ClientOption) (*Client[Request, Response], error) {
	cfg := defaultClientConfig
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

	// Trim trailing slashes and append method hash.
	addr = strings.TrimRight(addr, "/") + "/" + hash

	// TODO: enable once TLS support is implemented.
	// // Ensure address if a valid HTTP/S address.
	// addr, found := strings.CutPrefix(addr, "http://")
	// if found || !strings.HasPrefix(addr, "httsp://") {
	// 	addr = "https://" + addr
	// }

	return &Client[Request, Response]{
		client:        client,
		addr:          addr,
		hash:          hash,
		seed:          maphash.MakeSeed(),
		method:        cfg.method,
		cacheResponse: cfg.cacheResponse,
		useGob:        cfg.gob,
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
	var (
		body io.Reader
		data []byte
		err  error
	)

	if c.useGob {
		buf := clientPool.Get()
		defer clientPool.Put(buf)

		if err := gob.NewEncoder(buf).Encode(req); err != nil {
			return nil, fmt.Errorf("encoding request: %w", err)
		}

		body = buf
		data = buf.Bytes()
	} else {
		// TODO: use []byte pool
		data, err = goc.Encode(req)
		if err != nil {
			return nil, fmt.Errorf("encoding request: %w", err)
		}

		body = bytes.NewReader(data)
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

	httpReq, err := http.NewRequestWithContext(ctx, string(c.method), c.addr, body)
	if err != nil {
		return nil, fmt.Errorf("initializing request: %w", err)
	}

	if c.useGob {
		httpReq.Header.Add(HeaderAccept, MIMETypeGob)
		httpReq.Header.Add(HeaderContentType, MIMETypeGob)
	} else {
		httpReq.Header.Add(HeaderAccept, MIMETypeGoc)
		httpReq.Header.Add(HeaderContentType, MIMETypeGoc)
	}

	httpReq.Header.Add(HeaderMethodHash, c.hash)
	httpReq.ContentLength = int64(len(data))

	httpRes, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}

	if httpRes.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("http error: %s", httpRes.Status)
	}

	var res Response

	if c.useGob {
		err = gob.NewDecoder(httpRes.Body).Decode(&res)
	} else {
		err = goc.DecodeRead(httpRes.Body, &res)
	}

	_ = httpRes.Body.Close()

	if err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	if c.cacheResponse && payloadHash != 0 {
		_ = c.cache.CompareAndSwap(payloadHash, cachedResponse, weak.Make(&res))
	}

	return &res, nil
}
