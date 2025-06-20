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
	validate       bool
	withValidation bool

	server         *http.Server
	withHTTPServer bool
}
