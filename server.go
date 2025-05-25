package gorpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync/atomic"
)

type Server struct {
	mux     *http.ServeMux
	server  *http.Server
	running atomic.Bool
}

func NewServer(port int) *Server {
	var protocols http.Protocols
	protocols.SetUnencryptedHTTP2(true)

	return &Server{
		mux: http.NewServeMux(),
		server: &http.Server{
			Addr:      net.JoinHostPort(net.IPv4zero.String(), strconv.Itoa(port)),
			Protocols: &protocols,
		},
	}
}

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
			return nil
		case err := <-errs:
			return fmt.Errorf("server error: %w", err)
		}
	}
}

func Register[Request, Response any](s *Server, h HandlerFunc[Request, Response]) {
	if s.running.Load() {
		panic("cannot register a new handler for a running server")
	}

	s.mux.Handle("/"+h.Hash(), handler(h))
}
