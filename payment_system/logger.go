package payment_system

import (
	"bytes"
	"context"
	"github.com/ProtocolONE/p1pay.api/utils"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	defaultHttpClientTimeout = 10
)

type Transport struct {
	Transport http.RoundTripper
	Logger    *zap.SugaredLogger
}

type contextKey struct {
	name string
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := context.WithValue(req.Context(), &contextKey{name: "RequestStart"}, time.Now())
	req = req.WithContext(ctx)

	resp, err := t.transport().RoundTrip(req)
	if err != nil {
		return resp, err
	}

	t.log(req, resp)

	return resp, err
}

func (t *Transport) transport() http.RoundTripper {
	if t.Transport != nil {
		return t.Transport
	}

	return http.DefaultTransport
}

func (t *Transport) log(req *http.Request, resp *http.Response) {
	var reqBody []byte
	var resBody []byte

	if req.Body != nil {
		reqBody, _ = ioutil.ReadAll(req.Body)
	}
	req.Body = ioutil.NopCloser(bytes.NewBuffer(reqBody))

	if resp.Body != nil {
		resBody, _ = ioutil.ReadAll(resp.Body)
	}
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(resBody))

	data := []interface{}{
		"request_headers", utils.RequestResponseHeadersToString(req.Header),
		"request_body", string(reqBody),
		"response_headers", utils.RequestResponseHeadersToString(resp.Header),
		"response_body", string(resBody),
	}

	t.Logger.Infow(req.URL.Path, data...)
}