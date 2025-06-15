package gorpc

import "net/http"

var (
	httpProtocols = new(http.Protocols)
	// TODO: check if this is save, we need a copy of default transport, not a reference.
	httpDefaultTransport *http.Transport
	httpRoundTripper     http.RoundTripper
)

func init() {
	httpProtocols.SetUnencryptedHTTP2(true)

	defaultTransport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		panic("cannot cast http.DefaultTransport to *http.Transport")
	}

	httpDefaultTransport = defaultTransport.Clone()
	httpDefaultTransport.ForceAttemptHTTP2 = true
	httpDefaultTransport.Protocols = httpProtocols

	httpRoundTripper = http.RoundTripper(httpDefaultTransport)
}

type httpRoundTripperFunc func(*http.Request) (*http.Response, error)

func (r httpRoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	if r == nil {
		return httpRoundTripper.RoundTrip(req)
	}

	return r(req)
}
