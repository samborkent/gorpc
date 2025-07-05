package gorpc

import (
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
	seed                            maphash.Seed
	cacheResponse, useGob, validate bool
	roundTripper                    RoundTripperFunc[Request, Response]
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

	// Trim trailing slashes and append method hash.
	addr = strings.TrimRight(addr, "/") + "/" + hash

	// TODO: enable once TLS support is implemented.
	// // Ensure address if a valid HTTP/S address.
	// addr, found := strings.CutPrefix(addr, "http://")
	// if found || !strings.HasPrefix(addr, "httsp://") {
	// 	addr = "https://" + addr
	// }

	c := &Client[Request, Response]{
		client:        client,
		addr:          addr,
		hash:          hash,
		seed:          maphash.MakeSeed(),
		cacheResponse: cfg.cacheResponse,
		useGob:        cfg.gob,
		validate:      cfg.validate,
	}

	c.roundTripper = c.do

	return c, nil
}

// RegisterRoundTripper registers a custom round tripper for a given client.
func RegisterRoundTripper[Request, Response any](client *Client[Request, Response], roundtrippers ...RoundTripper[Request, Response]) {
	for _, roundTripper := range roundtrippers {
		client.roundTripper = roundTripper(client.roundTripper)
	}
}

func (c *Client[Request, Response]) Do(ctx context.Context, req *Request) (*Response, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if c.validate {
		return ValidationRoundTripper(c.roundTripper)(ctx, req)
	}

	return c.roundTripper(ctx, req)
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
		buf := clientPool.Get()
		defer clientPool.Put(buf)

		if err := goc.EncodeByteWrite(buf, req); err != nil {
			return nil, fmt.Errorf("encoding request: %w", err)
		}

		body = buf
		data = buf.Bytes()
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
			response := cachedResponse.Value()

			if response != nil {
				return response, nil
			}
		}
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.addr, body)
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
