package payment_system

import (
	"bytes"
	"context"
	"github.com/ProtocolONE/p1pay.api/database/dao"
	"github.com/ProtocolONE/p1pay.api/database/model"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"time"
)

const (
	defaultHttpClientTimeout = 10
)

type Transport struct {
	Transport http.RoundTripper
	Database  dao.Database
}

type contextKey struct {
	name string
}

func GetLoggableHttpClient(db dao.Database) *http.Client {
	return &http.Client{
		Transport: &Transport{Database: db},
		Timeout:   time.Duration(defaultHttpClientTimeout * time.Second),
	}
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := context.WithValue(req.Context(), &contextKey{name: "RequestStart"}, time.Now())
	req = req.WithContext(ctx)

	//t.logRequest(req)

	var headerToString = func(headers map[string][]string) string {
		var out string

		for k, v := range headers {
			out += k + ":" + v[0] + "\n "
		}

		return out
	}

	var reqBody []byte

	if req.Body != nil {
		reqBody, _ = ioutil.ReadAll(req.Body)
	}
	req.Body = ioutil.NopCloser(bytes.NewBuffer(reqBody))

	rLog := &model.Log{
		RequestHeaders: headerToString(req.Header),
		RequestBody:    string(reqBody),
	}

	resp, err := t.transport().RoundTrip(req)
	if err != nil {
		return resp, err
	}

	//t.logResponse(resp)

	var resBody []byte

	rLog.ResponseHeaders = headerToString(resp.Header)

	if resBody, err = ioutil.ReadAll(resp.Body); err == nil {
		rLog.ResponseBody = string(resBody)
	}
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(resBody))

	t.Database.Repository("log").InsertLog(rLog)

	return resp, err
}

func (t *Transport) logRequest(req *http.Request) {
	dump, err := httputil.DumpRequestOut(req, true)

	if err != nil {
		return
	}

	log.SetOutput(os.Stdout)
	log.Println(string(dump))
}

func (t *Transport) logResponse(resp *http.Response) {
	dump, err := httputil.DumpResponse(resp, true)

	if err != nil {
		return
	}

	log.SetOutput(os.Stdout)
	log.Println(string(dump))
}

func (t *Transport) transport() http.RoundTripper {
	if t.Transport != nil {
		return t.Transport
	}

	return http.DefaultTransport
}
