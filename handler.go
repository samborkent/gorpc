package gorpc

import (
	"context"
	"errors"
	"hash/maphash"
	"net/http"
	"sync"
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

func handler[Request, Response any](h HandlerFunc[Request, Response], cacheResponse bool) http.HandlerFunc {
	hsh := h.Hash()
	hshHandle := unique.Make(hsh)

	var seed maphash.Seed
	// TODO: use sync.Map?
	var cache map[uint64]weak.Pointer[Response]
	var cacheLock *sync.RWMutex

	if cacheResponse {
		seed = maphash.MakeSeed()
		cache = make(map[uint64]weak.Pointer[Response])
		cacheLock = new(sync.RWMutex)
	}

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

		// TODO; reject requests which have content length not set

		var req Request
		var payloadHash uint64

		// Decode request.
		if cacheResponse {
			// TODO: read until content length
			body, err := io.ReadAll(req.Body)
			if err != nil {
				// TODO: revise error
				http.Error(w, httpErrRequest, http.StatusBadRequest)
				return
			}

			payloadHash = maphash.Bytes(seed, body)

			cacheLock.RLock()
			res, ok := cache[payloadHash]
			cacheLock.RUnlock()

			if ok && res.Value != nil {
				//TODO: resolve race condition
				return res.Value, nil
			}

			req, err = goc.Decode[Request](body)
			if err != nil {
				http.Error(w, httpErrRequest, http.StatusBadRequest)
				return
			}
		} else {
			req, err = goc.DecodeFrom[Request](r.Body)
			if err != nil {
				http.Error(w, httpErrRequest, http.StatusBadRequest)
				return
			}
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
		if cacheResponse && payloadHash > 0 {
			res, err := goc.Encode(res)
			if err != nil {
				http.Error(w, httpErrResponse, http.StatusInternalServerError)
				return
			}

			cacheLock.Lock()
			// TODO: does it even make sense to use weak pointer cache for server?
			cache[payloadHash] = weak.Make(&res)
			cacheLock.Unlock()
		} else {
			if err := goc.EncodeTo(w, res); err != nil {
				http.Error(w, httpErrResponse, http.StatusInternalServerError)
				return
			}
		}
	}
}
