package gorpc_test

import (
	"bytes"
	"context"
	cryptorand "crypto/rand"
	jsonv1 "encoding/json"
	"encoding/json/v2"
	"errors"
	"math"
	mathrand "math/rand/v2"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/samborkent/gorpc"
)

type Object struct {
	Num   uint64
	Str   string
	Slice []int32
	Map   map[string]string
}

func BenchmarkHTTPJSONV1(b *testing.B) {
	ctx := b.Context()

	buf := new(bytes.Buffer)
	serverBuffer := new(bytes.Buffer)
	encoder := jsonv1.NewEncoder(buf)

	port := randPort()
	server := &http.Server{
		Addr: ":" + strconv.Itoa(port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer serverBuffer.Reset()

			var request Object

			if err := jsonv1.NewDecoder(r.Body).Decode(&request); err != nil {
				http.Error(w, "json decoding", http.StatusInternalServerError)
				return
			}

			if err := jsonv1.NewEncoder(serverBuffer).Encode(newObject()); err != nil {
				http.Error(w, "json encoding", http.StatusInternalServerError)
				return
			}

			_, err := w.Write(serverBuffer.Bytes())
			if err != nil {
				http.Error(w, "writing", http.StatusInternalServerError)
				return
			}
		}),
	}

	client := &http.Client{
		Timeout: time.Second,
	}

	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			b.Error("server error: " + err.Error())
		}
	}()

	b.Cleanup(func() {
		if err := server.Shutdown(b.Context()); err != nil {
			b.Log("error closing server: " + err.Error())
		}
	})

	for b.Loop() {
		buf.Reset()

		if err := encoder.Encode(newObject()); err != nil {
			b.Fatal("encoding error: " + err.Error())
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://127.0.0.1:"+strconv.Itoa(port), buf)
		if err != nil {
			b.Fatal("creating request error: " + err.Error())
		}

		res, err := client.Do(req)
		if err != nil {
			b.Fatal("sending request error: " + err.Error())
		}

		if res.StatusCode != http.StatusOK {
			b.Fatal("http error: " + res.Status)
		}

		var result Object

		if err := jsonv1.NewDecoder(res.Body).Decode(&result); err != nil {
			b.Fatal("decoding response error: " + err.Error())
		}

		if err := res.Body.Close(); err != nil {
			b.Error("closing response body error: " + err.Error())
		}
	}
}

func BenchmarkHTTPJSONV2(b *testing.B) {
	ctx := b.Context()

	buf := new(bytes.Buffer)
	serverBuffer := new(bytes.Buffer)

	port := randPort()
	server := &http.Server{
		Addr: ":" + strconv.Itoa(port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer serverBuffer.Reset()

			var result Object

			if err := json.UnmarshalRead(r.Body, &result); err != nil {
				http.Error(w, "json decoding", http.StatusInternalServerError)
				return
			}

			if err := json.MarshalWrite(serverBuffer, newObject()); err != nil {
				http.Error(w, "json encoding", http.StatusInternalServerError)
				return
			}

			_, err := w.Write(serverBuffer.Bytes())
			if err != nil {
				http.Error(w, "writing", http.StatusInternalServerError)
				return
			}
		}),
	}

	client := &http.Client{
		Timeout: time.Second,
	}

	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			b.Error("server error: " + err.Error())
		}
	}()

	b.Cleanup(func() {
		if err := server.Shutdown(b.Context()); err != nil {
			b.Log("error closing server: " + err.Error())
		}
	})

	for b.Loop() {
		buf.Reset()

		if err := json.MarshalWrite(buf, newObject()); err != nil {
			b.Fatal("encoding error: " + err.Error())
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://127.0.0.1:"+strconv.Itoa(port), buf)
		if err != nil {
			b.Fatal("creating request error: " + err.Error())
		}

		res, err := client.Do(req)
		if err != nil {
			b.Fatal("sending request error: " + err.Error())
		}

		if res.StatusCode != http.StatusOK {
			b.Fatal("http error: " + res.Status)
		}

		var result Object

		if err := json.UnmarshalRead(res.Body, &result); err != nil {
			b.Fatal("decoding response error: " + err.Error())
		}

		if err := res.Body.Close(); err != nil {
			b.Error("closing response body error: " + err.Error())
		}
	}
}

func BenchmarkGoRPC(b *testing.B) {
	ctx := b.Context()

	port := randPort()

	server, err := gorpc.NewServer(port)
	if err != nil {
		b.Fatal("new server error: " + err.Error())
	}

	gorpc.RegisterHandler(server, gorpc.HandlerFunc[Object, Object](func(ctx context.Context, req *Object) (*Object, error) {
		object := newObject()
		return &object, nil
	}))

	client, err := gorpc.NewClient[Object, Object]("http://127.0.0.1:" + strconv.Itoa(port))
	if err != nil {
		b.Fatal("new client error: " + err.Error())
	}

	go func() {
		if err := server.Start(ctx); err != nil {
			b.Error("server error: " + err.Error())
		}
	}()

	for b.Loop() {
		request := newObject()
		response, err := client.Do(ctx, &request)
		if err != nil {
			b.Fatal("request error: " + err.Error())
		}

		_ = response
	}
}

func BenchmarkGoRPCGob(b *testing.B) {
	ctx := b.Context()

	port := randPort()

	server, err := gorpc.NewServer(port, gorpc.WithGobServer())
	if err != nil {
		b.Fatal("new server error: " + err.Error())
	}

	gorpc.RegisterHandler(server, gorpc.HandlerFunc[Object, Object](func(ctx context.Context, req *Object) (*Object, error) {
		object := newObject()
		return &object, nil
	}))

	client, err := gorpc.NewClient[Object, Object]("http://127.0.0.1:"+strconv.Itoa(port), gorpc.WithGobClient())
	if err != nil {
		b.Fatal("new client error: " + err.Error())
	}

	go func() {
		if err := server.Start(ctx); err != nil {
			b.Error("server error: " + err.Error())
		}
	}()

	for b.Loop() {
		request := newObject()
		response, err := client.Do(ctx, &request)
		if err != nil {
			b.Fatal("request error: " + err.Error())
		}

		_ = response
	}
}

func newObject() Object {
	return Object{
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
}

func randPort() int {
	const minPort = 49152
	return minPort + (math.MaxUint16 - minPort)
}
