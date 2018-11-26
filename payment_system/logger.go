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

	var reqBody []byte

	if req.Body != nil {
		reqBody, _ = ioutil.ReadAll(req.Body)
	}
	req.Body = ioutil.NopCloser(bytes.NewBuffer(reqBody))

	resp, err := t.transport().RoundTrip(req)
	if err != nil {
		return resp, err
	}

	t.log(req.URL.Path, req.Header, reqBody, resp)

	return resp, err
}

func (t *Transport) transport() http.RoundTripper {
	if t.Transport != nil {
		return t.Transport
	}

	return http.DefaultTransport
}

func (t *Transport) log(reqUrl string, reqHeader http.Header, reqBody []byte, resp *http.Response) {
	var resBody []byte

	if resp.Body != nil {
		resBody, _ = ioutil.ReadAll(resp.Body)
	}
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(resBody))

	data := []interface{}{
		"request_headers", utils.RequestResponseHeadersToString(reqHeader),
		"request_body", string(reqBody),
		"response_headers", utils.RequestResponseHeadersToString(resp.Header),
		"response_body", string(resBody),
	}

	t.Logger.Infow(reqUrl, data...)
}
