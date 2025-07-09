package main

import (
	"bytes"
	"context"
	"log/slog"
	"math"
	"net"
	"os"
	"os/signal"
	"unsafe"
)

const (
	headerSize    = unsafe.Sizeof(Header{})
	maxPacketSize = 1214
	maxPacketNum  = uint32(math.MaxUint32 / headerSize)
)

type Packet struct {
	Header
	Data []byte
}

type Header struct {
	ConnID     uint64
	MsgID      uint64
	MsgSize    uint32
	NumPackets uint32
	Index      uint32
	Size       uint16
}

type Server struct {
	// handlers map[uint64]handlerFunc
	logger *slog.Logger
	port   int
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt)
	defer stop()

	server := &Server{
		// conns:  make(map[uint64]chan Packet),
		port:   8080,
		logger: slog.New(slog.NewJSONHandler(os.Stderr, nil)),
	}

	<-server.Listen(ctx)
}

func (s *Server) Listen(ctx context.Context) <-chan error {
	errs := make(chan error)

	go func() {
		defer close(errs)

	connLoop:
		for {
			select {
			case <-ctx.Done():
				break connLoop
			default:
				conn, err := net.ListenUDP("udp", &net.UDPAddr{
					IP:   net.IPv6zero,
					Port: s.port,
				})
				if err != nil {
					errs <- err
					continue
				}

				go s.handleConn(ctx, conn)
			}
		}
	}()

	return errs
}

func (s *Server) handleConn(ctx context.Context, conn *net.UDPConn) {
	defer func() {
		if r := recover(); r != nil {
			s.logger.ErrorContext(ctx, r.(error).Error())
		}
	}()

	defer func() {
		if err := conn.Close(); err != nil {
			s.logger.WarnContext(ctx, "closing connection: "+err.Error())
		}
	}()

	// TODO: use sync.Pool
	buf := &bytes.Buffer{}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			var headerBytes [headerSize]byte

			_, err := conn.Read(headerBytes[:])
			if err != nil {
				panic(err)
			}

			header := *(*Header)(unsafe.Pointer(&headerBytes[0]))

			// Ignore empty / invalid packets.
			if header.ConnID == 0 ||
				header.MsgID == 0 ||
				header.MsgSize == 0 ||
				header.NumPackets == 0 ||
				header.NumPackets > maxPacketNum ||
				header.Index > header.NumPackets-1 ||
				header.Size == 0 ||
				header.Size > maxPacketSize {
				return
			}

			// TODO: use sync.Pool
			data := make([]byte, header.Size)

			_, err = conn.Read(data)
			if err != nil {
				panic(err)
			}

			buf.Write(data)

			if buf.Len() == int(header.MsgSize) {
			}
		}
	}
}

// type handlerFunc func(ctx context.Context, conn *net.UDPConn, buf *bytes.Buffer) error
