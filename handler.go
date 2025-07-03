package gorpc

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"hash/maphash"
	"io"
	"net/http"
	"unique"
	"weak"

	"github.com/samborkent/gorpc/goc"
	isync "github.com/samborkent/gorpc/internal/sync"
)

// HandlerFunc is a generic function which takes any request and returns any response.
type HandlerFunc[Request, Response any] func(ctx context.Context, req *Request) (*Response, error)

// Hash return the method hash of the handler func.
// This hash is used to match client requests to server handlers.
func (h HandlerFunc[Request, Response]) Hash() string {
	return hashMethod[Request, Response]()
}

const (
	httpErrInvalidMethod       = "Invalid HTTP method"
	httpErrInvalidContentType  = "Invalid Content-Type header value"
	httpErrInvalidAcceptHeader = "Accept header does not allow goc encoding"
	httpErrMissingMethodHash   = "Missing X-Method-Hash header"
	httpErrInvalidMethodHash   = "Invalid X-Method-Hash header value"
	httpErrRequest             = "Error decoding request"
	httpErrResponse            = "Error encoding or writing response"
)

func handler[Request, Response any](h HandlerFunc[Request, Response], cfg handlerConfig) http.HandlerFunc {
	hsh := h.Hash()
	hshHandle := unique.Make(hsh)

	seed := maphash.MakeSeed()
	var cache isync.Map[uint64, weak.Pointer[Response]]

	return func(w http.ResponseWriter, r *http.Request) {
		if cfg.useGob {
			// Only requests with MIME type application/gob are supported.
			contentType := r.Header.Get(HeaderContentType)
			if contentType != MIMETypeGob {
				http.Error(w, httpErrInvalidContentType, http.StatusUnsupportedMediaType)
				return
			}

			acceptType := r.Header.Get(HeaderAccept)
			if acceptType != MIMETypeGob {
				http.Error(w, httpErrInvalidAcceptHeader, http.StatusNotAcceptable)
				return
			}
		} else {
			// Only requests with MIME type application/goc are supported.
			contentType := r.Header.Get(HeaderContentType)
			if contentType != MIMETypeGoc {
				http.Error(w, httpErrInvalidContentType, http.StatusUnsupportedMediaType)
				return
			}

			acceptType := r.Header.Get(HeaderAccept)
			if acceptType != MIMETypeGoc {
				http.Error(w, httpErrInvalidAcceptHeader, http.StatusNotAcceptable)
				return
			}
		}

		header := r.Header.Get(HeaderMethodHash)
		if header == "" {
			http.Error(w, httpErrMissingMethodHash, http.StatusBadRequest)
			return
		}

		// Check that the request belongs to this handler.
		if unique.Make(header) != hshHandle {
			http.Error(w, httpErrInvalidMethodHash, http.StatusForbidden)
			return
		}

		// TODO; reject requests which have content length not set

		var (
			req         Request
			err         error
			payloadHash uint64
		)

		// Decode request.
		if cfg.cacheResponse {
			// TODO: read until content length
			body, err := io.ReadAll(r.Body)
			_ = r.Body.Close()
			if err != nil {
				// TODO: revise error
				http.Error(w, httpErrRequest, http.StatusBadRequest)
				return
			}

			payloadHash = maphash.Bytes(seed, body)

			res, ok := cache.Load(payloadHash)
			if ok && res.Value() != nil {
				// TODO: move to separate handler
				if cfg.useGob {
					err = gob.NewDecoder(bytes.NewReader(body)).Decode(&req)
				} else {
					err = goc.Decode(body, &req)
				}

				if err != nil {
					http.Error(w, httpErrRequest, http.StatusBadRequest)
					return
				}
			}
		} else {
			if cfg.useGob {
				err = gob.NewDecoder(r.Body).Decode(&req)
			} else {
				err = goc.DecodeRead(r.Body, &req)
			}

			_ = r.Body.Close()

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

		if cfg.useGob {
			w.Header().Set(HeaderContentType, MIMETypeGob)
		} else {
			w.Header().Set(HeaderContentType, MIMETypeGoc)
		}

		w.Header().Set(HeaderXContentTypeOptions, nosniff)
		w.Header().Set(HeaderMethodHash, hsh)

		// Encode and return response.
		if cfg.cacheResponse && payloadHash > 0 {
			if cfg.useGob {
				if err := gob.NewEncoder(w).Encode(res); err != nil {
					http.Error(w, httpErrResponse, http.StatusInternalServerError)
					return
				}
			} else {
				payload, err := goc.Encode(res)
				if err != nil {
					http.Error(w, httpErrResponse, http.StatusInternalServerError)
					return
				}

				_, err = w.Write(payload)
				if err != nil {
					// TODO: send to optional logger
				}
			}

			// TODO: does it even make sense to use weak pointer cache for server?
			cache.Store(payloadHash, weak.Make(res))
		} else {
			// TODO: define constants
			w.Header().Set("Cache-Control", "no-store")

			if cfg.useGob {
				err = gob.NewEncoder(w).Encode(res)
			} else {
				_, err = goc.EncodeWrite(w, res)
			}

			if err != nil {
				http.Error(w, httpErrResponse, http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}

type handlerConfig struct {
	cacheResponse bool
	useGob        bool
}
