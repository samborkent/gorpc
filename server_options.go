package gorpc

import (
	"errors"
	"net/http"
)

type ServerOption func(*serverConfig) error

var (
	ErrOptionDuplicate = errors.New("received duplicate options")
	ErrNilServer       = errors.New("WithHTTPServer: server nil-pointer")
)

// TODO: WithHTTPHandler/WithHTTPMiddleware
// TODO: WithMiddleware (goRPC)

// WithServerCache enables a weak response cache for the goRPC client.
func WithServerCache() ServerOption {
	return func(cfg *serverConfig) error {
		if cfg.withCache {
			return ErrOptionDuplicate
		}

		cfg.cacheResponse = true
		cfg.withCache = true

		return nil
	}
}

// WithGobServer enables gob encoding instead of goc encoding for the server.
func WithGobServer() ServerOption {
	return func(cfg *serverConfig) error {
		if cfg.withGob {
			return ErrOptionDuplicate
		}

		cfg.gob = true
		cfg.withGob = true

		return nil
	}
}

// WithHTTPServer overwrites the internal HTTP server.
func WithHTTPServer(server *http.Server) ServerOption {
	return func(cfg *serverConfig) error {
		if cfg.withHTTPServer {
			return ErrOptionDuplicate
		}

		if server == nil {
			return ErrNilServer
		}

		// TODO: additional checks such as HTTP/2 support?

		cfg.server = server
		cfg.withHTTPServer = true

		return nil
	}
}

// WithServerValidation enables server request validation.
func WithServerValidation() ServerOption {
	return func(cfg *serverConfig) error {
		if cfg.withValidation {
			return ErrOptionDuplicate
		}

		cfg.validate = true
		cfg.withValidation = true

		return nil
	}
}

type serverConfig struct {
	cacheResponse bool
	withCache     bool

	gob     bool
	withGob bool

	server         *http.Server
	withHTTPServer bool

	validate       bool
	withValidation bool
}
