package fatzebra

import (
	"net/http"
)

// Transport represents the transport that injects credentials.
type Transport struct {
	c    *Client
	base http.RoundTripper
}

// RoundTrip implements round trips as required by http.RoundTrippr
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Host != t.c.host {
		return t.base.RoundTrip(req)
	}

	r := cloneRequest(req)
	r.SetBasicAuth(t.c.username, t.c.password)

	if t.base == nil {
		return http.DefaultTransport.RoundTrip(r)
	}

	return t.base.RoundTrip(r)
}

// Taken from https://github.com/golang/oauth2
func cloneRequest(r *http.Request) *http.Request {
	// shallow copy of the struct
	r2 := new(http.Request)
	*r2 = *r
	// deep copy of the Header
	r2.Header = make(http.Header, len(r.Header))
	for k, s := range r.Header {
		r2.Header[k] = append([]string(nil), s...)
	}
	return r2
}
