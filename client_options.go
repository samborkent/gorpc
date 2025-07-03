package gorpc

import (
	"errors"
	"net/http"
)

type ClientOption func(*clientConfig) error

var ErrClientNil = errors.New("WithHTTPClient: client nil-pointer")

// WithCache enables a weak response cache for the goRPC client.
func WithCache() ClientOption {
	return func(cfg *clientConfig) error {
		if cfg.withCache {
			return ErrOptionDuplicate
		}

		cfg.cacheResponse = true
		cfg.withCache = true

		return nil
	}
}

// WithClientValidation enables request and response validation for the goRPC client.
func WithClientValidation() ClientOption {
	return func(cfg *clientConfig) error {
		if cfg.withValidation {
			return ErrOptionDuplicate
		}

		cfg.validate = true
		cfg.withValidation = true

		return nil
	}
}

// WithHTTPClient allows the goRPC internal HTTP client to overwritten with a custom client.
func WithHTTPClient(client *http.Client) ClientOption {
	return func(cfg *clientConfig) error {
		if cfg.withHTTPClient {
			return ErrOptionDuplicate
		}

		if client == nil {
			return ErrClientNil
		}

		cfg.client = client
		cfg.withHTTPClient = true

		return nil
	}
}

// WithClientMethod specifies the HTTP method to use for the request. Default is POST.
func WithClientMethod(method Method) ClientOption {
	return func(cfg *clientConfig) error {
		if cfg.witClienthMethod {
			return ErrOptionDuplicate
		}

		cfg.method = method
		cfg.witClienthMethod = true

		return nil
	}
}

type clientConfig struct {
	cacheResponse bool
	withCache     bool

	client         *http.Client
	withHTTPClient bool

	method           Method
	witClienthMethod bool

	validate       bool
	withValidation bool
}

var defaultClientConfig = clientConfig{
	method: MethodPost,
}
