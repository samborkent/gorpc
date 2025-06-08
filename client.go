package gorpc

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/samborkent/gorpc/goc"
)

type Client[Request, Response any] struct {
	client     *http.Client
	addr, hash string
	validate bool
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

	return &res, nil
}
