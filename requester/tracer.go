package requester

import (
	"crypto/tls"
	"net/http"
	"net/http/httptrace"
	"time"
)

// Event is a stage of an outgoing HTTP request associated with a timestamp.
type Event struct {
	Name string        `json:"name"`
	Time time.Duration `json:"time"`
}

// tracer is a http.RoundTripper to be used as a http.Transport
// that records the events of an outgoing HTTP request.
type tracer struct {
	start     time.Time
	events    []Event
	transport http.RoundTripper
}

// RoundTrip implements http.RoundTripper. It attaches the client trace
// to the request context and calls http.DefaultTransport.RoundTrip
// with the new created request.
func (t *tracer) RoundTrip(r *http.Request) (*http.Response, error) {
	ctx := httptrace.WithClientTrace(r.Context(), t.trace())
	return t.transport.RoundTrip(r.WithContext(ctx))
}

// CloseIdleConnections closes any connections on its http.Transport
// which are sitting idle in a "keep-alive" state.
//
// If the tracer's Transport does not have a CloseIdleConnections method
// then this method does nothing.
func (t *tracer) CloseIdleConnections() {
	type closeIdler interface{ CloseIdleConnections() }
	if tr, ok := t.transport.(closeIdler); ok {
		tr.CloseIdleConnections()
	}
}

// trace returns a http.ClientTrace that timestamps and records the events
// of an outgoing HTTP request.
func (t *tracer) trace() *httptrace.ClientTrace {
	return &httptrace.ClientTrace{
		GetConn: func(string) {
			t.start = time.Now()
			t.addEvent("GetConn")
		},
		DNSStart: func(httptrace.DNSStartInfo) {
			t.addEvent("DNSStart")
		},
		DNSDone: func(httptrace.DNSDoneInfo) {
			t.addEvent("DNSDone")
		},
		ConnectStart: func(string, string) {
			t.addEvent("ConnectStart")
		},
		ConnectDone: func(string, string, error) {
			t.addEvent("ConnectDone")
		},
		GotConn: func(httptrace.GotConnInfo) {
			t.addEvent("GotConn")
		},
		TLSHandshakeStart: func() {
			t.addEvent("TLSHandshakeStart")
		},
		TLSHandshakeDone: func(tls.ConnectionState, error) {
			t.addEvent("TLSHandshakeDone")
		},

		WroteHeaders: func() {
			t.addEvent("WroteHeaders")
		},
		WroteRequest: func(httptrace.WroteRequestInfo) {
			t.addEvent("WroteRequest")
		},
		GotFirstResponseByte: func() {
			t.addEvent("GotFirstResponseByte")
		},
		PutIdleConn: func(error) {
			t.addEvent("PutIdleConn")
		},
	}
}

// addEvent timestamps and appends and event to the tracer's events slice.
func (t *tracer) addEvent(name string) {
	t.events = append(t.events, Event{Name: name, Time: time.Since(t.start)})
}

// newTracer returns an initialized tracer.
func newTracer() *tracer {
	p := &tracer{
		events:    make([]Event, 0, 20),
		transport: http.DefaultTransport,
	}
	return p
}
