package payment_system

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"time"
)

type Transport struct {
	Transport http.RoundTripper
}

type contextKey struct {
	name string
}

func GetLoggableHttpClient() *http.Client {
	return &http.Client{Transport: &Transport{}}
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := context.WithValue(req.Context(), &contextKey{name: "RequestStart"}, time.Now())
	req = req.WithContext(ctx)

	t.logRequest(req)

	resp, err := t.transport().RoundTrip(req)
	if err != nil {
		return resp, err
	}

	t.logResponse(resp)

	return resp, err
}

func (t *Transport) logRequest(req *http.Request) {
	dump, err := httputil.DumpRequestOut(req, true)

	if err != nil {
		return
	}

	fmt.Println(string(dump))
}

func (t *Transport) logResponse(resp *http.Response) {
	dump, err := httputil.DumpResponse(resp, true)

	if err != nil {
		return
	}

	fmt.Println(string(dump))
}

func (t *Transport) transport() http.RoundTripper {
	if t.Transport != nil {
		return t.Transport
	}

	return http.DefaultTransport
}
