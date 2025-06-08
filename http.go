package gorpc

var (
    httpProtocols = new(http.Protocols)
    // TODO: check if this is save, we need a copy of default transport, not a reference.
	httpDefaultTransport = http.DefaultTransport
)

func init() {
	httpProtocols.SetUnencryptedHTTP2(true)
	httpDefaultTransport.ForceAttemptHTTP2 = true
	httpDefaultTransport.Protocols = httpProtocols
}

type httpRoundTripperFunc func(*http.Request) (*http.Response, error)

func (r httpRoundTripperFunc) RoundTrip(req *http.Request) (*http.Respon
     if r == nil {
          return httpDefaultTransport.RoundTrip(req)
     }

     return r(req)
}
