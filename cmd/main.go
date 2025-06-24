package main

import (
	"bytes"
	"context"
	cryptorand "crypto/rand"
	mathrand "math/rand/v2"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"time"

	"github.com/samborkent/gorpc/goc"
)

func main() {
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt)
	defer stop()

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

			if err := goc.EncodeBuffer(buf, object); err != nil {
				panic(err.Error())
			}

			time.Sleep(time.Millisecond)
		}
	}
}
