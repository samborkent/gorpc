package gorpc

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"net/http"
	"strconv"
	"sync/atomic"
)

// Server implements a goRPC server.
type Server struct {
	mux     *http.ServeMux
	server  *http.Server
	running atomic.Bool
	port    int
	validate bool
}

const (
	portMin = 49152
	portMax = math.MaxUint16
)

// NewServer retuns a new goRPC [Server].
func NewServer(port int, options ...ServerOption) *Server {
	// Use a random port.
	if port <= 0 {
		port = portMin + rand.IntN(portMax-portMin)
	}

	cfg := serverConfig{}
	for _, option := range options {
		if err := option(&cfg); err != nil {
			return err
		}
	}

	var server *http.Server

	if cfg.withHTTPServer {
		server = cfg.server
	} else {
		server = &http.Server{
			Addr:      ":" + strconv.Itoa(port),
			Protocols: httpProtocols,
		}
	}

	return &Server{
		mux: http.NewServeMux(),
		server: server,
		port: port,
		validate: cfg.validate,
	}
}

// Register registers a [HandlerFunc] to a goRPC [Server]. Panics when the server is already running.
func Register[Request, Response any](s *Server, h HandlerFunc[Request, Response]) {
	if s.running.Load() {
		panic("goRPC: cannot register a new handler for a running server")
	}

	if s.validate {
		h = ValidationMiddleware(h)
	}

	s.mux.Handle("POST /"+h.Hash(), handler(h))
}

// Addr returns the server address.
func (s *Server) Addr() string {
	return s.server.Addr
}

// Port returns the server port.
func (s *Server) Port() int {
	return s.port
}

// Start starts a goRPC server.
func (s *Server) Start(ctx context.Context) error {
	s.running.Store(true)
	defer s.running.Store(false)

	s.server.Handler = s.mux

	errs := make(chan error, 1)
	defer close(errs)

	go func() {
		err := s.server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errs <- err
		}
	}()

	for {
		select {
		case <-ctx.Done():
			// TODO: use graceful shutdown
			return s.server.Close()
		case err := <-errs:
			return fmt.Errorf("goRPC: server error: %w", err)
		}
	}
}
