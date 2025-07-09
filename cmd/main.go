package main

import (
	"bytes"
	"context"
	cryptorand "crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	mathrand "math/rand/v2"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"time"

	"github.com/samborkent/gorpc/goc"
)

type Message struct {
	len  uint32
	data []byte
}

func NewMessage(data []byte) (Message, error) {
	if len(data) > math.MaxUint32 {
		return Message{}, errors.New("message exceeds max length")
	}

	return Message{
		len:  uint32(len(data)),
		data: data,
	}, nil
}

func (m Message) Serialize() []byte {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], m.len)
	return append(b[:], m.data...)
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt)
	defer stop()

	dialer := net.Dialer{
		Timeout: time.Second,
	}

	port := 8080

	conn, err := dialer.DialContext(ctx, "tcp", "127.0.0.1:8080")
	if err != nil {
		panic(err)
	}

	var msg Message

	conn.Write(msg.Serialize())

	listener, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.IPv4zero,
		Port: port,
	})
	if err != nil {
		panic(err)
	}
	defer listener.Close()

acceptLoop:
	for {
		select {
		case <-ctx.Done():
			break acceptLoop
		default:
			conn, err := listener.AcceptTCP()
			if err != nil {
				panic(err)
			}

			go func() {
				defer func() {
					if r := recover(); r != nil {
						switch rec := r.(type) {
						case string:
							fmt.Println(rec)
						case error:
							fmt.Println(rec.Error())
						case fmt.Stringer:
							fmt.Println(rec.String())
						}
					}
				}()

				defer func() {
					if err := conn.Close(); err != nil {
						panic(err)
					}
				}()

				buf := bytes.Buffer{}

			readLoop:
				for {
					var packet [math.MaxUint16]byte

					stopReading := false

					n, err := conn.Read(packet[:])
					if err != nil {
						if errors.Is(err, io.EOF) {
							stopReading = true
						} else {
							panic(err)
						}
					}

					if n < len(packet) {
						stopReading = true
					}

					_, err = buf.Write(packet[:n])
					if err != nil {
						panic(err)
					}

					if stopReading {
						break readLoop
					}
				}
			}()
		}
	}

	type Object struct {
		Num   uint64
		Str   string
		Slice []int32
		Map   map[string]string
	}

	object := Object{
		Num: mathrand.Uint64(),
		Str: cryptorand.Text(),
		Slice: []int32{
			mathrand.Int32(),
			mathrand.Int32(),
			mathrand.Int32(),
		},
		Map: map[string]string{
			cryptorand.Text(): cryptorand.Text(),
			cryptorand.Text(): cryptorand.Text(),
			cryptorand.Text(): cryptorand.Text(),
			cryptorand.Text(): cryptorand.Text(),
		},
	}

	server := &http.Server{Addr: "localhost:6060"}

	go func() {
		_ = server.ListenAndServe()
	}()

	for {
		select {
		case <-ctx.Done():
			_ = server.Close()

			return
		default:
			buf := new(bytes.Buffer)

			if err := goc.EncodeByteWrite(buf, object); err != nil {
				panic(err.Error())
			}

			var out Object

			if err := goc.DecodeByteRead(bytes.NewReader(buf.Bytes()), &out); err != nil {
				panic(err.Error())
			}

			time.Sleep(time.Millisecond)
		}
	}
}
