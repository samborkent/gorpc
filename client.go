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
}

func NewClient[Request, Response any](addr string) *Client[Request, Response] {
	protocols := new(http.Protocols)
	protocols.SetUnencryptedHTTP2(true)

	hash := hashMethod[Request, Response]()

	return &Client[Request, Response]{
		client: &http.Client{
			Transport: &http.Transport{
				ForceAttemptHTTP2: true,
				Protocols:         protocols,
			},
		},
		addr: strings.TrimRight(addr, "/") + "/" + hash,
		hash: hash,
	}
}

func (c *Client[Request, Response]) Do(ctx context.Context, req *Request) (*Response, error) {
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
