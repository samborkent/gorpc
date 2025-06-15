package gorpc_test

import (
	"context"
	cryptorand "crypto/rand"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/samborkent/gorpc"
)

func TestServerClient(t *testing.T) {
	t.Parallel()

	t.Log(gorpc.HandlerFunc[request, response](testHandler).Hash())

	server, err := gorpc.NewServer(-1)
	if err != nil {
		t.Fatal("got server error: " + err.Error())
	}

	gorpc.Register(server, testHandler)

	go func() {
		if err := server.Start(t.Context()); err != nil {
			t.Errorf("server error: %s", err.Error())
		}
	}()

	time.Sleep(100 * time.Millisecond)

	client, err := gorpc.NewClient[request, response]("http://127.0.0.1:" + strconv.Itoa(server.Port()))
	if err != nil {
		t.Fatal("got client error: " + err.Error())
	}

	t.Run("error", func(t *testing.T) {
		t.Parallel()

		resp, err := client.Do(t.Context(), &request{})
		if err == nil {
			t.Fatal("expected error")
		}

		if !strings.Contains(err.Error(), "404 Not Found") {
			t.Error("expected not found error")
		}

		if resp != nil {
			t.Error("response should be nil")
		}
	})
	t.Run("success", func(t *testing.T) {
		t.Parallel()

		resp, err := client.Do(t.Context(), &request{
			ID:       successResponse.ID,
			Password: "password",
		})
		if err != nil {
			t.Fatal("client error: " + err.Error())
		}

		if resp == nil {
			t.Fatal("response should not be nil")
		}

		if *resp != successResponse {
			t.Errorf("wrong response: got %+v, want %+v", resp, successResponse)
		}
	})
}

type request struct {
	ID       uint64
	Password string
}

type response struct {
	ID    uint64
	Name  string
	Email string
}

var successResponse = response{
	ID:    rand.Uint64(),
	Name:  cryptorand.Text(),
	Email: cryptorand.Text(),
}

func testHandler(ctx context.Context, req *request) (*response, error) {
	switch req.ID {
	case successResponse.ID:
		slog.InfoContext(ctx, "got request!!", slog.String("pass", req.Password))
		return &successResponse, nil
	default:
		return nil, &gorpc.Error{
			Code: http.StatusUnavailableForLegalReasons,
			Text: "FOOBAR",
		}
	}
}
