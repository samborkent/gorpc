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
}

const (
	portMin = 49152
	portMax = math.MaxUint16
)

// NewServer retuns a new goRPC [Server].
func NewServer(port int) *Server {
	// Use a random port.
	if port <= 0 {
		port = portMin + rand.IntN(portMax-portMin)
	}

	protocols := new(http.Protocols)
	protocols.SetUnencryptedHTTP2(true)

	return &Server{
		mux: http.NewServeMux(),
		server: &http.Server{
			Addr:      ":" + strconv.Itoa(port),
			Protocols: protocols,
		},
		port: port,
	}
}

// Register registers a [HandlerFunc] to a goRPC [Server]. Panics when the server is already running.
func Register[Request, Response any](s *Server, h HandlerFunc[Request, Response]) {
	if s.running.Load() {
		panic("goRPC: cannot register a new handler for a running server")
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
