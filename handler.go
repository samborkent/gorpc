package gorpc

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"unique"

	"github.com/samborkent/gorpc/goc"
)

type HandlerFunc[Request, Response any] func(ctx context.Context, req *Request) (*Response, error)

func (h HandlerFunc[Request, Response]) Hash() string {
	return hashMethod[Request, Response]()
}

func handler[Request, Response any](h HandlerFunc[Request, Response]) http.HandlerFunc {
	hsh := unique.Make(h.Hash())

	return func(w http.ResponseWriter, r *http.Request) {
		defer func() { _ = r.Body.Close() }()

		mimeType := r.Header.Get(HeaderContentType)
		if mimeType != MIMEType {
			slog.ErrorContext(r.Context(), "invalid Content-Type MIME header", slog.String("mime", mimeType))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		header := r.Header.Get(HeaderMethodHash)
		if header == "" {
			slog.ErrorContext(r.Context(), "missing method hash header")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if unique.Make(header) != hsh {
			slog.ErrorContext(r.Context(), "method hash mismatch", slog.String("header", header), slog.String("hash", h.Hash()))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		req, err := goc.DecodeFrom[Request](r.Body)
		if err != nil {
			slog.ErrorContext(r.Context(), "request decoding error: "+err.Error())
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		res, err := h(r.Context(), &req)
		if err != nil {
			slog.ErrorContext(r.Context(), "server error: "+err.Error())

			var e *Error
			if errors.As(err, &e) {
				http.Error(w, e.Text, e.Code)
				return
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := goc.EncodeTo(w, res); err != nil {
			slog.ErrorContext(r.Context(), "response encoding error: "+err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}
