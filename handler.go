package gorpc

import (
	"context"
	"errors"
	"net/http"
	"unique"

	"github.com/samborkent/gorpc/goc"
)

// HandlerFunc is a generic function which takes any request and returns any response.
type HandlerFunc[Request, Response any] func(ctx context.Context, req *Request) (*Response, error)

// Hash return the method hash of the handler func.
// This hash is used to match client requests to server handlers.
func (h HandlerFunc[Request, Response]) Hash() string {
	return hashMethod[Request, Response]()
}

const (
	httpErrInvalidMethod      = "Invalid HTTP method"
	httpErrInvalidContentType = "Invalid Content-Type header value"
	httpErrMissingMethodHash  = "Missing X-Method-Hash header"
	httpErrInvalidMethodHash  = "Invalid X-Method-Hash header value"
	httpErrRequest            = "Error decoding request"
	httpErrResponse           = "Error encoding or writing response"
)

func handler[Request, Response any](h HandlerFunc[Request, Response]) http.HandlerFunc {
	hsh := h.Hash()
	hshHandle := unique.Make(hsh)

	return func(w http.ResponseWriter, r *http.Request) {
		defer func() { _ = r.Body.Close() }()

		// Only POST requests are supported.
		if r.Method != http.MethodPost {
			http.Error(w, httpErrInvalidMethod, http.StatusBadRequest)
			return
		}

		// Only requests with MIME type application/goc are supported.
		mimeType := r.Header.Get(HeaderContentType)
		if mimeType != MIMEType {
			http.Error(w, httpErrInvalidContentType, http.StatusBadRequest)
			return
		}

		header := r.Header.Get(HeaderMethodHash)
		if header == "" {
			http.Error(w, httpErrMissingMethodHash, http.StatusBadRequest)
			return
		}

		// Check that the request belongs to this handler.
		if unique.Make(header) != hshHandle {
			http.Error(w, httpErrInvalidMethodHash, http.StatusBadRequest)
			return
		}

		// Decode request.
		req, err := goc.DecodeFrom[Request](r.Body)
		if err != nil {
			http.Error(w, httpErrRequest, http.StatusBadRequest)
			return
		}

		// Call handler func.
		res, err := h(r.Context(), &req)
		if err != nil {
			// If handler func returns an [Error], return it as HTTP error.
			var e *Error
			if errors.As(err, &e) {
				http.Error(w, e.Text, e.Code)
				return
			}

			// Otherwise, return entire error as 500 Internal Server Error.
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set(HeaderContentType, MIMEType)
		w.Header().Set(HeaderMethodHash, hsh)

		// Encode and return response.
		if err := goc.EncodeTo(w, res); err != nil {
			http.Error(w, httpErrResponse, http.StatusInternalServerError)
			return
		}
	}
}
